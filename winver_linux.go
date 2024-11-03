//go:build linux
// +build linux

package cades

// Version contains data about a Windows version
type WinVersion struct {
	Major uint8
	Minor uint8
	Build uint16
}

func (v WinVersion) isWindows7OrLower() bool {
	return false
}

func getWindowsVersion() WinVersion {
	ver := WinVersion{}
	return ver
}
