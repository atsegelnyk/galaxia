package galaxia

import (
	"context"
	"errors"
	"fmt"
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
}

func NewProcessor(opts ...ProcessorOption) *Processor {
	g := &Processor{
		entityRegistry: entityregistry.New(),
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

func (p *Processor) RegisterCMD(cmd *model.Command) error {
	return p.entityRegistry.RegisterCommand(cmd)
}

func (p *Processor) RegisterStage(stage *model.Stage) error {
	return p.RegisterStage(stage)
}

func (p *Processor) Start(ctx context.Context) {
	err := p.preflightCheck()
	if err != nil {
		log.Fatal(err)
	}
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
				err = p.processUpdate(&update)
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}

}

func (p *Processor) HandleUserUpdate(updater model.Updater) error {
	ses, err := p.sessionRepository.Get(updater.GetUserID())
	if err != nil {
		if !errors.Is(err, session.NotFoundError) {
			return err
		}
		ses = session.NewSession(updater.GetUserID())
	}
	return p.handleUserUpdate(ses, updater)
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

func (p *Processor) processUpdate(update *tgbotapi.Update) error {
	if update.Message != nil {
		if auther := p.entityRegistry.GetAuther(); auther != nil {
			err := auther.Authorize(update.Message.Chat.ID)
			if err != nil {
				return err
			}
		}

		ses, err := p.sessionRepository.Get(update.Message.Chat.ID)
		if err != nil {
			if !errors.Is(err, session.NotFoundError) {
				return err
			}
			ses = p.initSession(update)
		}
		ses.AppendStageMessage(int64(update.Message.MessageID))
		if update.Message.Command() != "" {
			return p.processCmd(ses, update)
		}

		return p.processMessage(ses, update)
	}

	if update.CallbackQuery != nil {
		ses, err := p.sessionRepository.Get(int64(update.CallbackQuery.From.ID))
		if err != nil {
			return err
		}
		return p.processCallbackQuery(ses, update)
	}
	return nil
}

// event processors by type

func (p *Processor) processCmd(session *session.Session, update *tgbotapi.Update) error {
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

	updater := action.Func()(session.UserContext, update)
	return p.handleUserUpdate(session, updater)
}

func (p *Processor) processMessage(session *session.Session, update *tgbotapi.Update) error {
	stageRef := session.GetCurrentStage()
	if stageRef.Empty() {
		cmd, _ := p.entityRegistry.GetCommand(update.Message.Chat.ID, StartCMDName)
		action, err := p.entityRegistry.GetAction(
			update.Message.Chat.ID,
			cmd.ActionRef(),
		)
		if err != nil {
			return err
		}
		response := action.Func()(session.UserContext, update)
		return p.handleUserUpdate(session, response)
	}

	stg, err := p.entityRegistry.GetStage(update.Message.Chat.ID, stageRef)
	if err != nil {
		return err
	}

	updater, err := p.processStage(session, stg, update)
	if err != nil {
		return err
	}

	return p.handleUserUpdate(session, updater)
}

func (p *Processor) processCallbackQuery(session *session.Session, update *tgbotapi.Update) error {
	callbackHandlerRef, err := session.GetPendingCallbackHandler(update.CallbackQuery.Data)
	if err != nil {
		return err
	}

	callbackHandler, err := p.entityRegistry.GetCallbackHandler(int64(update.CallbackQuery.From.ID), callbackHandlerRef)
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

	updater := action.Func()(session.UserContext, update)
	return p.handleUserUpdate(session, updater)
}

// user response handler

func (p *Processor) handleUserUpdate(ses *session.Session, updater model.Updater) error {
	if updater == nil {
		return nil
	}
	messagesSent := false

	if updater.GetCallbackResponse() != nil {
		err := p.respondCallbackQueryResponse(updater)
		if err != nil {
			log.Println(err)
		}
	}

	if updater.GetMessages() != nil {
		p.callbackMapper(ses, updater)
		err := p.respondMessages(ses, updater)
		messagesSent = true
		if err != nil {
			log.Println(err)
		}
	}

	return p.respondTransit(messagesSent, ses, updater)
}

// callbackID mapper

func (p *Processor) initStage(ses *session.Session, stg *model.Stage) (*model.Message, error) {
	initMessage, err := stg.Initialize(ses.UserID)
	if err != nil {
		return nil, err
	}
	if initMessage != nil {
		for _, row := range initMessage.ReplyKeyboard {
			for _, button := range row {
				ses.PendingInputs[button.Text] = button.ActionRef
			}
		}
	}
	return initMessage, nil
}

func (p *Processor) processStage(ses *session.Session, stg *model.Stage, update *tgbotapi.Update) (model.Updater, error) {
	if actionRef, ok := ses.PendingInputs[update.Message.Text]; ok {
		action, err := p.entityRegistry.GetAction(update.Message.Chat.ID, actionRef)
		if err != nil {
			return nil, err
		}
		updater := action.Func()(ses.UserContext, update)
		return updater, nil
	}

	if stg.CustomInputAllowed() {
		defActionRef := stg.DefaultActionRef()
		defAction, err := p.entityRegistry.GetAction(update.Message.Chat.ID, defActionRef)
		if err != nil {
			return nil, err
		}
		updater := defAction.Func()(ses.UserContext, update)
		return updater, nil
	}
	return nil, model.UnrecognizedInputError
}

// callbackID mapper

func (p *Processor) callbackMapper(ses *session.Session, updater model.Updater) {
	messages := updater.GetMessages()
	for _, message := range messages {
		if message.InlineKeyboard != nil {
			for _, row := range message.InlineKeyboard {
				for _, button := range row {
					callbackID := utils.GenerateCallbackID()
					button.Data = callbackID
					ses.RegisterCallback(callbackID, button.CallbackBehaviour, button.CallbackHandlerRef)
				}
			}
		}
	}
}

// responders by type

func (p *Processor) respondCallbackQueryResponse(updater model.Updater) error {
	if updater.GetCallbackResponse() != nil {
		callBackConfig := utils.TransformCallbackQueryResponse(updater.GetCallbackResponse())
		_, err := p.api.AnswerCallbackQuery(callBackConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Processor) respondMessages(ses *session.Session, updater model.Updater) error {
	var chattables []tgbotapi.Chattable

	userID := updater.GetUserID()
	messages := updater.GetMessages()

	for _, msg := range messages {
		if msg.Photo != nil {
			photoConfig := utils.TransformPhoto(userID, msg.Photo)
			chattables = append(chattables, photoConfig)
		}
		if msg.Video != nil {
			videoConfig := utils.TransformVideo(userID, msg.Video)
			chattables = append(chattables, videoConfig)
		}
		mcgConfig := utils.TransformMessage(userID, msg)
		chattables = append(chattables, mcgConfig)
	}
	for _, c := range chattables {
		sentMsg, err := p.api.Send(c)
		if err != nil {
			return err
		}
		ses.AppendStageMessage(int64(sentMsg.MessageID))
	}
	return nil
}

func (p *Processor) respondTransit(messagesSent bool, ses *session.Session, updater model.Updater) error {
	userID := updater.GetUserID()
	currentStageName := ses.GetCurrentStage()
	transitConfig := updater.GetTransitConfig()

	var chattables []tgbotapi.Chattable
	var deletees []int64

	if transitConfig != nil {
		next, err := p.entityRegistry.GetStage(userID, transitConfig.TargetRef)
		if err != nil {
			return err
		}
		ses.SetNextStage(next.SelfRef())

		ses.PendingInputs = make(map[string]model.ResourceRef)

		initializer, err := p.initStage(ses, next)
		if err != nil {
			return err
		}

		msgConfig := utils.TransformMessage(userID, initializer)
		chattables = append(chattables, msgConfig)

		if transitConfig.Clean {
			deletees = append(deletees, ses.StageMessages...)
			ses.Clean()
		}

	} else if currentStageName != "" && messagesSent {
		currentStage, err := p.entityRegistry.GetStage(userID, currentStageName)
		if err != nil {
			return err
		}
		initializer, err := currentStage.Initializer().Get(userID)
		if err != nil {
			return err
		}

		msgConfig := utils.TransformMessage(userID, initializer)
		chattables = append(chattables, msgConfig)
	}

	for _, msg := range chattables {
		sentMsg, err := p.api.Send(msg)
		if err != nil {
			return err
		}
		ses.AppendStageMessage(int64(sentMsg.MessageID))
	}

	for _, ID := range deletees {
		_, err := p.api.DeleteMessage(tgbotapi.NewDeleteMessage(ses.UserID, int(ID)))
		if err != nil {
			fmt.Println(err)
		}
	}
	return p.sessionRepository.Save(ses)
}
