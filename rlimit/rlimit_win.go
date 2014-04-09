// +build !linux,!darwin

// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package rlimit

func GetMaxOpenFiles() uint64 {
	return 10000
}
