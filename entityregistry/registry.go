package entityregistry

import (
	"fmt"
	"github.com/atsegelnyk/galaxia/auth"
	"github.com/atsegelnyk/galaxia/model"
	"sync"
)

type Registry struct {
	mu sync.Mutex

	auther           auth.Auther
	cmds             map[model.ResourceRef]*model.Command
	stages           map[model.ResourceRef]*model.Stage
	callbackHandlers map[model.ResourceRef]*model.CallbackHandler

	overrides map[int64]userOverrides
}

type userOverrides struct {
	cmds             map[model.ResourceRef]*model.Command
	stages           map[model.ResourceRef]*model.Stage
	callbackHandlers map[model.ResourceRef]*model.CallbackHandler
}

func New() *Registry {
	entityRegistry := &Registry{
		mu:               sync.Mutex{},
		cmds:             make(map[model.ResourceRef]*model.Command),
		stages:           make(map[model.ResourceRef]*model.Stage),
		callbackHandlers: make(map[model.ResourceRef]*model.CallbackHandler),
		overrides:        make(map[int64]userOverrides),
	}
	return entityRegistry
}

func (e *Registry) RegisterAuther(auther auth.Auther) {
	e.auther = auther
}

func (e *Registry) GetAuther() auth.Auther {
	return e.auther
}

func (e *Registry) RegisterCommand(cmd *model.Command) error {
	if _, ok := e.cmds[cmd.SelfRef()]; ok {
		return fmt.Errorf("command %s already exists", cmd.SelfRef())
	}
	e.mu.Lock()
	e.cmds[cmd.SelfRef()] = cmd
	e.mu.Unlock()
	return nil
}

func (e *Registry) RegisterStage(stg *model.Stage) error {
	if _, ok := e.stages[stg.SelfRef()]; ok {
		return fmt.Errorf("stage %s already exists", stg.SelfRef())
	}
	e.mu.Lock()
	e.stages[stg.SelfRef()] = stg
	e.mu.Unlock()
	return nil
}

func (e *Registry) RegisterCallbackHandler(handler *model.CallbackHandler) error {
	if _, ok := e.callbackHandlers[handler.SelfRef()]; ok {
		return fmt.Errorf("callback handler %s already exists", handler.SelfRef())
	}
	e.mu.Lock()
	e.callbackHandlers[handler.SelfRef()] = handler
	e.mu.Unlock()
	return nil
}

func (e *Registry) OverrideCommand(cmd *model.Command, users ...int64) {
	if len(users) == 0 {
		e.mu.Lock()
		e.cmds[cmd.SelfRef()] = cmd
		e.mu.Unlock()
		return
	}

	for _, user := range users {
		e.checkInitUserOverrides(user)
		uo := e.overrides[user]
		uo.cmds[cmd.SelfRef()] = cmd
	}
}

func (e *Registry) OverrideStage(stg *model.Stage, users ...int64) {
	if len(users) == 0 {
		e.mu.Lock()
		e.stages[stg.SelfRef()] = stg
		e.mu.Unlock()
		return
	}

	for _, user := range users {
		e.checkInitUserOverrides(user)
		uo := e.overrides[user]
		uo.stages[stg.SelfRef()] = stg
	}
}

func (e *Registry) OverrideCallbackHandler(handler *model.CallbackHandler, users ...int64) {
	if len(users) == 0 {
		e.mu.Lock()
		e.callbackHandlers[handler.SelfRef()] = handler
		e.mu.Unlock()
		return
	}

	for _, user := range users {
		e.checkInitUserOverrides(user)
		uo := e.overrides[user]
		uo.callbackHandlers[handler.SelfRef()] = handler
	}
}

func (e *Registry) GetCommand(userID int64, cmdRef model.ResourceRef) (*model.Command, error) {
	if override, ok := e.overrides[userID]; ok {
		if cmd, overrideOk := override.cmds[cmdRef]; overrideOk {
			return cmd, nil
		}
	}
	if cmd, ok := e.cmds[cmdRef]; ok {
		return cmd, nil
	}
	return nil, fmt.Errorf("cmd %v not found", cmdRef)
}

func (e *Registry) GetStage(userID int64, stageRef model.ResourceRef) (*model.Stage, error) {
	if override, ok := e.overrides[userID]; ok {
		if stg, overrideOk := override.stages[stageRef]; overrideOk {
			return stg, nil
		}
	}
	if stg, ok := e.stages[stageRef]; ok {
		return stg, nil
	}
	return nil, fmt.Errorf("stage %v not found", stageRef)
}

func (e *Registry) GetCallbackHandler(userID int64, callbackRef model.ResourceRef) (*model.CallbackHandler, error) {
	if override, ok := e.overrides[userID]; ok {
		if cb, overrideOk := override.callbackHandlers[callbackRef]; overrideOk {
			return cb, nil
		}
	}
	if cb, ok := e.callbackHandlers[callbackRef]; ok {
		return cb, nil
	}
	return nil, fmt.Errorf("callback handler %v not found", callbackRef)
}

func (e *Registry) checkInitUserOverrides(userID int64) {
	if _, ok := e.overrides[userID]; ok {
		return
	}
	uo := userOverrides{
		cmds:             make(map[model.ResourceRef]*model.Command),
		stages:           make(map[model.ResourceRef]*model.Stage),
		callbackHandlers: make(map[model.ResourceRef]*model.CallbackHandler),
	}
	e.mu.Lock()
	e.overrides[userID] = uo
	e.mu.Unlock()
}
