// +build darwin
// +build !cgo

package cpu

import "github.com/yonzay/tlsfiber/internal/gopsutil/common"

func perCPUTimes() ([]TimesStat, error) {
	return []TimesStat{}, common.ErrNotImplementedError
}

func allCPUTimes() ([]TimesStat, error) {
	return []TimesStat{}, common.ErrNotImplementedError
}
