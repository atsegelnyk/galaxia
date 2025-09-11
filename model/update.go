package model

type Updater interface {
	GetUserID() int64
	GetTransitConfig() *Transit
	GetMessages() []*Message
	GetCallbackResponse() *CallbackQueryResponse
}
