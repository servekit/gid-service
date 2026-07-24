// Package option defines functional options for constructing the service.
//
// gid-service currently has no injectable heavy resources (the snowflake
// generator is built from cfg.Snowflake and has no Close). The Options struct
// stays empty until a resource that benefits from caller injection (Redis,
// external clients, etc.) is added — at which point WithXxx accessors land
// here without breaking existing callers.
package option

// Option mutates Options.
type Option func(*Options)

// Options holds resolved dependencies for service construction.
type Options struct{}

// Apply evaluates all options and returns the resolved Options. A nil field
// means "not injected — service owns it".
func Apply(opts ...Option) Options {
	var o Options
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
