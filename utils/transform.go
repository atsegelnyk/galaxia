package utils

import (
	model2 "galaxia/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func TransformMessage(userID int64, model *model2.Message) *tgbotapi.MessageConfig {
	messageConfig := tgbotapi.NewMessage(userID, model.Text)
	if model.InlineKeyboard != nil {
		keyboard := tgbotapi.NewInlineKeyboardMarkup()
		for _, r := range model.InlineKeyboard {
			row := tgbotapi.NewInlineKeyboardRow()
			for _, b := range r {
				button := tgbotapi.NewInlineKeyboardButtonData(b.Text, b.Data)
				row = append(row, button)
			}
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
		}
		messageConfig.ReplyMarkup = &keyboard
	}
	if model.ReplyKeyboard != nil {
		keyboard := tgbotapi.NewReplyKeyboard()
		for _, r := range model.ReplyKeyboard {
			row := tgbotapi.NewKeyboardButtonRow()
			for _, b := range r {
				button := tgbotapi.KeyboardButton{
					Text: b.Text,
				}
				row = append(row, button)
			}
			keyboard.Keyboard = append(keyboard.Keyboard, row)
		}
		messageConfig.ReplyMarkup = &keyboard
	}
	return &messageConfig
}

func TransformCallbackQueryResponse(model *model2.CallbackQueryResponse) tgbotapi.CallbackConfig {
	return tgbotapi.CallbackConfig{
		Text:            model.Text,
		CallbackQueryID: model.CallbackQueryID,
	}
}
