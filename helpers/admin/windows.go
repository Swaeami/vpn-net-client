//go:build windows

package admin

import (
	"syscall"
)

func IsAdmin() (bool, error) {
	isAdmin, err := isRunningAsAdmin()
	if err != nil {
		return false, err
	}

	if !isAdmin {
		return false, nil
	}
	return true, nil
}

func isRunningAsAdmin() (bool, error) {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	isUserAnAdmin := shell32.NewProc("IsUserAnAdmin")

	ret, _, _ := isUserAnAdmin.Call()
	return ret != 0, nil
}
