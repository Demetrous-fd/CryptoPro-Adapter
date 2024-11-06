package cades

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	cp "github.com/otiai10/copy"
	"golang.org/x/exp/slog"
)

var SHA1_PATTERN = regexp.MustCompile(`(?m)^SHA1 .*$`)
var CONTAINER_PATTERN = regexp.MustCompile(`(?m)^(?:Container|Контейнер) .*$`)
var CONTAINER_NAMES_PATTERN = regexp.MustCompile(`(?m)^\\\\\.\\.*\|\\\\\.\\.*$`)

type CadesManager struct{}

type InstallPfxResult struct {
	Container  Container `json:"container"`
	Thumbprint string    `json:"thumbprint"`
	Output     string    `json:"output"`
	OK         bool      `json:"ok"`
}

func (cm *CadesManager) InstallPfx(path string, password string, exportable bool) (*InstallPfxResult, error) {
	var (
		output string
		err    error
		args   []string = []string{"-inst", "-pfx", "-file", path, "-silent"}
	)
	if password != "" {
		args = append(args, "-pin", password)
	}
	if exportable {
		args = append(args, "-keep_exportable")
	}
	res := InstallPfxResult{}

	output, err = NewCertManagerProcess(args...)
	res.Output = output
	if err != nil {
		slog.Debug(fmt.Sprintf("Fail to start certmgr: %s", err))
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
		containerName := CONTAINER_PATTERN.FindString(output)
		containerName = strings.Split(containerName, ": ")[1]
		containerName = strings.TrimSpace(containerName)

		container, err := cm.GetContainer(containerName)
		if err != nil {
			res.Container = Container{
				ContainerName:       containerName,
				UniqueContainerName: containerName,
			}
		} else {
			res.Container = *container
		}
	}
	return &res, err
}

func (cm *CadesManager) LinkCertWithContainer(certPath, containerName string) (bool, error) {
	var result bool
	output, err := NewCertManagerProcess("-inst", "-inst_to_cont", "-file", certPath, "-cont", containerName, "-silent")
	result = strings.Contains(output, "[ErrorCode: 0x00000000]")

	if err != nil {
		slog.Debug(fmt.Sprintf("LinkCertWithContainer Error: %s", err))
		return result, err
	}

	return result, nil
}

func (cm *CadesManager) DeleteCertificate(thumbprint string) (bool, error) {
	thumbprint = strings.ToLower(thumbprint)
	output, err := NewCertManagerProcess("-delete", "-certificate", "-thumbprint", thumbprint)
	if err != nil {
		slog.Debug(fmt.Sprintf("Fail to delete certificate with thumbprint %s: %s", thumbprint, err))
		slog.Debug(fmt.Sprintf("Output log: %s", output))
		return false, err
	}

	if strings.Contains(output, "[ErrorCode: 0x00000000]") {
		slog.Debug(fmt.Sprintf("Certificate with thumbprint: %s deleted", thumbprint))
		return true, nil
	}
	return false, nil

}

func (cm *CadesManager) DeleteContainer(container *Container) (bool, error) {
	output, err := NewCertManagerProcess("-delete", "-container", container.UniqueContainerName)
	if err != nil {
		slog.Debug(fmt.Sprintf("Fail to delete container with uname %s: %s", container.UniqueContainerName, err))
		slog.Debug(fmt.Sprintf("Output log: %s", output))
		return false, err
	}

	if strings.Contains(output, "[ErrorCode: 0x00000000]") {
		slog.Debug(fmt.Sprintf("Container with name: %s deleted", container.ContainerName))
		return true, nil
	}
	return false, nil
}

func (cm *CadesManager) IsCertificateExists(thumbprint string, store string) (bool, error) {
	args := []string{"-list", "-thumbprint", thumbprint}
	if store != "" {
		args = append(args, "-store", store)
	}

	output, err := NewCertManagerProcess(args...)
	if err != nil {
		slog.Debug(fmt.Sprintf("Fail to find certificate by thumbprint: %s", thumbprint))
		slog.Debug(fmt.Sprintf("Output log: %s", output))
		return false, err
	}

	result := strings.Contains(output, "[ErrorCode: 0x00000000]")
	return result, nil
}

