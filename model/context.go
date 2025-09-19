package model

type UserContext struct {
	UserID   int64                  `json:"user_id"`
	Lang     string                 `json:"lang,omitempty"`
	Name     string                 `json:"name,omitempty"`
	LastName string                 `json:"last_name,omitempty"`
	Username string                 `json:"username,omitempty"`
	Misc     map[string]interface{} `json:"misc,omitempty"`

	CallbackData *string
}
