package cades

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	var pathMgr string

	if runtime.GOOS == "linux" {
		pathMgr = fmt.Sprintf("/opt/cprocsp/bin/%s/nmcades", runtime.GOARCH)
	} else {
		pathMgr = "C:\\Program Files (x86)\\Crypto Pro\\CAdES Browser Plug-in\\nmcades.exe"
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

func NewCertManagerProcess(args ...string) (string, error) {
	var pathMgr string
	err_message := "Не удаётся найти файл Certmgr. Если вы установили КриптоПро не по умолчанию, создайте переменную среды CRYPTOPRO_FOLDER и укажите путь до папки КриптоПро."

	path, cryptopro_folder_set := os.LookupEnv("CRYPTOPRO_FOLDER")
	if runtime.GOOS == "linux" {
		if cryptopro_folder_set {
			pathMgr = fmt.Sprintf("%s/certmgr", path)
		} else if _, err := os.Stat("/opt/cprocsp/bin/amd64/certmgr"); err == nil {
			pathMgr = "/opt/cprocsp/bin/amd64/certmgr"
		} else if _, err := os.Stat("/opt/cprocsp/bin/386/certmgr"); err == nil {
			pathMgr = "/opt/cprocsp/bin/386/certmgr"
		} else {
			panic(err_message)
		}

	} else {
		if _, err := os.Stat("C:\\Program Files\\Crypto Pro\\CSP\\certmgr.exe"); err == nil {
			pathMgr = "C:\\Program Files\\Crypto Pro\\CSP\\certmgr.exe"
		} else if _, err := os.Stat("C:\\Program Files (x86)\\Crypto Pro\\CSP\\certmgr.exe"); err == nil {
			pathMgr = "C:\\Program Files (x86)\\Crypto Pro\\CSP\\certmgr.exe"
		} else if cryptopro_folder_set {
			pathMgr = fmt.Sprintf("%s\\CSP\\certmgr.exe", path)
		} else {
			panic(err_message)
		}
	}
	cmd := exec.Command(pathMgr, args...)

	stdoutStderr, err := cmd.CombinedOutput()
	slog.Debug(fmt.Sprintf("CertMgr start with args: %q", cmd.Args))

	d := charmap.CodePage866.NewDecoder()
	data, errDecode := d.Bytes(stdoutStderr)
	if errDecode != nil {
		return "", errDecode
	}

	return string(data), err
}
