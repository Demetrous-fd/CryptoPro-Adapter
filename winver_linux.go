//go:build linux
// +build linux

package cades

// Version contains data about a Windows version
type Version struct {
	Major uint8
	Minor uint8
	Build uint16
}

func (v Version) isWindows7OrLower() bool {
	return false
}

func getWindowsVersion() Version {
	ver := Version{}
	return ver
}
