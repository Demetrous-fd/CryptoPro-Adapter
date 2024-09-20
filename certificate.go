package cades

import (
	"encoding/json"
	"log"
	"time"
)

type Certificate CadesObject

type ValidExport struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Expired bool   `json:"expired"`
}

type PrivateKeyExport struct {
	ProviderName        string `json:"providerName"`
	ContainerName       string `json:"containerName"`
	UniqueContainerName string `json:"uniqueContainerName"`
}

type CertificateExport struct {
	Valid         ValidExport      `json:"valid"`
	Issuer        string           `json:"issuer"`
	Subject       string           `json:"subject"`
	Thumbprint    string           `json:"thumbprint"`
	SerialNumber  string           `json:"serialNumber"`
	HasPrivateKey bool             `json:"hasPrivateKey"`
	PrivateKey    PrivateKeyExport `json:"privateKey,omitempty"`
}

func (certificate *Certificate) HasPrivateKey() (bool, error) {
	data, err := CallMethod((*CadesObject)(certificate), "HasPrivateKey", []CadesParam{})
	if err != nil {
		return false, err
	}

	if exists, ok := data.ReturnValue.Value.(bool); ok {
		return exists, nil
	}

	return false, ErrEmpty
}

func (certificate *Certificate) PrivateKey() (*PrivateKey, error) {
	object, err := GetPropertyWithObject((*CadesObject)(certificate), "PrivateKey")
	if err != nil {
		return &PrivateKey{}, err
	}

	pk := PrivateKey{
		Cades: object.Cades,
		ObjId: object.ObjId,
	}
	return &pk, nil
}

func (certificate *Certificate) Thumbprint() (string, error) {
	return GetProperty[string]((*CadesObject)(certificate), "Thumbprint")
}

func (certificate *Certificate) Version() (int32, error) {
	value, err := GetProperty[float64]((*CadesObject)(certificate), "Version")
	if err != nil {
		return 0, err
	}
	return int32(value), nil
}

func (certificate *Certificate) SubjectName() (string, error) {
	return GetProperty[string]((*CadesObject)(certificate), "SubjectName")
}

func (certificate *Certificate) IssuerName() (string, error) {
	return GetProperty[string]((*CadesObject)(certificate), "IssuerName")
}

func (certificate *Certificate) SerialNumber() (string, error) {
	return GetProperty[string]((*CadesObject)(certificate), "SerialNumber")
}

func (certificate *Certificate) ValidFromDate() (time.Time, error) {
	value, err := GetProperty[string]((*CadesObject)(certificate), "ValidFromDate")
	if err != nil {
		return time.Time{}, err
	}

	t, err := time.Parse("2006-01-02T15:04:05.999Z", value)
	return t, err
}

func (certificate *Certificate) ValidToDate() (time.Time, error) {
	value, err := GetProperty[string]((*CadesObject)(certificate), "ValidToDate")
	if err != nil {
		return time.Time{}, err
	}

	t, err := time.Parse("2006-01-02T15:04:05.999Z", value)
	return t, err
}

func (certificate *Certificate) IsExpire() (bool, error) {
	date, err := certificate.ValidToDate()
	now := time.Now()

	if err != nil {
		return false, err
	}

	return now.After(date), err
}

func (certificate *Certificate) ToExport() (*CertificateExport, error) {
	ec := &ErrorCollector{}
	validFrom := SafeExecute[time.Time](ec, certificate.ValidFromDate)
	validTo := SafeExecute[time.Time](ec, certificate.ValidToDate)
	expire := SafeExecute[bool](ec, certificate.IsExpire)

	issuer := SafeExecute[string](ec, certificate.IssuerName)
	subject := SafeExecute[string](ec, certificate.SubjectName)
	thumbprint := SafeExecute[string](ec, certificate.Thumbprint)
	serialNumber := SafeExecute[string](ec, certificate.SerialNumber)
	hasKey := SafeExecute[bool](ec, certificate.HasPrivateKey)

	export := CertificateExport{
		Valid: ValidExport{
			From:    validFrom.String(),
			To:      validTo.String(),
			Expired: expire,
		},
		Issuer:        issuer,
		Subject:       subject,
		Thumbprint:    thumbprint,
		SerialNumber:  serialNumber,
		HasPrivateKey: hasKey,
	}

	key := SafeExecuteWithObject[PrivateKey](ec, certificate.PrivateKey)
	emptyKey := &PrivateKey{}

	if key != emptyKey {
		providerName := SafeExecute[string](ec, key.ProviderName)
		containerName := SafeExecute[string](ec, key.ContainerName)
		uniqueContainerName := SafeExecute[string](ec, key.UniqueContainerName)

		export.PrivateKey = PrivateKeyExport{
			ProviderName:        providerName,
			ContainerName:       containerName,
			UniqueContainerName: uniqueContainerName,
		}
	}

	if ec.Error != nil {
		log.Println("Certificate.ToExport error: ", ec.Error)
		return &export, ec.Error
	}
	return &export, nil
}

func (certificate *Certificate) ToJson() ([]byte, error) {
	export, err := certificate.ToExport()
	if err != nil {
		log.Println("Certificate.ToJson error: ", err)
		return []byte{}, err
	}

	exportJson, err := json.Marshal(export)
	if err != nil {
		log.Println("Certificate.ToJson error: ", err)
		return []byte{}, err
	}

	return exportJson, nil
}
