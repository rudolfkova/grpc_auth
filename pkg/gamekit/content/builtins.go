package content

import stdctx "context"

// RegisterBuiltins регистрирует базовые op на раннере.
func RegisterBuiltins(r *Runner) {
	r.RegisterOp("noop", func(stdctx.Context, *RunContext, map[string]any) error { return nil })
	r.RegisterOp("debug_log", func(_ stdctx.Context, rcx *RunContext, args map[string]any) error {
		msg, _ := args["message"].(string)
		if rcx != nil && rcx.Log != nil {
			rcx.Log(msg)
		}
		return nil
	})
}
