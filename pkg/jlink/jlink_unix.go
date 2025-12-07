//go:build !windows

package jlink

import (
	"fmt"

	"github.com/ebitengine/purego"
)

func openLibrary(name string) (uintptr, error) {
	// Linux/macOS 下使用 purego.Dlopen
	handle, err := purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return 0, fmt.Errorf("failed to load library %s: %w", name, err)
	}
	return handle, nil
}

func closeLibrary(handle uintptr) {
	purego.Dlclose(handle)
}
