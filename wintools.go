//go:build windows
// +build windows

package cades

// https://github.com/tejainece/winver/tree/master

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
	"golang.org/x/text/encoding/charmap"
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

func GetUserSid(username string) (string, error) {
	cmd := exec.Command("wmic", "useraccount", "where", fmt.Sprintf("name='%s'", username), "get", "sid")

	stdoutStderr, err := cmd.CombinedOutput()

	d := charmap.CodePage866.NewDecoder()
	data, errDecode := d.Bytes(stdoutStderr)
	if errDecode != nil {
		return "", errDecode
	}

	output := string(data)
	output = strings.Split(output, "\r\n")[1]
	output = strings.ReplaceAll(output, "\r", "")
	output = strings.TrimSpace(output)
	return output, err
}

func renameRegistryKey(oldPath string, newPath string) error {
	oldKey, err := registry.OpenKey(registry.LOCAL_MACHINE, oldPath, registry.READ|registry.WRITE)
	if err != nil {
		return fmt.Errorf("failed to open old key: %w", err)
	}
	defer oldKey.Close()

	newKey, _, err := registry.CreateKey(registry.LOCAL_MACHINE, newPath, registry.WRITE)
	if err != nil {
		return fmt.Errorf("failed to create new key: %w", err)
	}
	defer newKey.Close()

	values, err := oldKey.ReadValueNames(0)
	if err != nil {
		return fmt.Errorf("failed to read values from old key: %w", err)
	}

	for _, value := range values {
		val, _, err := oldKey.GetBinaryValue(value)
		if err != nil {
			return fmt.Errorf("failed to get value %s: %w", value, err)
		}
		err = newKey.SetBinaryValue(value, val)
		if err != nil {
			return fmt.Errorf("failed to set value %s: %w", value, err)
		}
	}

	err = registry.DeleteKey(registry.LOCAL_MACHINE, oldPath)
	if err != nil {
		return fmt.Errorf("failed to delete old key: %w", err)
	}

	return nil
}

func getCryptoProUsersRegistryKey() (string, error) {
	keyPath := `SOFTWARE\WOW6432Node\Crypto Pro\Settings\Users`
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.READ|registry.WRITE)
	if err == nil {
		defer key.Close()
		return keyPath, nil
	}

	keyPath = `SOFTWARE\CryptoPro\Settings\Users`
	key, err = registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.READ|registry.WRITE)
	if err == nil {
		defer key.Close()
		return keyPath, nil
	}
	return "", err
}

// containerName - имя контейнера, без пути \\.\(REGISTRY|HDIMAGE|FAT12) и т.д.
func DirectRenameContainerRegistry(userSid string, containerName string, newNameInCP1251 string) (bool, error) {
	keysPath, err := getCryptoProUsersRegistryKey()
	if err != nil {
		return false, err
	}

	keyPath := fmt.Sprintf(`%s\%s\keys\%s`, keysPath, userSid, containerName)
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.READ|registry.WRITE)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, fmt.Errorf("registry key does not exist: %s", keyPath)
		} else {
			return false, fmt.Errorf("error opening registry key: %s", err)
		}
	}
	defer key.Close()

	decoder := charmap.Windows1251.NewDecoder()
	newNameUtf, err := decoder.String(newNameInCP1251)
	if err != nil {
		return false, fmt.Errorf("cant decode cp1251 string: %s", err)
	}

	nameBin := NewPrivateKeyName(newNameInCP1251)
	oldNameBin, _, err := key.GetBinaryValue("name.key")
	if err != nil {
		return false, fmt.Errorf("cant save old data from %s\\name.key: %s", keyPath, err)
	}

	err = key.SetBinaryValue("name.key", nameBin)
	if err != nil {
		return false, fmt.Errorf("cant change %s\\name.key: %s", keyPath, err)
	}

	newKeyPath := fmt.Sprintf(`%s\%s\keys\%s`, keysPath, userSid, newNameUtf)
	err = renameRegistryKey(keyPath, newKeyPath)
	if err != nil {
		err = key.SetBinaryValue("name.key", oldNameBin)
		if err != nil {
			return false, fmt.Errorf("cant change %s\\name.key: %s", keyPath, err)
		}
	}
	return true, nil
}
