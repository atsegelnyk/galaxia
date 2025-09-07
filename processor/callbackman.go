package processor

import (
	"errors"
	"galaxia/model"
	"sync"
)

var CallbackNotFoundErr = errors.New("callback not found")

type CallbackManager struct {
	mu              sync.Mutex
	userCallbackMap map[int64]UserCallbacks
}

type UserCallbacks map[string]*PendingCallback

type PendingCallback struct {
	Input   string
	Context *model.CallbackContext
}

func NewCallback(input string, cbCtx *model.CallbackContext) *PendingCallback {
	return &PendingCallback{
		Input:   input,
		Context: cbCtx,
	}
}

func NewCallbackManager() *CallbackManager {
	return &CallbackManager{
		userCallbackMap: make(map[int64]UserCallbacks),
	}
}

func (c *CallbackManager) AddUserCallback(userId int64, callbackID string, callback *PendingCallback) {
	if _, ok := c.userCallbackMap[userId]; ok {
		c.mu.Lock()
		c.userCallbackMap[userId][callbackID] = callback
		c.mu.Unlock()
		return
	}

	c.userCallbackMap[userId] = make(UserCallbacks)
	c.mu.Lock()
	c.userCallbackMap[userId][callbackID] = callback
	c.mu.Unlock()
	return
}

func (c *CallbackManager) GetUserCallback(userId int64, data string) (*PendingCallback, error) {
	if user, ok := c.userCallbackMap[userId]; ok {
		if callback, cbOK := user[data]; cbOK {
			if !callback.Context.Retain {
				delete(user, data)
			}
			return callback, nil
		}
		return nil, CallbackNotFoundErr
	}
	return nil, CallbackNotFoundErr
}
