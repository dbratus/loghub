// +build !linux,!darwin

package rlimit

func GetMaxOpenFiles() uint64 {
	return 10000
}
