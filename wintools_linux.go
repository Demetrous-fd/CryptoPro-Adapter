//go:build linux
// +build linux

package cades

import "fmt"

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

func GetUserSid(username string) (string, error) {
	return "", fmt.Errorf("not implemented for this os")
}

func renameRegistryKey(oldPath string, newPath string) error {
	return fmt.Errorf("not implemented for this os")
}

func DirectRenameContainerRegistry(userSid string, containerName string, newNameInCP1251 string) (bool, error) {
	return false, fmt.Errorf("not implemented for this os")
}
