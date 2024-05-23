package actions

import "context"

// Reinstall represents the action and its options.
type Reinstall struct {
	Install
}

// Run runs the action.
func (action *Reinstall) Run(ctx context.Context, names ...string) error {
	action.Force = true
	return action.Install.Run(ctx, names...)
}
