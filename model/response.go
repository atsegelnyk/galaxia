package model

type UserUpdateOption func(*UserUpdate)

type UserUpdate struct {
	UserID                int64
	Transit               *Transit
	Messages              []*Message
	CallbackQueryResponse *CallbackQueryResponse
}

func NewUserResponse(userID int64, options ...UserUpdateOption) *UserUpdate {
	response := &UserUpdate{
		UserID: userID,
	}
	for _, option := range options {
		if option != nil {
			option(response)
		}
	}
	return response
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

type CallbackQueryResponse struct {
	Text            string
	CallbackQueryID string
}

type Transit struct {
	Clean     bool
	TargetRef ResourceRef
}

func (u *UserUpdate) GetUserID() int64 {
	return u.UserID
}

func (u *UserUpdate) GetTransitConfig() *Transit {
	return u.Transit
}

func (u *UserUpdate) GetMessages() []*Message {
	return u.Messages
}

func (u *UserUpdate) GetCallbackResponse() *CallbackQueryResponse {
	return u.CallbackQueryResponse
}
