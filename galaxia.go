package galaxia

import (
	"context"
	"errors"
	"github.com/atsegelnyk/galaxia/auth"
	"github.com/atsegelnyk/galaxia/metrics"
	"time"

	"github.com/atsegelnyk/galaxia/entityregistry"
	"github.com/atsegelnyk/galaxia/model"
	"github.com/atsegelnyk/galaxia/session"
	"github.com/atsegelnyk/galaxia/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

const StartCMDName = "start"

type ProcessorOption func(*Processor)

type Processor struct {
	api *tgbotapi.BotAPI

	sessionRepository session.Repository
	entityRegistry    *entityregistry.Registry
	exporter          *metrics.PrometheusExporter
	auther            auth.Auther
}

func NewProcessor(opts ...ProcessorOption) *Processor {
	g := &Processor{
		entityRegistry: entityregistry.New(),
		auther:         auth.NewFakeAlwaysAuther(),
		exporter:       metrics.NewPrometheusExporter(),
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func WithApi(api *tgbotapi.BotAPI) ProcessorOption {
	return func(g *Processor) {
		g.api = api
	}
}

func WithBotToken(token string) ProcessorOption {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	return func(g *Processor) {
		g.api = api
	}
}

func WithEntityRegistry(er *entityregistry.Registry) ProcessorOption {
	return func(g *Processor) {
		g.entityRegistry = er
	}
}

func WithSessionRepository(r session.Repository) ProcessorOption {
	return func(g *Processor) {
		g.sessionRepository = r
	}
}

func WithAuther(a auth.Auther) ProcessorOption {
	return func(g *Processor) {
		g.auther = a
	}
}

func WithMetricAddr(addr string) ProcessorOption {
	return func(g *Processor) {
		g.exporter.Listen = addr
	}
}

func (p *Processor) Start(ctx context.Context) {
	err := p.preflightCheck()
	if err != nil {
		log.Fatal(err)
	}

	go p.exporter.Serve(ctx)
	log.Println("start pooling")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := p.api.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			go func() {
				err = p.handleUpdate(&update)
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}

}

func (p *Processor) AsyncUpdate(update *model.UserUpdate) error {
	ses, err := p.sessionRepository.Get(update.UserID)
	if err != nil {
		if !errors.Is(err, session.NotFoundError) {
			return err
		}
		ses = session.NewSession(update.UserID)
	}
	return p.processUserUpdate(ses, update)
}

func (p *Processor) preflightCheck() error {
	_, err := p.entityRegistry.GetCommand(0, StartCMDName)
	if err != nil {
		return err
	}
	if p.entityRegistry == nil {
		return errors.New("entity registry is nil")
	}
	if p.sessionRepository == nil {
		return errors.New("session repository is nil")
	}
	return nil
}

// event processor

func (p *Processor) initSession(update *tgbotapi.Update) *session.Session {
	return session.NewSession(
		update.Message.Chat.ID,
		session.WithUsername(update.Message.From.UserName),
		session.WithLang(update.Message.From.LanguageCode),
		session.WithName(update.Message.From.FirstName),
		session.WithLastName(update.Message.From.LastName),
	)
}

func (p *Processor) enrichSession(ses *session.Session, update *tgbotapi.Update) {
	ses.UserContext.Username = update.Message.From.UserName
	ses.UserContext.Lang = update.Message.From.LanguageCode
	ses.UserContext.Name = update.Message.From.FirstName
	ses.UserContext.LastName = update.Message.From.LastName
}

func (p *Processor) handleUpdate(update *tgbotapi.Update) error {
	if update.Message != nil {
		err := p.auther.AuthN(update.Message.Chat.ID)
		if err != nil {
			if errors.Is(err, auth.UnauthorizedErr) {
				p.exporter.Increase(metrics.UnauthenticatedRequestsCountMetric)
				return nil
			}
			return err
		}

		p.exporter.Increase(metrics.UserMessagesSentCountMetric)
		ses, err := p.sessionRepository.Get(update.Message.Chat.ID)
		if err != nil {
			if !errors.Is(err, session.NotFoundError) {
				return err
			}
			ses = p.initSession(update)
		} else {
			p.enrichSession(ses, update)
		}

		ses.AppendStageMessages(update.Message.MessageID)
		if update.Message.Command() != "" {
			return p.handleCMD(ses, update)
		}
		return p.handleMessage(ses, update)
	}

	if update.CallbackQuery != nil {
		ses, err := p.sessionRepository.Get(int64(update.CallbackQuery.From.ID))
		if err != nil {
			return err
		}
		return p.handleCallbackQuery(ses, update)
	}
	return nil
}

// event processors by type

func (p *Processor) handleCMD(session *session.Session, update *tgbotapi.Update) error {
	cmd, err := p.entityRegistry.GetCommand(update.Message.Chat.ID, model.ResourceRef(update.Message.Command()))
	if err != nil {
		return err
	}
	action, err := p.entityRegistry.GetAction(
		update.Message.Chat.ID,
		cmd.ActionRef(),
	)
	if err != nil {
		return err
	}

	p.exporter.IncreaseWithLabels(metrics.CmdExecutedCountMetric, map[string]string{
		metrics.CmdRefLabel: string(cmd.SelfRef()),
	})
	start := time.Now()
	userUpdate := action.Func()(session.UserContext, update)
	p.exporter.ObserveWithLabels(metrics.RequestDurationBucketMetric, time.Since(start), map[string]string{
		metrics.ActionRefLabel: string(action.SelfRef()),
	})
	return p.processUserUpdate(session, userUpdate)
}

func (p *Processor) handleMessage(session *session.Session, update *tgbotapi.Update) error {
	stageRef := session.GetCurrentStage()
	if stageRef.Empty() {
		startCmd, _ := p.entityRegistry.GetCommand(update.Message.Chat.ID, StartCMDName)
		action, err := p.entityRegistry.GetAction(
			update.Message.Chat.ID,
			startCmd.ActionRef(),
		)
		if err != nil {
			return err
		}
		p.exporter.IncreaseWithLabels(metrics.CmdExecutedCountMetric, map[string]string{
			metrics.CmdRefLabel: string(startCmd.SelfRef()),
		})
		start := time.Now()
		userUpdate := action.Func()(session.UserContext, update)
		p.exporter.ObserveWithLabels(metrics.RequestDurationBucketMetric, time.Since(start), map[string]string{
			metrics.ActionRefLabel: string(action.SelfRef()),
		})
		return p.processUserUpdate(session, userUpdate)
	}

	stg, err := p.entityRegistry.GetStage(update.Message.Chat.ID, stageRef)
	if err != nil {
		return err
	}

	userUpdate, err := p.handleStage(session, stg, update)
	if err != nil {
		return err
	}

	return p.processUserUpdate(session, userUpdate)
}

func (p *Processor) handleStage(ses *session.Session, stg *model.Stage, update *tgbotapi.Update) (*model.UserUpdate, error) {
	if actionRef, ok := ses.PendingInputs[update.Message.Text]; ok {
		action, err := p.entityRegistry.GetAction(update.Message.Chat.ID, actionRef)
		if err != nil {
			return nil, err
		}
		p.exporter.IncreaseWithLabels(metrics.StageActionProcessedCountMetric, map[string]string{
			metrics.StageRefLabel:  string(stg.SelfRef()),
			metrics.ActionRefLabel: string(actionRef),
		})
		start := time.Now()
		userUpdate := action.Func()(ses.UserContext, update)
		p.exporter.ObserveWithLabels(metrics.RequestDurationBucketMetric, time.Since(start), map[string]string{
			metrics.ActionRefLabel: string(action.SelfRef()),
		})
		return userUpdate, nil
	}

	if stg.CustomInputAllowed() {
		defActionRef := stg.DefaultActionRef()
		defAction, err := p.entityRegistry.GetAction(update.Message.Chat.ID, defActionRef)
		if err != nil {
			return nil, err
		}
		p.exporter.IncreaseWithLabels(metrics.StageActionProcessedCountMetric, map[string]string{
			metrics.StageRefLabel:  string(stg.SelfRef()),
			metrics.ActionRefLabel: string(defActionRef),
		})
		start := time.Now()
		userUpdate := defAction.Func()(ses.UserContext, update)
		p.exporter.ObserveWithLabels(metrics.RequestDurationBucketMetric, time.Since(start), map[string]string{
			metrics.ActionRefLabel: string(defActionRef),
		})
		return userUpdate, nil
	}
	return nil, model.UnrecognizedInputError
}

func (p *Processor) handleCallbackQuery(ses *session.Session, update *tgbotapi.Update) error {
	pendingCallbak, err := ses.GetPendingCallback(update.CallbackQuery.Data)
	if err != nil {
		return err
	}

	callbackHandler, err := p.entityRegistry.GetCallbackHandler(int64(update.CallbackQuery.From.ID), pendingCallbak.HandlerRef)
	if err != nil {
		return err
	}
	action, err := p.entityRegistry.GetAction(
		int64(update.CallbackQuery.From.ID),
		callbackHandler.ActionRef(),
	)
	if err != nil {
		return err
	}

	ses.UserContext.CallbackData = &pendingCallbak.UserData
	p.exporter.IncreaseWithLabels(metrics.CallbacksProcessedCountMetric, map[string]string{
		metrics.CallbackHandlerRefLabel: string(callbackHandler.SelfRef()),
	})
	start := time.Now()
	userUpdate := action.Func()(ses.UserContext, update)
	p.exporter.ObserveWithLabels(metrics.RequestDurationBucketMetric, time.Since(start), map[string]string{
		metrics.ActionRefLabel: string(action.SelfRef()),
	})
	return p.processUserUpdate(ses, userUpdate)
}

// user response handlerx

func (p *Processor) processUserUpdate(ses *session.Session, update *model.UserUpdate) error {
	stageReInit := false
	if update.Messages != nil {
		p.callbackMapper(ses, update)
		stageReInit = true
	}

	err := p.processTransit(stageReInit, ses, update)
	if err != nil {
		return err
	}
	sentMessages, err := p.respond(update)
	ses.AppendStageMessages(sentMessages...)
	if err != nil {
		return err
	}
	ses.UserContext.CallbackData = nil
	return p.sessionRepository.Save(ses)
}

// callbackID mapper

func (p *Processor) callbackMapper(ses *session.Session, update *model.UserUpdate) {
	for _, message := range update.Messages {
		if message.InlineKeyboard != nil {
			for _, row := range message.InlineKeyboard {
				for _, button := range row {
					callbackID := utils.GenerateCallbackID()
					button.Data = callbackID
					ses.RegisterCallback(callbackID,
						button.CallbackBehaviour,
						button.UserData,
						button.CallbackHandlerRef,
					)
				}
			}
		}
	}
}

func (p *Processor) processTransit(stageReInit bool, ses *session.Session, update *model.UserUpdate) error {
	currentStageRef := ses.GetCurrentStage()

	if update.Transit != nil {
		next, err := p.entityRegistry.GetStage(update.UserID, update.Transit.TargetRef)
		if err != nil {
			return err
		}

		ses.SetNextStage(next.SelfRef())
		initialMessages, err := p.initStage(ses, next)
		if err != nil {
			return err
		}
		update.Messages = append(update.Messages, initialMessages...)
		if update.Transit.Clean {
			update.ToDeleteMessages = append(update.ToDeleteMessages, ses.StageMessages...)
			ses.Clean()
		}

		p.exporter.IncreaseWithLabels(metrics.StageReachedCountMetric, map[string]string{
			metrics.StageRefLabel: string(next.SelfRef()),
		})
	} else if !currentStageRef.Empty() && stageReInit {
		currentStage, err := p.entityRegistry.GetStage(update.UserID, currentStageRef)
		if err != nil {
			return err
		}
		initialMessages, err := currentStage.Initializer().Init(update.UserID, currentStageRef)
		if err != nil {
			return err
		}
		update.Messages = append(update.Messages, initialMessages...)
	}
	return nil
}

func (p *Processor) initStage(ses *session.Session, stg *model.Stage) ([]*model.Message, error) {
	ses.PendingInputs = make(map[string]model.ResourceRef)
	initMessages, err := stg.Initialize(ses.UserID, stg.SelfRef())
	if err != nil {
		return nil, err
	}

	last := initMessages[len(initMessages)-1]
	if initMessages[len(initMessages)-1] != nil {
		for _, row := range last.ReplyKeyboard {
			for _, button := range row {
				ses.PendingInputs[button.Text] = button.ActionRef
			}
		}
	}

	return initMessages, nil
}

func (p *Processor) respond(update *model.UserUpdate) ([]int, error) {
	var sentMessages []int

	if update.CallbackQueryResponse != nil {
		callBackConfig := utils.TransformCallbackQueryResponse(update.CallbackQueryResponse)
		_, err := p.api.AnswerCallbackQuery(callBackConfig)
		if err != nil {
			log.Println(err)
		}
	}

	for _, msg := range update.Messages {
		var chattables []tgbotapi.Chattable

		if msg.Photo != nil {
			photoConfig := utils.TransformPhoto(update.UserID, msg.Photo)
			chattables = append(chattables, photoConfig)
		}
		if msg.Video != nil {
			videoConfig := utils.TransformVideo(update.UserID, msg.Video)
			chattables = append(chattables, videoConfig)
		}
		mcgConfig := utils.TransformMessage(update.UserID, msg)
		chattables = append(chattables, mcgConfig)

		for _, c := range chattables {
			sent, err := p.api.Send(c)
			if err != nil {
				return sentMessages, err
			}
			p.exporter.Increase(metrics.BotMessagesSentCountMetric)
			sentMessages = append(sentMessages, sent.MessageID)
		}
	}

	for _, ID := range update.ToDeleteMessages {
		_, err := p.api.DeleteMessage(tgbotapi.NewDeleteMessage(update.UserID, ID))
		if err != nil {
			return sentMessages, err
		}
	}
	return sentMessages, nil
}
