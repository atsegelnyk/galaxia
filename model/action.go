package model

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type UserCallback func(update *tgbotapi.Update, misc interface{}) Responser

type UserAction func(update *tgbotapi.Update) Responser

type InputHandler func(update *tgbotapi.Update) Responser
