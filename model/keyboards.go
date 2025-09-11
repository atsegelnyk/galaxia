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
	Text      string
	ActionRef ResourceRef
}

func NewReplyButton(name string) *ReplyButton {
	return &ReplyButton{
		Text: name,
	}
}

func (b *ReplyButton) LinkAction(actionRef ResourceRef) *ReplyButton {
	b.ActionRef = actionRef
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

func (b *InlineButton) LinkCallbackHandler(handlerRef ResourceRef) *InlineButton {
	b.CallbackHandlerRef = handlerRef
	return b
}

func NewKeyboard[T any](layout KeyboardLayout, buttons ...T) [][]T {
	var keyboard [][]T
	switch layout {
	case OnePerRow:
		for _, button := range buttons {
			keyboard = append(keyboard, []T{button})
		}
	case TwoPerRow:
		for i := 0; i < len(buttons); i += 2 {
			if i+1 >= len(buttons) {
				keyboard = append(keyboard, buttons[i:])
				break
			}
			row := []T{buttons[i], buttons[i+1]}
			keyboard = append(keyboard, row)
		}
	case ThreePerRow:
		for i := 0; i < len(buttons); i += 3 {
			if i+2 >= len(buttons) {
				keyboard = append(keyboard, buttons[i:])
				break
			}
			row := []T{buttons[i], buttons[i+1], buttons[i+2]}
			keyboard = append(keyboard, row)
		}
	default:
		for _, button := range buttons {
			keyboard = append(keyboard, []T{button})
		}
	}
	return keyboard
}
