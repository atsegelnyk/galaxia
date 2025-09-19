package model

type UserUpdateOption func(*UserUpdate)

type UserUpdate struct {
	UserID                int64
	Transit               *Transit
	Messages              []*Message
	CallbackQueryResponse *CallbackQueryResponse
	ToDeleteMessages      []int
}

func NewUserUpdate(userID int64, options ...UserUpdateOption) *UserUpdate {
	update := &UserUpdate{
		UserID: userID,
	}
	for _, option := range options {
		if option != nil {
			option(update)
		}
	}
	return update
}

func WithTransit(targetStageRef ResourceRef, clean bool) UserUpdateOption {
	return func(response *UserUpdate) {
		response.Transit = &Transit{
			TargetRef: targetStageRef,
			Clean:     clean,
		}
	}
}

func WithMessages(msg ...*Message) UserUpdateOption {
	return func(response *UserUpdate) {
		response.Messages = append(response.Messages, msg...)
	}
}

func WithCallbackQueryResponse(callbackQueryResponse *CallbackQueryResponse) UserUpdateOption {
	return func(response *UserUpdate) {
		response.CallbackQueryResponse = callbackQueryResponse
	}
}

func WithToDeleteMessages(toDeleteMessages []int) UserUpdateOption {
	return func(response *UserUpdate) {
		response.ToDeleteMessages = toDeleteMessages
	}
}

type CallbackQueryResponse struct {
	Text            string
	CallbackQueryID string
}

type Transit struct {
	Clean     bool
	TargetRef ResourceRef
}
