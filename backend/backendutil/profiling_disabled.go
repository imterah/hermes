//go:build !debug

package backendutil

var endProfileFunc func()

func ConfigureProfiling() error {
	return nil
}
