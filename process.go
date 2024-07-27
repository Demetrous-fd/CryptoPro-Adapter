package cades

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"runtime"
	"unsafe"

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

func GetMessage(stdout io.ReadCloser) string {
	length, err := ReadHeader(stdout)
	if err != nil || length == 0 {
		return ""
	}

	data := make([]byte, length)

	_, err = stdout.Read(data)
	if err != nil {
		return ""
	}

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

	if runtime.GOOS == "linux" {
		pathMgr = fmt.Sprintf("/opt/cprocsp/bin/%s/certmgr", runtime.GOARCH)
	} else {
		pathMgr = "C:\\Program Files\\Crypto Pro\\CSP\\certmgr.exe"
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
