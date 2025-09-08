package processor

import (
	"context"
	"github.com/atsegelnyk/galaxia/auth"
	"github.com/atsegelnyk/galaxia/entityregistry"
	model2 "github.com/atsegelnyk/galaxia/model"
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

func (p *GalaxiaProcessor) RegisterCMD(cmd *model2.Command) error {
	return p.entityRegistry.RegisterCommand(cmd)
}

func (p *GalaxiaProcessor) RegisterStage(stage *model2.Stage) error {
	return p.RegisterStage(stage)
}

func (p *GalaxiaProcessor) Start(ctx context.Context) {
	startCmd := p.entityRegistry.GetCommand(0, StartCMDName)
	if startCmd == nil {
		log.Fatal("start command is not registered")
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

// event processor

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
			return err
		}
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
	cmd := p.entityRegistry.GetCommand(update.Message.Chat.ID, model2.ResourceRef(update.Message.Command()))
	responser := cmd.Handler()(update)
	return p.handleUserResponse(session, responser)
}

func (p *GalaxiaProcessor) processMessage(session *session.Session, update *tgbotapi.Update) error {
	stageName := session.GetCurrentStage()
	if stageName == "" {
		cmd := p.entityRegistry.GetCommand(update.Message.Chat.ID, StartCMDName)
		response := cmd.Handler()(update)
		return p.handleUserResponse(session, response)
	}

	stg := p.entityRegistry.GetStage(update.Message.Chat.ID, stageName)
	responser, err := stg.ProcessUserEvent(update)
	if err != nil {
		log.Println(err)
	}
	return p.handleUserResponse(session, responser)
}

func (p *GalaxiaProcessor) processCallbackQuery(session *session.Session, update *tgbotapi.Update) error {
	callbackHandlerRef, err := session.GetPendingCallback(update.CallbackQuery.Data)
	if err != nil {
		return err
	}

	callbackHandler := p.entityRegistry.GetCallbackHandler(int64(update.CallbackQuery.From.ID), callbackHandlerRef)
	responser := callbackHandler.Func()(update)
	return p.handleUserResponse(session, responser)
}

// user response handler

func (p *GalaxiaProcessor) handleUserResponse(ses *session.Session, responser model2.Responser) error {
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
		err := p.respondMessages(responser)
		messagesSent = true
		if err != nil {
			log.Println(err)
		}
	}

	return p.respondTransit(messagesSent, ses, responser)
}

// callbackID mapper

func (p *GalaxiaProcessor) callbackMapper(ses *session.Session, responser model2.Responser) {
	messages := responser.GetMessages()
	for _, message := range messages {
		if message.InlineKeyboard != nil {
			for _, row := range message.InlineKeyboard {
				for _, button := range row {
					callbackID := utils.GenerateCallbackID()
					button.Data = callbackID
					ses.RegisterCallback(callbackID, button.CallbackHandlerRef)
				}
			}
		}
	}
}

// responders by type

func (p *GalaxiaProcessor) respondCallbackQueryResponse(responser model2.Responser) error {
	if responser.GetCallbackResponse() != nil {
		callBackConfig := utils.TransformCallbackQueryResponse(responser.GetCallbackResponse())
		_, err := p.api.AnswerCallbackQuery(callBackConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *GalaxiaProcessor) respondMessages(responser model2.Responser) error {
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
		_, err := p.api.Send(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *GalaxiaProcessor) respondTransit(messagesSent bool, ses *session.Session, responser model2.Responser) error {
	userID := responser.GetUserID()
	currentStageName := ses.GetCurrentStage()
	transitConfig := responser.GetTransitConfig()

	var chattables []tgbotapi.Chattable

	if transitConfig != nil {
		next := p.entityRegistry.GetStage(userID, transitConfig.TargetRef)
		ses.SetStage(next.SelfRef())

		initializer, err := next.Initializer().Get(userID)
		if err != nil {
			return err
		}

		msgConfig := utils.TransformMessage(userID, initializer)
		chattables = append(chattables, msgConfig)
	} else if currentStageName != "" && messagesSent {
		currentStage := p.entityRegistry.GetStage(userID, currentStageName)
		initializer, err := currentStage.Initializer().Get(userID)
		if err != nil {
			return err
		}

		msgConfig := utils.TransformMessage(userID, initializer)
		chattables = append(chattables, msgConfig)
	}

	for _, msg := range chattables {
		_, err := p.api.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}
