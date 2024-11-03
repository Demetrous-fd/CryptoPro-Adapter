//go:build windows
// +build windows

package cades

// https://github.com/tejainece/winver/tree/master

import (
	"syscall"
)

// Version contains data about a Windows version
type WinVersion struct {
	Major uint8
	Minor uint8
	Build uint16
}

func (v WinVersion) isWindows7OrLower() bool {
	return v.Major <= 6 && v.Minor <= 1
}

func getWindowsVersion() WinVersion {
	v, _ := syscall.GetVersion()
	ver := WinVersion{
		Major: byte(v),
		Minor: uint8(v >> 8),
		Build: uint16(v >> 16),
	}
	return ver
}
