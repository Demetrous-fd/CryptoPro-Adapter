package cades

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"os"
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
	} else if cObj, ok := value.(CadesObject); ok {
		paramType = "object"
		paramValue = cObj.ObjId
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
			GetProperty: name,
			// Property:    name,
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

func SetProperty(c *CadesObject, name string, params []CadesParam) (bool, error) {
	defaultValue := false
	body := &CadesRequestBody{
		Tabid: c.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   c.Cades.RequestId,
			ObjId:       c.ObjId,
			Destination: "nmcades",
			SetProperty: name,
			Params:      params,
		},
	}

	data, err := c.Cades.SendRequest(body)
	if err != nil {
		return defaultValue, err

	}

	result, ok := data.ReturnValue.Value.(string)
	if ok {
		return result == "OK", nil
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
			GetProperty: name,
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
	Error error
}

func SafeExecuteVoid(ec *ErrorCollector, f func() error) {
	if ec.Error != nil {
		return
	}

	err := f()
	if err != nil {
		ec.Error = err
		return
	}
}

func SafeExecute[T any](ec *ErrorCollector, f func() (T, error)) T {
	defaultValue := DefaultTypeValue[T]{}.Value
	if ec.Error != nil {
		return defaultValue
	}

	value, err := f()
	if err != nil {
		ec.Error = err
		return value
	}

	return value
}

func SafeExecuteWithObject[T any](w *ErrorCollector, f func() (*T, error)) *T {
	defaultValue := DefaultTypeValue[T]{}.Value
	if w.Error != nil {
		return &defaultValue
	}

	value, err := f()
	if err != nil {
		w.Error = err
		return value
	}

	return value
}

func LoadCertificate(data []byte) (*x509.Certificate, error) {
	certRaw := data
	block, _ := pem.Decode(certRaw)
	if block != nil {
		certRaw = block.Bytes
	}

	certificate, err := x509.ParseCertificate(certRaw)
	if err != nil {
		return certificate, err
	}
	return certificate, nil
}

func GetThumbprint(certificate *x509.Certificate) string {
	fingerprintRaw := sha1.Sum(certificate.Raw)
	thumbprint := hex.EncodeToString(fingerprintRaw[:])
	return thumbprint
}

func GetCertificateThumbprint(data []byte) (string, error) {
	cert, err := LoadCertificate(data)
	if err != nil {
		return "", err
	}

	thumbprint := GetThumbprint(cert)
	return thumbprint, nil
}

func GetCertificateThumbprintFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return GetCertificateThumbprint(data)
}

type AlgorithmInfoAsn1 struct {
	AlgorithmOID  asn1.ObjectIdentifier
	ParameterOIDs []asn1.ObjectIdentifier
}

type SubjectPublicKeyInfoAsn1 struct {
	AlgorithmInfo AlgorithmInfoAsn1
	PublicKey     asn1.BitString
}

func ParseSubjectPublicKeyInfo(cert *x509.Certificate) (*SubjectPublicKeyInfoAsn1, error) {
	var publicKeyInfo SubjectPublicKeyInfoAsn1

	_, err := asn1.Unmarshal(cert.RawSubjectPublicKeyInfo, &publicKeyInfo)
	return &publicKeyInfo, err
}

func GetCertificateShortPublicKey(publicKeyInfo *SubjectPublicKeyInfoAsn1) string {
	var shortPublicKey []byte

	publicKey := publicKeyInfo.PublicKey.Bytes
	switch publicKey[1] {
	case 0x40:
		shortPublicKey = publicKey[2 : 8+2]
	case 0x81:
		shortPublicKey = publicKey[3 : 8+3]
	default:
		shortPublicKey = publicKey[:8]
	}

	return hex.EncodeToString(shortPublicKey)
}

func GetShortPublicKeyFromPrivateKey(headerData []byte) string {
	var shortPublicKey []byte
	for i := 0; i < len(headerData)-1; i++ {
		if headerData[i] == 0x8a && headerData[i+1] == 0x08 {
			shortPublicKey = headerData[i+2 : i+10]
			break
		}
	}
	return hex.EncodeToString(shortPublicKey)
}
