package content

import (
	stdctx "context"
	"fmt"
	"sync"
)

// RunContext передаётся в каждый op при исполнении сценария.
type RunContext struct {
	PlayerID  int64
	ItemDefID string
	Log       func(string)
	// HostData — указатель на структуру движка (type assert в зарегистрированных op).
	HostData any
}

// Handler обрабатывает один op; args уже смержены (каталог + шаг).
type Handler func(ctx stdctx.Context, rcx *RunContext, args map[string]any) error

// Runner реестр op и исполнитель шагов сценария.
type Runner struct {
	mu  sync.RWMutex
	ops map[string]Handler
}

// NewRunner создаёт раннер с встроенными op (noop, debug_log).
func NewRunner() *Runner {
	r := &Runner{ops: make(map[string]Handler)}
	RegisterBuiltins(r)
	return r
}

// RegisterOp добавляет или заменяет op (обычно из game-service после NewRunner).
func (r *Runner) RegisterOp(name string, h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ops == nil {
		r.ops = make(map[string]Handler)
	}
	r.ops[name] = h
}

// Run выполняет все шаги. Для interact baseArgs обычно = MergeInteractBase(catalog, tile);
// на каждый шаг: mergeShallow(baseArgs, step.Args) — аргументы шага сценария перекрывают каталог и instance_args.
func (r *Runner) Run(ctx stdctx.Context, rcx *RunContext, sc *Scenario, baseArgs map[string]any) error {
	if sc == nil {
		return fmt.Errorf("content: scenario is nil")
	}
	for i, step := range sc.Steps {
		args := mergeShallow(baseArgs, step.Args)
		r.mu.RLock()
		h, ok := r.ops[step.Op]
		r.mu.RUnlock()
		if !ok || h == nil {
			return fmt.Errorf("content: unknown op %q at step %d", step.Op, i)
		}
		if err := h(ctx, rcx, args); err != nil {
			return fmt.Errorf("content: step %d op %s: %w", i, step.Op, err)
		}
	}
	return nil
}
