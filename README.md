# CryptoPro Adapter

Модуль для работы с certmgr и nmcades на Windows и Linux.

### Требования

- КриптоПро ЭЦП 4-5 версии
- КриптоПро ЭЦП Browser plug-in

### Запуск процессов
- `NewNMCadesProcess() (*CadesProcess, error)`
	```golang
	type CadesProcess struct {
		cmd    *exec.Cmd
		stdout *io.ReadCloser
		stdin  *io.WriteCloser
	}
	func PostMessage(file io.WriteCloser, message []byte) error
	func GetMessage(stdout io.ReadCloser) string
	```
- `NewCertManagerProcess(args ...string) (string, error)`


### CadesManager
```golang
type CadesManager struct{}

type Container struct {
	ContainerName       string `json:"containerName"`
	UniqueContainerName string `json:"uniqueContainerName"`
}

type InstallPfxResult struct {
	Container  Container `json:"container"`
	Thumbprint string    `json:"thumbprint"`
	Output     string    `json:"output"`
	OK         bool      `json:"ok"`
}

func (cm *CadesManager) GetContainer(partOfContainerName string) (*Container, error)
func (cm *CadesManager) GetListOfContainers() ([]Container, error)
func (cm *CadesManager) CopyContainer(container *Container, newLocation string) (*Container, error)
func (cm *CadesManager) LinkCertWithContainer(certPath, containerName string) (bool, error)
func (cm *CadesManager) DeleteCertificate(thumbprint string) (bool, error)
func (cm *CadesManager) DeleteContainer(container *Container) (bool, error) 
func (cm *CadesManager) IsCertificateExists(thumbprint string) (bool, error)
func (cm *CadesManager) RenameContainer(container *Container, newContainerName string) (*Container, error)
func (cm *CadesManager) InstallContainerFromFolder(containerFolderPath string, rootContainersFolderPath string, containerStorageName string, containerName string) (*Container, error)
func (cm *CadesManager) InstallPfx(path string, password string, exportable bool) (*InstallPfxResult, error)
func (cm *CadesManager) ExportContainerToPfx(filePath string, containerName string, password string) (string, error)
func (cm *CadesManager) InstallCertificate(filePath string, storeName string, autoDist bool) error
func (cm *CadesManager) GetCSPInfo() (string, error)
```


### Utils
- Методы для получения Отпечатока SHA1 из сертификата
  ```golang
  func GetCertificateThumbprintFromFile(path string) (string, error)
  func GetCertificateThumbprint(data []byte) (string, error)
  func GetThumbprint(certificate *x509.Certificate) string
  ```
- Метод для получения открытого ключа из сертификата
  ```golang
  func GetCertificatePublicKey(publicKeyInfo *SubjectPublicKeyInfoAsn1) string
  ```
- Методы для получения первых 8 байт открытого ключа из сертификата и контейнера
  ```golang
  func ParseSubjectPublicKeyInfo(cert *x509.Certificate) (*SubjectPublicKeyInfoAsn1, error)
  func GetCertificateShortPublicKey(publicKeyInfo *SubjectPublicKeyInfoAsn1) string
  func GetShortPublicKeyFromPrivateKey(headerData []byte) string
  ```
- Метод для создания нового имени ключевого носителя (name.key)
  ```golang
  func NewPrivateKeyName(cp1251Name string) []byte
  ```
- Методы для прямого переименования контейнера
  ```golang
  func DirectRenameContainerFolder(path string, newNameInCP1251 string) (bool, error)
  func DirectRenameContainerHDImage(username string, uniqueContainerName string, newNameInCP1251 string) (bool, error)

  func GetUserSid(username string) (string, error)
  func DirectRenameContainerRegistry(userSid string, containerName string, newNameInCP1251 string) (bool, error)
  ```
- Метод для парсинга сертификата
  ```golang
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

  func ParseGostCertificate(x509Certificate *x509.Certificate) (*GostCertificate, error)
  ```
