package cades

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"os"
	"time"
)

var GostAlgorithmNames = map[string]string{
	"1.2.643.2.2.19":    "ГОСТ Р 34.11-94 256 бит",
	"1.2.643.7.1.1.1.1": "ГОСТ Р 34.11-2012 256 бит",
	"1.2.643.7.1.1.1.2": "ГОСТ Р 34.11-2012 512 бит",
}

var SubjectAndIssuerNames = map[string]string{
	"2.5.4.3":                    "common_name",
	"2.5.4.4":                    "surname",
	"2.5.4.5":                    "serial_number",
	"2.5.4.6":                    "country_name",
	"2.5.4.7":                    "locality_name",
	"2.5.4.8":                    "state_or_province_name",
	"2.5.4.9":                    "street_address",
	"2.5.4.10":                   "organization_name",
	"2.5.4.11":                   "organizational_unit_name",
	"2.5.4.12":                   "title",
	"2.5.4.15":                   "business_category",
	"2.5.4.17":                   "postal_code",
	"2.5.4.20":                   "telephone_number",
	"2.5.4.41":                   "name",
	"2.5.4.42":                   "given_name",
	"2.5.4.43":                   "initials",
	"2.5.4.44":                   "generation_qualifier",
	"2.5.4.45":                   "unique_identifier",
	"2.5.4.46":                   "dn_qualifier",
	"2.5.4.65":                   "pseudonym",
	"2.5.4.97":                   "organization_identifier",
	"2.23.133.2.1":               "tpm_manufacturer",
	"2.23.133.2.2":               "tpm_model",
	"2.23.133.2.3":               "tpm_version",
	"2.23.133.2.4":               "platform_manufacturer",
	"2.23.133.2.5":               "platform_model",
	"2.23.133.2.6":               "platform_version",
	"1.2.840.113549.1.9.1":       "email_address",
	"1.3.6.1.4.1.311.60.2.1.1":   "incorporation_locality",
	"1.3.6.1.4.1.311.60.2.1.2":   "incorporation_state_or_province",
	"1.3.6.1.4.1.311.60.2.1.3":   "incorporation_country",
	"0.9.2342.19200300.100.1.1":  "user_id",
	"0.9.2342.19200300.100.1.25": "domain_component",
	"0.2.262.1.10.7.20":          "name_distinguisher",
	"1.2.643.100.1":              "ogrn",
	"1.2.643.100.3":              "snils",
	"1.2.643.100.4":              "innle",
	"1.2.643.100.5":              "ogrnip",
	"1.2.643.3.131.1.1":          "inn",
}

type AlgorithmInfo struct {
	OID        string   `json:"oid"`
	Name       string   `json:"name"`
	Parameters []string `json:"parameters"`
}

type GostCertificate struct {
	Issuer         map[string]string `json:"issuer"`
	Subject        map[string]string `json:"subject"`
	SerialNumber   string            `json:"serial_number"`
	Thumbprint     string            `json:"thumbprint"`
	PublicKey      string            `json:"public_key"`
	ShortPublicKey string            `json:"short_public_key"`
	Algorithm      AlgorithmInfo     `json:"algorithm"`
	NotAfter       time.Time         `json:"not_after"`
	NotBefore      time.Time         `json:"not_before"`
}

type AlgorithmInfoAsn1 struct {
	AlgorithmOID  asn1.ObjectIdentifier
	ParameterOIDs []asn1.ObjectIdentifier
}

type SubjectPublicKeyInfoAsn1 struct {
	AlgorithmInfo AlgorithmInfoAsn1
	PublicKey     asn1.BitString
}

type DefaultTypeValue[T any] struct {
	Value T
}

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

func ParseSubjectPublicKeyInfo(cert *x509.Certificate) (*SubjectPublicKeyInfoAsn1, error) {
	var publicKeyInfo SubjectPublicKeyInfoAsn1

	_, err := asn1.Unmarshal(cert.RawSubjectPublicKeyInfo, &publicKeyInfo)
	return &publicKeyInfo, err
}

func GetCertificatePublicKey(publicKeyInfo *SubjectPublicKeyInfoAsn1) string {
	var publicKey []byte

	publicKeyRaw := publicKeyInfo.PublicKey.Bytes
	switch publicKeyRaw[1] {
	case 0x40:
		publicKey = publicKeyRaw[2:]
	case 0x81:
		publicKey = publicKeyRaw[3:]
	default:
		publicKey = publicKeyRaw[:]
	}

	return hex.EncodeToString(publicKey)
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

func NewPrivateKeyName(cp1251Name string) []byte {
	if len(cp1251Name) > 125 {
		cp1251Name = cp1251Name[:125]
	}

	iam5Block := make([]byte, 0)

	iam5Block = append(iam5Block, 0x16, byte(len(cp1251Name)))
	iam5Block = append(iam5Block, []byte(cp1251Name)...)

	result := []byte{0x30, byte(len(iam5Block))}
	result = append(result, iam5Block...)
	return result
}

func ParseGostCertificate(x509Certificate *x509.Certificate) (*GostCertificate, error) {
	certificate := GostCertificate{
		Issuer:  make(map[string]string),
		Subject: make(map[string]string),
	}

	if x509Certificate.PublicKey != nil {
		return &certificate, errors.New("is not gost certificate")
	}

	for _, v := range x509Certificate.Subject.Names {
		name, ok := SubjectAndIssuerNames[v.Type.String()]
		if ok {
			certificate.Subject[name] = v.Value.(string)
		} else {
			certificate.Subject[v.Type.String()] = v.Value.(string)
		}
	}

	for _, v := range x509Certificate.Issuer.Names {
		name, ok := SubjectAndIssuerNames[v.Type.String()]
		if ok {
			certificate.Issuer[name] = v.Value.(string)
		} else {
			certificate.Issuer[v.Type.String()] = v.Value.(string)
		}
	}

	subjectPublicKeyInfo, err := ParseSubjectPublicKeyInfo(x509Certificate)
	if err == nil {
		certificate.Algorithm.OID = subjectPublicKeyInfo.AlgorithmInfo.AlgorithmOID.String()
		certificate.Algorithm.Name = GostAlgorithmNames[certificate.Algorithm.OID]

		for _, v := range subjectPublicKeyInfo.AlgorithmInfo.ParameterOIDs {
			certificate.Algorithm.Parameters = append(certificate.Algorithm.Parameters, v.String())
		}

		certificate.PublicKey = GetCertificatePublicKey(subjectPublicKeyInfo)
		shortPublicKey := GetCertificateShortPublicKey(subjectPublicKeyInfo)
		certificate.ShortPublicKey = shortPublicKey
	}

	certificate.NotAfter = x509Certificate.NotAfter
	certificate.NotBefore = x509Certificate.NotBefore
	certificate.SerialNumber = x509Certificate.SerialNumber.String()

	fingerprintRaw := sha1.Sum(x509Certificate.Raw)
	fingerprint := hex.EncodeToString(fingerprintRaw[:])
	certificate.Thumbprint = fingerprint
	return &certificate, nil
}
