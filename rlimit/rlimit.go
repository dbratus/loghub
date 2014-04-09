// +build linux darwin

package rlimit

import "syscall"

func GetMaxOpenFiles() uint64 {
	var lim syscall.Rlimit
	syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim)
	return (lim.Cur / 4) * 3
}
