package entityregistry

import (
	"fmt"
	"github.com/atsegelnyk/galaxia/model"
	"sync"
)

type Registry struct {
	mu sync.Mutex

	actions          map[model.ResourceRef]*model.Action
	cmds             map[model.ResourceRef]*model.Command
	stages           map[model.ResourceRef]*model.Stage
	callbackHandlers map[model.ResourceRef]*model.CallbackHandler

	overrides map[int64]userOverrides
}

type userOverrides struct {
	actions          map[model.ResourceRef]*model.Action
	cmds             map[model.ResourceRef]*model.Command
	stages           map[model.ResourceRef]*model.Stage
	callbackHandlers map[model.ResourceRef]*model.CallbackHandler
}

func New() *Registry {
	entityRegistry := &Registry{
		mu:               sync.Mutex{},
		actions:          make(map[model.ResourceRef]*model.Action),
		cmds:             make(map[model.ResourceRef]*model.Command),
		stages:           make(map[model.ResourceRef]*model.Stage),
		callbackHandlers: make(map[model.ResourceRef]*model.CallbackHandler),
		overrides:        make(map[int64]userOverrides),
	}
	return entityRegistry
}

func (r *Registry) RegisterCommand(cmd *model.Command) error {
	if _, ok := r.cmds[cmd.SelfRef()]; ok {
		return fmt.Errorf("command %s already exists", cmd.SelfRef())
	}
	r.mu.Lock()
	r.cmds[cmd.SelfRef()] = cmd
	r.mu.Unlock()
	return nil
}

func (r *Registry) RegisterStage(stg *model.Stage) error {
	if _, ok := r.stages[stg.SelfRef()]; ok {
		return fmt.Errorf("stage %s already exists", stg.SelfRef())
	}
	r.mu.Lock()
	r.stages[stg.SelfRef()] = stg
	r.mu.Unlock()
	return nil
}

func (r *Registry) RegisterCallbackHandler(handler *model.CallbackHandler) error {
	if _, ok := r.callbackHandlers[handler.SelfRef()]; ok {
		return fmt.Errorf("callback handler %s already exists", handler.SelfRef())
	}
	r.mu.Lock()
	r.callbackHandlers[handler.SelfRef()] = handler
	r.mu.Unlock()
	return nil
}

func (r *Registry) RegisterAction(act *model.Action) error {
	if _, ok := r.actions[act.SelfRef()]; ok {
		return fmt.Errorf("action %s already exists", act.SelfRef())
	}
	r.mu.Lock()
	r.actions[act.SelfRef()] = act
	r.mu.Unlock()
	return nil
}

func (r *Registry) OverrideCommand(cmd *model.Command, users ...int64) {
	if len(users) == 0 {
		r.mu.Lock()
		r.cmds[cmd.SelfRef()] = cmd
		r.mu.Unlock()
		return
	}

	for _, user := range users {
		r.checkInitUserOverrides(user)
		uo := r.overrides[user]
		uo.cmds[cmd.SelfRef()] = cmd
	}
}

func (r *Registry) OverrideStage(stg *model.Stage, users ...int64) {
	if len(users) == 0 {
		r.mu.Lock()
		r.stages[stg.SelfRef()] = stg
		r.mu.Unlock()
		return
	}

	for _, user := range users {
		r.checkInitUserOverrides(user)
		uo := r.overrides[user]
		uo.stages[stg.SelfRef()] = stg
	}
}

func (r *Registry) OverrideCallbackHandler(handler *model.CallbackHandler, users ...int64) {
	if len(users) == 0 {
		r.mu.Lock()
		r.callbackHandlers[handler.SelfRef()] = handler
		r.mu.Unlock()
		return
	}

	for _, user := range users {
		r.checkInitUserOverrides(user)
		uo := r.overrides[user]
		uo.callbackHandlers[handler.SelfRef()] = handler
	}
}

func (r *Registry) OverrideAction(act *model.Action, users ...int64) {
	if len(users) == 0 {
		r.mu.Lock()
		r.actions[act.SelfRef()] = act
		r.mu.Unlock()
		return
	}
	for _, user := range users {
		r.checkInitUserOverrides(user)
		uo := r.overrides[user]
		uo.actions[act.SelfRef()] = act
	}
}

func (r *Registry) GetCommand(userID int64, cmdRef model.ResourceRef) (*model.Command, error) {
	if override, ok := r.overrides[userID]; ok {
		if cmd, overrideOk := override.cmds[cmdRef]; overrideOk {
			return cmd, nil
		}
	}
	if cmd, ok := r.cmds[cmdRef]; ok {
		return cmd, nil
	}
	return nil, fmt.Errorf("cmd %v not found", cmdRef)
}

func (r *Registry) GetStage(userID int64, stageRef model.ResourceRef) (*model.Stage, error) {
	if override, ok := r.overrides[userID]; ok {
		if stg, overrideOk := override.stages[stageRef]; overrideOk {
			return stg, nil
		}
	}
	if stg, ok := r.stages[stageRef]; ok {
		return stg, nil
	}
	return nil, fmt.Errorf("stage %v not found", stageRef)
}

func (r *Registry) GetCallbackHandler(userID int64, callbackRef model.ResourceRef) (*model.CallbackHandler, error) {
	if override, ok := r.overrides[userID]; ok {
		if cb, overrideOk := override.callbackHandlers[callbackRef]; overrideOk {
			return cb, nil
		}
	}
	if cb, ok := r.callbackHandlers[callbackRef]; ok {
		return cb, nil
	}
	return nil, fmt.Errorf("callback handler %v not found", callbackRef)
}

func (r *Registry) GetAction(userID int64, actionRef model.ResourceRef) (*model.Action, error) {
	if override, ok := r.overrides[userID]; ok {
		if act, overrideOk := override.actions[actionRef]; overrideOk {
			return act, nil
		}
	}
	if act, ok := r.actions[actionRef]; ok {
		return act, nil
	}
	return nil, fmt.Errorf("action %s not found", actionRef)
}

func (r *Registry) checkInitUserOverrides(userID int64) {
	if _, ok := r.overrides[userID]; ok {
		return
	}
	uo := userOverrides{
		cmds:             make(map[model.ResourceRef]*model.Command),
		stages:           make(map[model.ResourceRef]*model.Stage),
		callbackHandlers: make(map[model.ResourceRef]*model.CallbackHandler),
	}
	r.mu.Lock()
	r.overrides[userID] = uo
	r.mu.Unlock()
}
