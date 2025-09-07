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

type AuthHandler func(userID int64) bool

type GalaxiaProcessorOption func(*GalaxiaProcessor)

type GalaxiaProcessor struct {
	api *tgbotapi.BotAPI

	auther         auth.Auther
	callbackMan    *CallbackManager
	sessionMan     *session.Manager
	entityRegistry *entityregistry.Registry
}

func NewGalaxiaProcessor(opts ...GalaxiaProcessorOption) *GalaxiaProcessor {
	g := &GalaxiaProcessor{
		sessionMan:     session.NewManager(),
		entityRegistry: entityregistry.New(),
		callbackMan:    NewCallbackManager(),
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
		err := p.auther.Authorize(update.Message.Chat.ID)
		if err != nil {
			return err
		}

		ses := p.sessionMan.GetForUserID(update.Message.Chat.ID)
		if update.Message.Command() != "" {
			return p.processCmd(ses, update)
		}

		return p.processMessage(ses, update)
	}

	if update.CallbackQuery != nil {
		ses := p.sessionMan.GetForUserID(int64(update.CallbackQuery.From.ID))
		return p.processCallbackQuery(ses, update)
	}
	return nil
}

// event processors by type

func (p *GalaxiaProcessor) processCmd(session *session.Session, update *tgbotapi.Update) error {
	cmd := p.entityRegistry.GetCommand(update.Message.Chat.ID, update.Message.Command())
	responser := cmd.Handler()(update)
	return p.handleUserResponse(session, responser)
}

func (p *GalaxiaProcessor) processMessage(session *session.Session, update *tgbotapi.Update) error {
	stg := session.GetCurrentStage()
	if stg == nil {
		cmd := p.entityRegistry.GetCommand(update.Message.Chat.ID, StartCMDName)
		response := cmd.Handler()(update)
		return p.handleUserResponse(session, response)
	}

	responser, err := stg.ProcessUserEvent(update)
	if err != nil {
		log.Println(err)
	}
	return p.handleUserResponse(session, responser)
}

func (p *GalaxiaProcessor) processCallbackQuery(session *session.Session, update *tgbotapi.Update) error {
	callback, err := p.callbackMan.GetUserCallback(int64(update.CallbackQuery.From.ID), update.CallbackQuery.Data)
	if err != nil {
		log.Println(err)
		return err
	}

	responser := callback.Context.Callback(update, callback.Context.Misc)
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
		p.callbackMapper(responser)
		err := p.respondMessages(responser)
		messagesSent = true
		if err != nil {
			log.Println(err)
		}
	}

	return p.respondTransit(messagesSent, ses, responser)
}

// callbackID mapper

func (p *GalaxiaProcessor) callbackMapper(responser model2.Responser) {
	messages := responser.GetMessages()
	for _, message := range messages {
		if message.InlineKeyboard != nil {
			for _, row := range message.InlineKeyboard {
				for _, button := range row {
					callbackID := utils.GenerateCallbackID()
					button.Data = callbackID
					p.callbackMan.AddUserCallback(responser.GetUserID(),
						callbackID,
						NewCallback(button.Text, button.Context),
					)
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
	currentStage := ses.GetCurrentStage()
	transitConfig := responser.GetTransitConfig()

	var chattables []tgbotapi.Chattable

	if transitConfig != nil {
		next := p.entityRegistry.GetStage(userID, transitConfig.TargetRef)
		ses.SetStage(next)

		initializer, err := next.Initializer().Get(userID)
		if err != nil {
			return err
		}

		msgConfig := utils.TransformMessage(userID, initializer)
		chattables = append(chattables, msgConfig)
	} else if currentStage != nil && messagesSent {
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
