package bootstrap

type BotSchema struct {
	Auther   *AutherSchema   `json:"auther"`
	Commands []CommandSchema `json:"commands"`
	Stages   []StageSchema   `json:"stages"`
}

type CommandSchema struct {
	Name   string       `json:"name"`
	Action ActionSchema `json:"action"`
}

type StageSchema struct {
	Name        string             `json:"name"`
	Initializer *InitializerSchema `json:"initializer,omitempty"`
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
	Name   string       `json:"name"`
	Action ActionSchema `json:"action"`
}

type InitializerMessageSchema struct {
	Text string `json:"text"`
}

type ActionSchema struct {
	Message string         `json:"message"`
	Transit *TransitSchema `json:"transit,omitempty"`
}

type TransitSchema struct {
	TargetRef string `json:"target_ref"`
	Clean     bool   `json:"clean"`
}

type AutherSchema struct {
	Whitelist []int64 `json:"whitelist"`
	Blacklist []int64 `json:"blacklist"`
}