type Container struct {
	ContainerName       string `json:"containerName"`
	UniqueContainerName string `json:"uniqueContainerName"`
}

func (cm *CadesManager) GetListOfContainers() ([]Container, error) {
	result := []Container{}
	output, err := NewCSPTestProcess("-keyset", "-enum_cont", "-verifycontext", "-fqcn", "-un")
	if err != nil {
		slog.Debug("Fail to get list of containers")
		slog.Debug(fmt.Sprintf("Output log: %s", output))
		return result, err
	}

	containers := CONTAINER_NAMES_PATTERN.FindAllString(output, -1)
	for _, containerRaw := range containers {
		containerNames := strings.Split(containerRaw, "|")
		if len(containerNames) < 2 {
			continue
		}
		container := Container{
			ContainerName:       strings.TrimSpace(containerNames[0]),
			UniqueContainerName: strings.TrimSpace(containerNames[1]),
		}
		result = append(result, container)
	}

	return result, nil
}

func (cm *CadesManager) GetContainer(partOfContainerName string) (*Container, error) {
	result := &Container{}
	output, err := NewCSPTestProcess("-keyset", "-enum_cont", "-verifycontext", "-fqcn", "-un")
	if err != nil {
		slog.Debug("Fail to get list of containers")
		slog.Debug(fmt.Sprintf("Output log: %s", output))
		return result, err
	}

	escaped_pattern := fmt.Sprintf(`(?m)^.*%s.*$`, regexp.QuoteMeta(partOfContainerName))
	pattern, err := regexp.Compile(escaped_pattern)
	if err != nil {
		return result, err
	}

	containerRaw := pattern.FindString(output)
	containerNames := strings.Split(containerRaw, "|")
	if len(containerNames) < 2 {
		return result, fmt.Errorf("Container: %s not found", partOfContainerName)
	}

	result = &Container{
		ContainerName:       strings.TrimSpace(containerNames[0]),
		UniqueContainerName: strings.TrimSpace(containerNames[1]),
	}

	return result, nil
}

func (cm *CadesManager) CopyContainer(container *Container, newLocation string) (*Container, error) {
	result := &Container{}
	var containerExists bool
	output, err := NewCSPTestProcess("-keycopy", "-contsrc", container.UniqueContainerName, "-contdest", newLocation, "-silent")
	if err != nil {
		if strings.Contains(output, "ErrorCode: 0x8009000b") {
			slog.Debug(fmt.Sprintf("Output log: %s", output))
			return result, ErrContainerNotExportable
		} else if !strings.Contains(output, "ErrorCode: 0x8009000f") {
			slog.Debug(fmt.Sprintf("Output log: %s", output))
			return result, err
		}

		containerExists = true
		slog.Debug(fmt.Sprintf("Container[%s] exists", newLocation))
	}

	result, err = cm.GetContainer(newLocation)
	if err != nil {
		return result, err
	}

	if containerExists {
		return result, ErrContainerExists
	}

	return result, nil
}

