package bootstrap

import (
	"bytes"
	"encoding/json"
	"github.com/atsegelnyk/galaxia/auth"
	"github.com/atsegelnyk/galaxia/entityregistry"
	"github.com/atsegelnyk/galaxia/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"os"
	"text/template"
)

func FromFile(path string) (*entityregistry.Registry, error) {
	schema, err := readConfigFromFile(path)
	if err != nil {
		return nil, err
	}
	er := entityregistry.New()

	er.RegisterAuther(
		bootstrapAuther(schema.Auther),
	)

	for _, cmd := range schema.Commands {
		err = bootstrapCommand(cmd, er)
		if err != nil {
			return nil, err
		}
	}
	for _, stg := range schema.Stages {
		err = bootstrapStage(stg, er)
		if err != nil {
			return nil, err
		}
	}
	return er, nil
}

func readConfigFromFile(path string) (*BotSchema, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	configData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var schema BotSchema
	err = json.Unmarshal(configData, &schema)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

func bootstrapCommand(cmdSchema CommandSchema, er *entityregistry.Registry) error {
	cmdAction := bootstrapAction(cmdSchema.Action)
	cmd := model.NewCommand(cmdSchema.Name, cmdAction)
	return er.RegisterCommand(cmd)
}

func bootstrapStage(stageSchema StageSchema, er *entityregistry.Registry) error {
	var stageInitializerOption model.StageOption

	if stageSchema.Initializer != nil {
		var initializerKeyboardOption model.MessageOption
		if stageSchema.Initializer.Keyboard != nil {
			var buttons []*model.ReplyButton
			for _, buttonSchema := range stageSchema.Initializer.Keyboard.Buttons {
				buttons = append(buttons, bootstrapReplyButton(buttonSchema))
			}
			keyboard := model.NewReplyKeyboard(
				bootstrapKeyboardLayout(stageSchema.Initializer.Keyboard.Layout),
				buttons...,
			)
			initializerKeyboardOption = model.WithReplyKeyboard(keyboard)
		}

		message := model.NewMessage(
			model.WithText(stageSchema.Initializer.Message),
			initializerKeyboardOption,
		)
		stageInitializerOption = model.WithInitializer(
			model.NewStaticStageInitializer(message),
		)
	}

	stage := model.NewStage(stageSchema.Name,
		stageInitializerOption,
	)
	return er.RegisterStage(stage)
}

func bootstrapReplyButton(buttonSchema InitializerKeyboardButtonSchema) *model.ReplyButton {
	buttonAction := bootstrapAction(buttonSchema.Action)
	return model.NewReplyButton(buttonSchema.Name).LinkAction(buttonAction)
}

func bootstrapAction(actionSchema ActionSchema) model.UserActionFunc {
	return func(ctx *model.UserContext, update *tgbotapi.Update) model.Updater {
		msgText, err := executeUserTemplate(actionSchema.Message, ctx)
		var message *model.Message
		if err != nil {
			message = model.NewMessage(
				model.WithText(err.Error()),
			)
		} else {
			message = model.NewMessage(model.WithText(msgText))
		}
		var transitOption model.UserUpdateOption
		if actionSchema.Transit != nil {
			transitOption = model.WithTransit(
				model.ResourceRef(actionSchema.Transit.TargetRef),
				actionSchema.Transit.Clean,
			)
		}
		return model.NewUserResponse(update.Message.Chat.ID,
			model.WithMessages(message),
			transitOption,
		)
	}
}

func bootstrapKeyboardLayout(layout string) model.KeyboardLayout {
	switch layout {
	case "ONE_PER_ROW":
		return model.OnePerRow
	case "TWO_PER_ROW":
		return model.TwoPerRow
	case "THREE_PER_ROW":
		return model.ThreePerRow
	case "FOUR_PER_ROW":
		return model.FourPerRow
	case "FIVE_PER_ROW":
		return model.FivePerRow
	default:
		return model.OnePerRow
	}
}

func executeUserTemplate(tplText string, data interface{}) (string, error) {
	msgBuf := bytes.NewBuffer(nil)
	tpl, err := template.New("").Parse(tplText)
	if err != nil {
		return "", err
	}
	err = tpl.Execute(msgBuf, data)
	if err != nil {
		return "", err
	}
	return msgBuf.String(), nil
}

func bootstrapAuther(autherSchema *AutherSchema) auth.Auther {
	if autherSchema.Blacklist != nil {
		return auth.NewBlacklistAuther(autherSchema.Blacklist)
	}
	if autherSchema.Whitelist != nil {
		return auth.NewWhiteListAuther(autherSchema.Whitelist)
	}
	return nil
}
