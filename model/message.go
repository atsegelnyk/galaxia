package model

type Message struct {
	Text  string
	Photo []byte
	Video []byte

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

func WithPhoto(photo []byte) MessageOption {
	return func(msg *Message) {
		msg.Photo = photo
	}

}

func WithVideo(video []byte) MessageOption {
	return func(msg *Message) {
		msg.Video = video
	}
}
