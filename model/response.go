package model

type Responser interface {
	GetUserID() int64
	GetTransitConfig() *Transit
	GetMessages() []*Message
	GetCallbackResponse() *CallbackQueryResponse
}

type UserResponseOption func(*UserResponse)

type UserResponse struct {
	UserID                int64
	Transit               *Transit
	Messages              []*Message
	CallbackQueryResponse *CallbackQueryResponse
}

func NewUserResponse(userID int64, options ...UserResponseOption) *UserResponse {
	response := &UserResponse{
		UserID: userID,
	}
	for _, option := range options {
		option(response)
	}
	return response
}

func WithTransitTarget(targetStageRef ResourceRef) UserResponseOption {
	return func(response *UserResponse) {
		response.Transit = &Transit{
			TargetRef: targetStageRef,
		}
	}
}

func WithMessages(msg ...*Message) UserResponseOption {
	return func(response *UserResponse) {
		response.Messages = append(response.Messages, msg...)
	}
}

func WithCallbackQueryResponse(callbackQueryResponse *CallbackQueryResponse) UserResponseOption {
	return func(response *UserResponse) {
		response.CallbackQueryResponse = callbackQueryResponse
	}
}

type CallbackQueryResponse struct {
	Text            string
	CallbackQueryID string
}

type Transit struct {
	TargetRef ResourceRef
}

func (u *UserResponse) GetUserID() int64 {
	return u.UserID
}

func (u *UserResponse) GetTransitConfig() *Transit {
	return u.Transit
}

func (u *UserResponse) GetMessages() []*Message {
	return u.Messages
}

func (u *UserResponse) GetCallbackResponse() *CallbackQueryResponse {
	return u.CallbackQueryResponse
}