- Методы для реализаций интерфейсов COM через NMCades
  ```golang
    type DefaultTypeValue[T any] struct {
        Value T
    }
    type ErrorCollector struct {
        err error
    }
  ```
  - `ValueToParam(value any) *CadesParam`
  - `ArgumentsToParams(maxArgs uint8, args []any) []CadesParam`
  - `GetProperty[T any](c *CadesObject, name string) (T, error)`
  - `GetPropertyWithObject(c *CadesObject, name string) (*CadesObject, error)`
  - `CallMethod(c *CadesObject, name string, params []CadesParam) (*CadesResponseData, error)`
  - `CallVoidMethod(c *CadesObject, name string, params []CadesParam) error`
  - `SafeExecute[T any](ec *ErrorCollector, f func() (T, error)) T`
  - `SafeExecuteWithObject[T any](w *ErrorCollector, f func() (*T, error)) *T`

### NMCades

```golang
type CallbackData struct {
	Id     uint32 `json:"callback_id"`
	Object string `json:"object"`
	Type   string `json:"type"`
	Value  string `json:"value"`
}

type CadesRequestData struct {
	CallbackId  uint32       `json:"callback_id,omitempty"`
	RequestId   uint32       `json:"requestid,omitempty"`
	ObjId       uint32       `json:"objid,omitempty"`
	Object      string       `json:"object,omitempty"`
	Destination string       `json:"destination,omitempty"`
	Type        string       `json:"type,omitempty"`
	Value       string       `json:"value,omitempty"`
	Method      string       `json:"method,omitempty"`
	Property    string       `json:"get_property,omitempty"`
	Url         string       `json:"url,omitempty"`
	Params      []CadesParam `json:"params,omitempty"`
}

type ReturnValue struct {
	Type       string      `json:"type"`
	Value      interface{} `json:"value"`
	Methods    []string    `json:"methods,omitempty"`
	Properties []string    `json:"properties,omitempty"`
}

type CadesResponseData struct {
	RequestId   uint32      `json:"requestid,omitempty"`
	ReturnValue ReturnValue `json:"retval,omitempty"`
	Type        string      `json:"type,omitempty"`
	Value       interface{} `json:"value,omitempty"`
	Message     string      `json:"message,omitempty"`
}

type CadesRequestBody struct {
	Tabid string            `json:"tabid"`
	Data  *CadesRequestData `json:"data"`
}

type CadesResponseBody struct {
	Tabid string           `json:"tabid"`
	Data  *json.RawMessage `json:"data"`
}

type CadesParam struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type CadesProcess struct {
	cmd    *exec.Cmd
	stdout *io.ReadCloser
	stdin  *io.WriteCloser
}

type Cades struct {
	Id        string
	RequestId uint32
	ObjId     uint32
	Process   *CadesProcess
}

type CadesObject struct {
	Cades *Cades
	ObjId uint32
}
```

- `NewCades() (*Cades, error)` Создание экземляра nmcades
  - `(cades *Cades) SendRequest(request *CadesRequestBody) (*CadesResponseData, error)`
  - `(cades *Cades) Close()`

#### Пример использования nmcades

```golang
package main

import (
	"log"
	cades "github.com/Demetrous-fd/CryptoPro-Adapter"
)

func getCertificate(thumbprint string) (*cades.Certificate, error) {
	defaultValue := cades.DefaultTypeValue[cades.Certificate]{}.Value

	cadesObj, err := cades.NewCades()
	if err != nil {
		log.Println("Fail to create a cades object", err)
		return &defaultValue, err
	}

	store, err := NewStore(cadesObj)
	if err != nil {
		log.Println("Fail to create a store object", err)
		return &defaultValue, err
	}

	err = store.Open(cades.CAPICOM_CURRENT_USER_STORE, cades.CAPICOM_MY_STORE)
	if err != nil {
		log.Println("Fail to open store", err)
		return &defaultValue, err
	}

	certificates, err := store.Certificates()
	if err != nil {
		log.Println("Fail to create a Certificates object", err)
		return &defaultValue, err
	}

	certs, err := certificates.Find(cades.CAPICOM_CERTIFICATE_FIND_SHA1_HASH, thumbprint)
	if err != nil {
		log.Println("Fail to create a Certificates object", err)
		return &defaultValue, err
	}

	count, err := certs.Count()
	if err != nil {
		log.Println("Fail to execute Certificates.Count method", err)
		return &defaultValue, err
	}

	if count <= 0 {
		return &defaultValue, cades.ErrEmpty
	}

	certificate, err := certificates.Item(1)
	if err != nil {
		log.Println("Fail to execute Certificates.Item method", err)
		return &defaultValue, err
	}

	return certificate, nil
}
```