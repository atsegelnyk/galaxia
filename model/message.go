package model

type Message struct {
	Text  string
	Media []*Media

	ReplyKeyboard  [][]*ReplyButton
	InlineKeyboard [][]*InlineButton
}

type MessageOption func(*Message)

func NewMessage(opts ...MessageOption) *Message {
	msg := &Message{}
	for _, opt := range opts {
		opt(msg)
	}
	return msg
}

type Media struct {
	Photos []byte
	Videos []byte
}

func WithText(text string) MessageOption {
	return func(msg *Message) {
		msg.Text = text
	}
}

func WithReplyKeyboard(keyboard [][]*ReplyButton) MessageOption {
	return func(msg *Message) {
		msg.ReplyKeyboard = keyboard
	}
}

func WithInlineKeyboard(keyboard [][]*InlineButton) MessageOption {
	return func(msg *Message) {
		msg.InlineKeyboard = keyboard
	}
}
