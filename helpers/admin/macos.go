//go:build darwin

package admin

import (
	"os"
)

func IsAdmin() (bool, error) {
	if os.Geteuid() != 0 {
		return false, nil
	}
	return true, nil
}
