package cades

import (
	"bytes"
	"crypto/sha1"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"
)

func DateToUTCStr(d time.Time) string {
	return d.UTC().Format("2006-01-02T15:04:05.999Z")
}

func ValueToParam(value any) *CadesParam {
	var paramType string
	paramValue := value
	if _, ok := value.(string); ok {
		paramType = "string"
	} else if d, ok := value.(time.Time); ok {
		paramType = "string"
		paramValue = DateToUTCStr(d)
	} else if _, ok := value.(bool); ok {
		paramType = "boolean"
	} else {
		paramType = "number"
	}
	return &CadesParam{Type: paramType, Value: paramValue}
}

func ArgumentsToParams(maxArgs uint8, args []any) []CadesParam {
	params := make([]CadesParam, maxArgs)
	var lastIndex uint8 = 0
	for _, value := range args {
		if lastIndex >= maxArgs {
			break
		}
		param := ValueToParam(value)
		params[lastIndex] = *param
		lastIndex++
	}
	return params[:lastIndex]
}

type DefaultTypeValue[T any] struct {
	Value T
}

func GetProperty[T any](c *CadesObject, name string) (T, error) {
	defaultValue := DefaultTypeValue[T]{}.Value
	body := &CadesRequestBody{
		Tabid: c.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   c.Cades.RequestId,
			ObjId:       c.ObjId,
			Destination: "nmcades",
			Property:    name,
		},
	}

	data, err := c.Cades.SendRequest(body)
	if err != nil {
		return defaultValue, err

	}

	value, ok := data.ReturnValue.Value.(T)
	if ok {
		return value, nil
	}

	return defaultValue, ErrEmpty
}

func GetPropertyWithObject(c *CadesObject, name string) (*CadesObject, error) {
	defaultValue := DefaultTypeValue[CadesObject]{}.Value
	body := &CadesRequestBody{
		Tabid: c.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   c.Cades.RequestId,
			ObjId:       c.ObjId,
			Destination: "nmcades",
			Property:    name,
		},
	}

	data, err := c.Cades.SendRequest(body)
	if err != nil {
		return &defaultValue, err
	}

	if data.ReturnValue.Type == "object" {
		c.Cades.ObjId++
		defaultValue.Cades = c.Cades
		defaultValue.ObjId = c.Cades.ObjId
		return &defaultValue, nil
	}

	return &defaultValue, ErrEmpty
}

func CallMethod(c *CadesObject, name string, params []CadesParam) (*CadesResponseData, error) {
	body := &CadesRequestBody{
		Tabid: c.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   c.Cades.RequestId,
			ObjId:       c.ObjId,
			Destination: "nmcades",
			Method:      name,
			Params:      params,
		},
	}

	data, err := c.Cades.SendRequest(body)
	if err != nil {
		return data, err
	}

	return data, nil
}

func CallVoidMethod(c *CadesObject, name string, params []CadesParam) error {
	data, err := CallMethod(c, name, params)
	if err != nil {
		return err
	}

	if data.ReturnValue.Value != "OK" {
		return ErrMethodExecution
	}

	return nil
}

type ErrorCollector struct {
	err error
}

func SafeExecute[T any](ec *ErrorCollector, f func() (T, error)) T {
	defaultValue := DefaultTypeValue[T]{}.Value
	if ec.err != nil {
		return defaultValue
	}

	value, err := f()
	if err != nil {
		ec.err = err
		return value
	}

	return value
}

func SafeExecuteWithObject[T any](w *ErrorCollector, f func() (*T, error)) *T {
	defaultValue := DefaultTypeValue[T]{}.Value
	if w.err != nil {
		return &defaultValue
	}

	value, err := f()
	if err != nil {
		w.err = err
		return value
	}

	return value
}

func GetCertificateThumbprintFromFile(path string) (string, error) {
	// read file content
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// pass cert bytes
	cert, err := x509.ParseCertificate(data)
	if err != nil {
		return "", err
	}

	// generate fingerprint with sha1
	// you can also use md5, sha256, etc.
	fingerprint := sha1.Sum(cert.Raw)

	var buf bytes.Buffer
	for _, f := range fingerprint {
		fmt.Fprintf(&buf, "%02X", f)
	}
	thumbprint := strings.ToLower(buf.String())
	return thumbprint, nil
}
