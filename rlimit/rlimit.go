// +build linux,darwin

// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package rlimit

import "syscall"

func GetMaxOpenFiles() uint64 {
	var lim syscall.Rlimit
	syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim)
	return (lim.Cur / 4) * 3
}
