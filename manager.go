package cades

import (
	"fmt"
	"log/slog"
	"regexp"
	"runtime"
	"strings"
)

var SHA1_PATTERN = regexp.MustCompile(`(?m)^SHA1 .*$`)
var CONTAINER_PATTERN = regexp.MustCompile(`(?m)^(?:Container|Контейнер) .*$`)

type CadesManager struct{}

type InstallPfxResult struct {
	Container  string `json:"container"`
	Thumbprint string `json:"thumbprint"`
	Output     string `json:"output"`
	OK         bool   `json:"ok"`
}

func (cm *CadesManager) InstallPfx(path string, password string) (*InstallPfxResult, error) {
	var (
		output string
		err    error
		args   []string = []string{"-inst", "-pfx", "-file", path, "-pin", password}
	)
	res := InstallPfxResult{}

	if runtime.GOOS == "windows" {
		ver := getWindowsVersion()
		isWin7OrLower := ver.isWindows7OrLower()
		if !isWin7OrLower {
			args = append(args, "-silent")
		}
	} else {
		args = append(args, "-silent")
	}

	output, err = NewCertManagerProcess(args...)
	res.Output = output
	if err != nil {
		slog.Error(fmt.Sprintf("Fail to start certmgr: %s", err))
		return &res, err
	}

	if strings.Contains(output, "SHA1") {
		thumbprints := SHA1_PATTERN.FindAllString(output, -1)
		thumbprintsCount := len(thumbprints)
		if thumbprintsCount >= 1 {
			thumbprint := thumbprints[thumbprintsCount-1]
			thumbprint = strings.Split(thumbprint, ": ")[1]
			res.Thumbprint = thumbprint[0:40]
			res.OK = true
		}
	}

	if strings.Contains(output, "Container") || strings.Contains(output, "Контейнер") {
		container := CONTAINER_PATTERN.FindString(output)
		container = strings.Split(container, ": ")[1]
		container = strings.TrimSpace(container)
		res.Container = container
	}
	return &res, err
}

type LinkCertResult struct {
	Output string `json:"output"`
	OK     bool   `json:"ok"`
}

func (cm *CadesManager) LinkCertWithContainer(certPath, containerName string) (*LinkCertResult, error) {
	result := &LinkCertResult{}
	output, err := NewCertManagerProcess("-inst", "-inst_to_cont", "-file", certPath, "-cont", containerName, "-silent")

	result.Output = output
	result.OK = strings.Contains(output, "[ErrorCode: 0x00000000]")

	if err != nil {
		slog.Error(fmt.Sprintf("LinkCertWithContainer Error: %s", err))
		return result, err
	}

	return result, nil
}

func (cm *CadesManager) DeleteCertificate(thumbprint string) (string, error) {
	thumbprint = strings.ToLower(thumbprint)
	output, err := NewCertManagerProcess("-delete", "-certificate", "-thumbprint", thumbprint)
	if err != nil {
		slog.Error(fmt.Sprintf("Fail to delete certificate with thumbprint %s: %s", thumbprint, err))
		slog.Error(fmt.Sprintf("Output log: %s", output))
	}
	return output, err
}

func (cm *CadesManager) DeleteContainer(containerName string) (string, error) {
	output, err := NewCertManagerProcess("-delete", "-container", containerName)
	if err != nil {
		slog.Error(fmt.Sprintf("Fail to delete container with uname %s: %s", containerName, err))
		slog.Error(fmt.Sprintf("Output log: %s", output))
	}
	return output, err
}

func (cm *CadesManager) IsCertificateExists(thumbprint string) (bool, error) {
	output, err := NewCertManagerProcess("-list", "-thumbprint", thumbprint)
	if err != nil {
		slog.Debug(fmt.Sprintf("Fail to find certificate by thumbprint: %s", thumbprint))
		slog.Debug(fmt.Sprintf("Output log: %s", output))
		return false, err
	}

	result := strings.Contains(output, "[ErrorCode: 0x00000000]")
	return result, nil
}
