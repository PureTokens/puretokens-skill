//go:build darwin || linux

package main

import (
	"os"
	"syscall"
)

func singleLinkFile(_ string, info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat.Nlink == 1
}

func clientPathSafe(_ string, info os.FileInfo) bool {
	return info.Mode()&os.ModeSymlink == 0
}
