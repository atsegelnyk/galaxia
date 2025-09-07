package entityregistry

import (
	"fmt"
	"github.com/atsegelnyk/galaxia/model"
	"sync"
)

type Registry struct {
	mu sync.Mutex

	cmds   map[string]*model.Command
	stages map[string]*model.Stage

	overrides map[int64]userOverrides
}

type userOverrides struct {
	cmds   map[string]*model.Command
	stages map[string]*model.Stage
}

func New() *Registry {
	entityRegistry := &Registry{
		mu:        sync.Mutex{},
		cmds:      make(map[string]*model.Command),
		stages:    make(map[string]*model.Stage),
		overrides: make(map[int64]userOverrides),
	}
	return entityRegistry
}

func (e *Registry) RegisterCommand(cmd *model.Command) error {
	if _, ok := e.cmds[cmd.Name()]; ok {
		return fmt.Errorf("command %s already exists", cmd.Name())
	}
	e.mu.Lock()
	e.cmds[cmd.Name()] = cmd
	e.mu.Unlock()
	return nil
}

func (e *Registry) RegisterStage(stg *model.Stage) error {
	if _, ok := e.stages[stg.Name()]; ok {
		return fmt.Errorf("stage %s already exists", stg.Name())
	}
	e.mu.Lock()
	e.stages[stg.Name()] = stg
	e.mu.Unlock()
	return nil
}

func (e *Registry) OverrideCommand(cmd *model.Command, users ...int64) {
	if len(users) == 0 {
		e.mu.Lock()
		e.cmds[cmd.Name()] = cmd
		e.mu.Unlock()
		return
	}

	for _, user := range users {
		if _, ok := e.overrides[user]; !ok {
			uo := userOverrides{
				cmds:   make(map[string]*model.Command),
				stages: make(map[string]*model.Stage),
			}
			uo.cmds[cmd.Name()] = cmd
			e.mu.Lock()
			e.overrides[user] = uo
			e.mu.Unlock()
			continue
		}
		uo := e.overrides[user]
		uo.cmds[cmd.Name()] = cmd
	}
}

func (e *Registry) OverrideStage(stg *model.Stage, users ...int64) {
	if len(users) == 0 {
		e.mu.Lock()
		e.stages[stg.Name()] = stg
		e.mu.Unlock()
		return
	}

	for _, user := range users {
		if _, ok := e.overrides[user]; !ok {
			uo := userOverrides{
				cmds:   make(map[string]*model.Command),
				stages: make(map[string]*model.Stage),
			}
			uo.stages[stg.Name()] = stg
			e.mu.Lock()
			e.overrides[user] = uo
			e.mu.Unlock()
			continue
		}
		uo := e.overrides[user]
		uo.stages[stg.Name()] = stg
	}
}

func (e *Registry) GetCommand(userID int64, name string) *model.Command {
	if override, ok := e.overrides[userID]; ok {
		return override.cmds[name]
	}
	return e.cmds[name]
}

func (e *Registry) GetStage(userID int64, name string) *model.Stage {
	if override, ok := e.overrides[userID]; ok {
		return override.stages[name]
	}
	return e.stages[name]
}
