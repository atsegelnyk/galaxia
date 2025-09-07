package model

const (
	OnePerRow KeyboardLayout = iota
	TwoPerRow
	ThreePerRow
	FourPerRow
	FivePerRow
	All
	Custom
)

type KeyboardLayout int

type ReplyButton struct {
	Text   string
	Action UserAction
}

func NewReplyButton(name string) *ReplyButton {
	return &ReplyButton{
		Text: name,
	}
}

func NewReplyKeyboard(layout KeyboardLayout, buttons ...*ReplyButton) [][]*ReplyButton {
	var keyboard [][]*ReplyButton
	switch layout {
	case OnePerRow:
		for _, button := range buttons {
			row := []*ReplyButton{button}
			keyboard = append(keyboard, row)
		}
	case TwoPerRow:
		for i := 0; i < len(buttons); i += 2 {
			if i+1 >= len(buttons) {
				keyboard = append(keyboard, buttons[i:])
				break
			}
			row := []*ReplyButton{buttons[i], buttons[i+1]}
			keyboard = append(keyboard, row)
		}
	case ThreePerRow:
		for i := 0; i < len(buttons); i += 3 {
			if i+2 >= len(buttons) {
				keyboard = append(keyboard, buttons[i:])
				break
			}
			row := []*ReplyButton{buttons[i], buttons[i+1], buttons[i+2]}
			keyboard = append(keyboard, row)
		}
	default:
		for _, button := range buttons {
			row := []*ReplyButton{button}
			keyboard = append(keyboard, row)
		}
	}
	return keyboard
}

func (b *ReplyButton) LinkAction(action UserAction) *ReplyButton {
	b.Action = action
	return b
}

type InlineButton struct {
	Text    string
	Data    string
	Context *CallbackContext
}

type CallbackContext struct {
	Retain   bool
	Callback UserCallback
	Misc     interface{}
}

func NewInlineButton(name string, context *CallbackContext) *InlineButton {
	return &InlineButton{
		Text:    name,
		Context: context,
	}
}

func (b *InlineButton) LinkAction(action UserCallback) *InlineButton {
	b.Context.Callback = action
	return b
}

func (b *InlineButton) AddMisc(misc interface{}) *InlineButton {
	b.Context.Misc = misc
	return b
}

func NewInlineKeyboard(layout KeyboardLayout, buttons ...*InlineButton) [][]*InlineButton {
	var keyboard [][]*InlineButton
	switch layout {
	case OnePerRow:
		for _, button := range buttons {
			row := []*InlineButton{button}
			keyboard = append(keyboard, row)
		}
	case TwoPerRow:
		for i := 0; i < len(buttons); i += 2 {
			if i+1 > len(buttons) {
				keyboard = append(keyboard, buttons[i:])
				break
			}
			row := []*InlineButton{buttons[i], buttons[i+1]}
			keyboard = append(keyboard, row)
		}
	case ThreePerRow:
		for i := 0; i < len(buttons); i += 3 {
			if i+2 > len(buttons) {
				keyboard = append(keyboard, buttons[i:])
				break
			}
			row := []*InlineButton{buttons[i], buttons[i+1], buttons[i+2]}
			keyboard = append(keyboard, row)
		}
	default:
		for _, button := range buttons {
			row := []*InlineButton{button}
			keyboard = append(keyboard, row)
		}
	}
	return keyboard
}
