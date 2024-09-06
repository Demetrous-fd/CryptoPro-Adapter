package cades

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"unsafe"

	"golang.org/x/exp/slog"
	"golang.org/x/text/encoding/charmap"
)

var nativeEndian binary.ByteOrder

type CadesProcess struct {
	Cmd    *exec.Cmd
	Stdout *io.ReadCloser
	Stdin  *io.WriteCloser
}

func DetermineByteOrder() {
	// determine native byte order so that we can read message size correctly
	var one int16 = 1
	b := (*byte)(unsafe.Pointer(&one))
	if *b == 0 {
		nativeEndian = binary.BigEndian
	} else {
		nativeEndian = binary.LittleEndian
	}
}

func WriteHeader(writer io.Writer, length int) error {
	header := make([]byte, 4)
	nativeEndian.PutUint32(header, (uint32)(length))

	if n, err := writer.Write(header); err != nil || n != len(header) {
		return err
	}

	return nil
}

func PostMessage(file io.WriteCloser, message []byte) error {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	length := len(message)

	if err := WriteHeader(writer, length); err != nil {
		return err
	}

	if n, err := writer.Write(message); err != nil || n != length {
		return err
	}

	writer.Flush()
	file.Write(buffer.Bytes())

	return nil
}

func ReadHeader(stdout io.ReadCloser) (uint32, error) {
	length := make([]byte, 4)

	_, err := stdout.Read(length)
	if err != nil {
		return nativeEndian.Uint32(length), err
	}

	return nativeEndian.Uint32(length), nil
}

func GetMessageAsBytes(stdout io.ReadCloser) []byte {
	length, err := ReadHeader(stdout)
	if err != nil || length == 0 {
		return []byte{}
	}

	data := make([]byte, length)

	_, err = stdout.Read(data)
	if err != nil {
		return []byte{}
	}

	return data
}

func GetMessage(stdout io.ReadCloser) string {
	data := GetMessageAsBytes(stdout)
	return string(data)
}

func NewNMCadesProcess() (*CadesProcess, error) {
	if nativeEndian == nil {
		DetermineByteOrder()
	}

	pathMgr, err := getCryptoProUtilPath("nmcades")
	if err != nil {
		return &CadesProcess{}, err
	}
	cmd := exec.Command(pathMgr)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error(fmt.Sprintf("Fail connect to StdOut: %s", err))
		return &CadesProcess{}, err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		slog.Error(fmt.Sprintf("Fail connect to StdIn: %s", err))
		return &CadesProcess{}, err
	}

	err = cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Fail start nmcades: %s", err))
		return &CadesProcess{}, err
	}

	return &CadesProcess{Cmd: cmd, Stdout: &stdout, Stdin: &stdin}, nil
}

func getCryptoProUtilPath(filename string) (string, error) {
	errMessage := fmt.Sprintf("Не удаётся найти файл %s. Если вы установили КриптоПро не по умолчанию, создайте переменную среды CRYPTOPRO_FOLDER и укажите путь до папки КриптоПро.", filename)
	folders := []string{}
	var result string
	var err error

	path, cryptoProFolderSet := os.LookupEnv("CRYPTOPRO_FOLDER")
	if cryptoProFolderSet {
		folders = append(folders, path)
	}

	if runtime.GOOS == "linux" {
		folders = append(
			folders,
			"/opt/cprocsp/bin/amd64",
			"/opt/cprocsp/bin/386",
			fmt.Sprintf("/opt/cprocsp/bin/%s", runtime.GOARCH),
		)
	} else {
		filename = fmt.Sprintf("%s.exe", filename)
		folders = append(
			folders,
			"C:\\Program Files\\Crypto Pro\\CSP",
			"C:\\Program Files (x86)\\Crypto Pro\\CSP",
			"C:\\Program Files (x86)\\Crypto Pro\\CAdES Browser Plug-in",
		)
	}

	for _, folder := range folders {
		path = filepath.Join(folder, filename)
		if _, err := os.Stat(path); err == nil {
			result = path
			break
		}
	}

	if result == "" {
		err = errors.New(errMessage)
	}
	return result, err
}

func removeSilentArg(args []string) []string {
	newArgs := []string{}
	for _, arg := range args {
		if arg != "-silent" {
			newArgs = append(newArgs, arg)
			continue
		}
	}
	return newArgs
}

func NewCertManagerProcess(args ...string) (string, error) {
	pathMgr, err := getCryptoProUtilPath("certmgr")
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		ver := getWindowsVersion()
		isWin7OrLower := ver.isWindows7OrLower()
		if isWin7OrLower {
			args = removeSilentArg(args)
		}
	}
	cmd := exec.Command(pathMgr, args...)

	stdoutStderr, err := cmd.CombinedOutput()
	slog.Debug(fmt.Sprintf("CertMgr start with args: %q", cmd.Args))

	if runtime.GOOS == "windows" {
		d := charmap.CodePage866.NewDecoder()
		data, errDecode := d.Bytes(stdoutStderr)
		if errDecode != nil {
			return "", errDecode
		}
		stdoutStderr = data
	}

	return string(stdoutStderr), err
}

func NewCSPTestProcess(args ...string) (string, error) {
	pathMgr, err := getCryptoProUtilPath("csptest")
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		ver := getWindowsVersion()
		isWin7OrLower := ver.isWindows7OrLower()
		if isWin7OrLower {
			args = removeSilentArg(args)
		}
	}
	cmd := exec.Command(pathMgr, args...)

	stdoutStderr, err := cmd.CombinedOutput()
	slog.Debug(fmt.Sprintf("csptest start with args: %q", cmd.Args))

	if runtime.GOOS == "windows" {
		d := charmap.CodePage866.NewDecoder()
		data, errDecode := d.Bytes(stdoutStderr)
		if errDecode != nil {
			return "", errDecode
		}
		stdoutStderr = data
	}

	return string(stdoutStderr), err
}
