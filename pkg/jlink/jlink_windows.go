//go:build windows

package jlink

import (
	"fmt"
	"syscall"
)

func openLibrary(name string) (uintptr, error) {
	// Windows 下使用 syscall 加载 DLL
	handle, err := syscall.LoadDLL(name)
	if err != nil {
		return 0, fmt.Errorf("failed to load library %s: %w", name, err)
	}
	return uintptr(handle), nil
}

func closeLibrary(handle uintptr) {
	syscall.FreeLibrary(syscall.Handle(handle))
}