func (cm *CadesManager) InstallContainerFromFolder(containerFolderPath string, rootContainersFolderPath string, containerStorageName string, containerName string) (*Container, error) {
	containerFolderName := filepath.Base(containerFolderPath)
	newContainerPath := filepath.Join(rootContainersFolderPath, containerFolderName)

	if _, err := os.Stat(newContainerPath); err != nil {
		cp.Copy(containerFolderPath, newContainerPath)
	}

	container, err := cm.GetContainer(containerFolderName)
	if err != nil {
		slog.Debug(err.Error())
		return container, err
	}

	if containerStorageName == "" {
		if runtime.GOOS == "windows" {
			containerStorageName = "REGISTRY"
		} else {
			containerStorageName = "HDIMAGE"
		}
	}

	if containerName == "" {
		containerNameRaw := strings.Split(container.ContainerName, `\`)
		containerName = containerNameRaw[len(containerNameRaw)-1]
	}
	location := fmt.Sprintf(`\\.\%s\%s mi`, containerStorageName, containerName)

	newContainer, err := cm.CopyContainer(container, location)
	if err != nil {
		slog.Debug(fmt.Sprintf("InstallContainerFromFolder: Cant a copy container %s -> %s; error: %s", container.ContainerName, location, err.Error()))
	}

	if runtime.GOOS == "linux" {
		err := os.RemoveAll(newContainerPath)
		if err != nil {
			slog.Debug(fmt.Sprintf("InstallContainerFromFolder: Cant remove a copy of the container folder: %s; error: %s", newContainerPath, err.Error()))
			return newContainer, nil
		}
	}
	return newContainer, err
}

func (cm *CadesManager) RenameContainer(container *Container, newContainerName string) (*Container, error) {
	result := &Container{}

	containerStorageName := strings.ReplaceAll(container.ContainerName, `\\.\`, "")
	containerStorageName = strings.Split(containerStorageName, `\`)[0]

	location := fmt.Sprintf(`\\.\%s\%s`, containerStorageName, newContainerName)
	if location == container.ContainerName || location == container.UniqueContainerName {
		slog.Debug(fmt.Sprintf("RenameContainer: The new container name matches the old one. OLD[%s] = NEW[%s]", container.ContainerName, location))
		return container, nil
	}

	result, err := cm.CopyContainer(container, location)
	if err != nil {
		if errors.Is(err, ErrContainerExists) {
			return container, nil
		}

		slog.Debug(fmt.Sprintf("RenameContainer: Cant copy container: %s to %s", container.ContainerName, location))
		return result, err
	}

	if result.ContainerName == container.ContainerName || result.UniqueContainerName == container.UniqueContainerName {
		slog.Debug(fmt.Sprintf("RenameContainer: The new container name matches the old one. OLD[%s] = NEW[%s]", container.ContainerName, location))
		return container, nil
	}

	_, err = cm.DeleteContainer(container)
	if err != nil {
		slog.Debug(fmt.Sprintf("RenameContainer: Cant delete old container: %s", container.ContainerName))
		return result, nil
	}
	return result, nil
}

func (cm *CadesManager) ExportContainerToPfx(filePath string, containerName string, password string) (string, error) {
	var args []string = []string{"-export", "-container", containerName, "-pfx", "-dest", filePath}
	if password != "" {
		args = append(args, "-pin", password)
	}

	output, err := NewCertManagerProcess(args...)
	if err != nil {
		slog.Debug(fmt.Sprintf("Fail to export container[%s] to pfx[%s], error: %s", containerName, filePath, err))
		slog.Debug(fmt.Sprintf("Certmgr log: %s", output))
		if strings.Contains(output, "ErrorCode: 0x8009000b") {
			return "", ErrContainerNotExportable
		}
		return "", err
	}

	return filePath, nil
}

func (cm *CadesManager) InstallCertificate(filePath string, storeName string, autoDist bool) error {
	var args []string = []string{"-inst", "-file", filePath}
	if storeName != "" {
		args = append(args, "-store", storeName)
	}

	if autoDist {
		args = append(args, "-autodist")
	}

	output, err := NewCertManagerProcess(args...)
	if err != nil {
		if storeName == "" {
			slog.Debug(fmt.Sprintf("Fail to install certificate[%s] to default store, error: %s", filePath, err))
		} else {
			slog.Debug(fmt.Sprintf("Fail to install certificate[%s] to store[%s], error: %s", filePath, storeName, err))
		}
		slog.Debug(fmt.Sprintf("Certmgr log: %s", output))
		return err
	}

	return nil
}
