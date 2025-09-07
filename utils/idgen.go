package utils

import gonanoid "github.com/matoous/go-nanoid/v2"

func GenerateCallbackID() string {
	id, _ := gonanoid.New()
	return id
}
