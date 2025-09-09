package processor

import (
	"context"
	"errors"
	"fmt"
	"github.com/atsegelnyk/galaxia/auth"
	"github.com/atsegelnyk/galaxia/entityregistry"
	"github.com/atsegelnyk/galaxia/model"
	"github.com/atsegelnyk/galaxia/session"
	"github.com/atsegelnyk/galaxia/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

const StartCMDName = "start"

type GalaxiaProcessorOption func(*GalaxiaProcessor)

type GalaxiaProcessor struct {
	api *tgbotapi.BotAPI

	auther            auth.Auther
	sessionRepository session.Repository
	entityRegistry    *entityregistry.Registry
}

func NewGalaxiaProcessor(opts ...GalaxiaProcessorOption) *GalaxiaProcessor {
	g := &GalaxiaProcessor{
		entityRegistry: entityregistry.New(),
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func WithApi(api *tgbotapi.BotAPI) GalaxiaProcessorOption {
	return func(g *GalaxiaProcessor) {
		g.api = api
	}
}

func WithAuthHandler(auther auth.Auther) GalaxiaProcessorOption {
	return func(g *GalaxiaProcessor) {
		g.auther = auther
	}
}

func WithBotToken(token string) GalaxiaProcessorOption {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	return func(g *GalaxiaProcessor) {
		g.api = api
	}
}

func WithEntityRegistry(er *entityregistry.Registry) GalaxiaProcessorOption {
	return func(g *GalaxiaProcessor) {
		g.entityRegistry = er
	}
}

func WithSessionRepository(r session.Repository) GalaxiaProcessorOption {
	return func(g *GalaxiaProcessor) {
		g.sessionRepository = r
	}
}

func (p *GalaxiaProcessor) RegisterCMD(cmd *model.Command) error {
	return p.entityRegistry.RegisterCommand(cmd)
}

func (p *GalaxiaProcessor) RegisterStage(stage *model.Stage) error {
	return p.RegisterStage(stage)
}

func (p *GalaxiaProcessor) Start(ctx context.Context) {
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

func (p *GalaxiaProcessor) preflightCheck() error {
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

func (p *GalaxiaProcessor) initSession(update *tgbotapi.Update) *session.Session {
	return session.NewSession(
		update.Message.Chat.ID,
		session.WithUsername(update.Message.From.UserName),
		session.WithLang(update.Message.From.LanguageCode),
		session.WithName(update.Message.From.FirstName),
		session.WithLastName(update.Message.From.LastName),
	)
}

func (p *GalaxiaProcessor) processUpdate(update *tgbotapi.Update) error {
	if update.Message != nil {
		if p.auther != nil {
			err := p.auther.Authorize(update.Message.Chat.ID)
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

func (p *GalaxiaProcessor) processCmd(session *session.Session, update *tgbotapi.Update) error {
	cmd, err := p.entityRegistry.GetCommand(update.Message.Chat.ID, model.ResourceRef(update.Message.Command()))
	if err != nil {
		return err
	}
	responser := cmd.Handler()(session.UserContext, update)
	return p.handleUserResponse(session, responser)
}

func (p *GalaxiaProcessor) processMessage(session *session.Session, update *tgbotapi.Update) error {
	stageRef := session.GetCurrentStage()
	if stageRef.Empty() {
		cmd, _ := p.entityRegistry.GetCommand(update.Message.Chat.ID, StartCMDName)
		response := cmd.Handler()(session.UserContext, update)
		return p.handleUserResponse(session, response)
	}

	stg, err := p.entityRegistry.GetStage(update.Message.Chat.ID, stageRef)
	if err != nil {
		return err
	}
	responser, err := stg.ProcessUserEvent(session.UserContext, update)
	if err != nil {
		log.Println(err)
	}
	return p.handleUserResponse(session, responser)
}

func (p *GalaxiaProcessor) processCallbackQuery(session *session.Session, update *tgbotapi.Update) error {
	callbackHandlerRef, err := session.GetPendingCallbackHandler(update.CallbackQuery.Data)
	if err != nil {
		return err
	}

	callbackHandler, err := p.entityRegistry.GetCallbackHandler(int64(update.CallbackQuery.From.ID), callbackHandlerRef)
	if err != nil {
		return err
	}
	responser := callbackHandler.Func()(session.UserContext, update)
	return p.handleUserResponse(session, responser)
}

// user response handler

func (p *GalaxiaProcessor) handleUserResponse(ses *session.Session, responser model.Responser) error {
	if responser == nil {
		return nil
	}
	messagesSent := false

	if responser.GetCallbackResponse() != nil {
		err := p.respondCallbackQueryResponse(responser)
		if err != nil {
			log.Println(err)
		}
	}

	if responser.GetMessages() != nil {
		p.callbackMapper(ses, responser)
		err := p.respondMessages(ses, responser)
		messagesSent = true
		if err != nil {
			log.Println(err)
		}
	}

	return p.respondTransit(messagesSent, ses, responser)
}

// callbackID mapper

func (p *GalaxiaProcessor) callbackMapper(ses *session.Session, responser model.Responser) {
	messages := responser.GetMessages()
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

func (p *GalaxiaProcessor) respondCallbackQueryResponse(responser model.Responser) error {
	if responser.GetCallbackResponse() != nil {
		callBackConfig := utils.TransformCallbackQueryResponse(responser.GetCallbackResponse())
		_, err := p.api.AnswerCallbackQuery(callBackConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *GalaxiaProcessor) respondMessages(ses *session.Session, responser model.Responser) error {
	var chattables []tgbotapi.Chattable

	userID := responser.GetUserID()
	messages := responser.GetMessages()

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

func (p *GalaxiaProcessor) respondTransit(messagesSent bool, ses *session.Session, responser model.Responser) error {
	userID := responser.GetUserID()
	currentStageName := ses.GetCurrentStage()
	transitConfig := responser.GetTransitConfig()

	var chattables []tgbotapi.Chattable
	var deletees []int64

	if transitConfig != nil {
		next, err := p.entityRegistry.GetStage(userID, transitConfig.TargetRef)
		if err != nil {
			return err
		}
		ses.SetNextStage(next.SelfRef())

		initializer, err := next.Initializer().Get(userID)
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
