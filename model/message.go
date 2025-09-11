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

func (m *Message) WithText(text string) *Message {
	m.Text = text
	return m
}

func WithReplyKeyboard(keyboard [][]*ReplyButton) MessageOption {
	return func(msg *Message) {
		msg.ReplyKeyboard = keyboard
	}
}

func (m *Message) WithReplyKeyboard(keyboard [][]*ReplyButton) *Message {
	m.ReplyKeyboard = keyboard
	return m
}

func WithInlineKeyboard(keyboard [][]*InlineButton) MessageOption {
	return func(msg *Message) {
		msg.InlineKeyboard = keyboard
	}
}

func (m *Message) WithInlineKeyboard(keyboard [][]*InlineButton) *Message {
	m.InlineKeyboard = keyboard
	return m
}

func WithPhoto(photo []byte) MessageOption {
	return func(msg *Message) {
		msg.Photo = photo
	}

}

func (m *Message) WithPhoto(photo []byte) *Message {
	m.Photo = photo
	return m
}

func WithVideo(video []byte) MessageOption {
	return func(msg *Message) {
		msg.Video = video
	}
}

func (m *Message) WithVideo(video []byte) *Message {
	m.Video = video
	return m
}
