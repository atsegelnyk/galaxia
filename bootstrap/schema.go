package bootstrap

type BotSchema struct {
	Auther           *AutherSchema           `json:"auther"`
	Actions          []ActionSchema          `json:"actions"`
	CallbackHandlers []CallbackHandlerSchema `json:"callback_handlers"`
	Commands         []CommandSchema         `json:"commands"`
	Stages           []StageSchema           `json:"stages"`
}

type ActionSchema struct {
	Name    string         `json:"name"`
	Message string         `json:"message"`
	Transit *TransitSchema `json:"transit,omitempty"`
}

type CallbackHandlerSchema struct {
	Name      string `json:"name"`
	ActionRef string `json:"action_ref"`
}

type CommandSchema struct {
	Name      string `json:"name"`
	ActionRef string `json:"action_ref"`
}

type StageSchema struct {
	Name             string             `json:"name"`
	DefaultActionRef string             `json:"default_action_ref,omitempty"`
	InputAllowed     bool               `json:"input_allowed,omitempty"`
	Initializer      *InitializerSchema `json:"initializer,omitempty"`
}

type InitializerSchema struct {
	Message  string                     `json:"message"`
	Keyboard *InitializerKeyboardSchema `json:"keyboard"`
}

type InitializerKeyboardSchema struct {
	Layout  string                            `json:"layout"`
	Buttons []InitializerKeyboardButtonSchema `json:"buttons"`
}

type InitializerKeyboardButtonSchema struct {
	Name      string `json:"name"`
	ActionRef string `json:"action_ref"`
}

type InitializerMessageSchema struct {
	Text string `json:"text"`
}

type TransitSchema struct {
	TargetRef string `json:"target_ref"`
	Clean     bool   `json:"clean"`
}

type AutherSchema struct {
	Whitelist []int64 `json:"whitelist"`
	Blacklist []int64 `json:"blacklist"`
}
