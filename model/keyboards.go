package model

type KeyboardLayout int

const (
	OnePerRow KeyboardLayout = iota
	TwoPerRow
	ThreePerRow
	FourPerRow
	FivePerRow
	Custom
)

type CallbackBehaviour int

const (
	Retain CallbackBehaviour = iota
	DeleteCallbackBehaviour
)

type ReplyButton struct {
	Text   string
	Action UserActionFunc
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

func (b *ReplyButton) LinkAction(action UserActionFunc) *ReplyButton {
	b.Action = action
	return b
}

type InlineButton struct {
	Text string
	Data string

	CallbackBehaviour  CallbackBehaviour
	CallbackHandlerRef ResourceRef
}

func NewInlineButton(name string) *InlineButton {
	return &InlineButton{
		Text: name,
	}
}

func (b *InlineButton) LinkCallbackHandler(handlerRef Referencer) *InlineButton {
	b.CallbackHandlerRef = handlerRef.SelfRef()
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
