package cades

type X509EnrollmentRoot CadesObject

func CreateX509EnrollmentRoot(cades *Cades) *X509EnrollmentRoot {
	return &X509EnrollmentRoot{
		Cades: cades,
		ObjId: 0,
	}
}

type CspAlgorithms CadesObject

func (alg *CspAlgorithms) Count() (int, error) {
	value, err := GetProperty[float64]((*CadesObject)(alg), "Count")
	return int(value), err
}
func (alg *CspAlgorithms) ItemByIndex(index int) (*CspAlgorithm, error) {
	param := ValueToParam(index)
	_, err := CallMethod((*CadesObject)(alg), "ItemByIndex", []CadesParam{*param})
	if err != nil {
		return &CspAlgorithm{}, err
	}

	alg.Cades.ObjId++
	obj := &CspAlgorithm{
		Cades: alg.Cades,
		ObjId: alg.Cades.ObjId,
	}
	return obj, nil
}

type CCspInformation CadesObject

func (info *CCspInformation) Name() (string, error) {
	return GetProperty[string]((*CadesObject)(info), "Name")
}
func (info *CCspInformation) Type() (int, error) {
	value, err := GetProperty[float64]((*CadesObject)(info), "Type")
	return int(value), err
}
func (info *CCspInformation) LegacyCsp() (bool, error) {
	return GetProperty[bool]((*CadesObject)(info), "LegacyCsp")
}
func (info *CCspInformation) CspAlgorithms() (*CspAlgorithms, error) {
	obj, err := GetPropertyWithObject((*CadesObject)(info), "CspAlgorithms")

	if err != nil {
		return &CspAlgorithms{}, err
	}

	return (*CspAlgorithms)(obj), err
}

type CCspInformations CadesObject

func (info *CCspInformations) AddAvailableCsps() error {
	return CallVoidMethod((*CadesObject)(info), "AddAvailableCsps", []CadesParam{})
}
func (info *CCspInformations) Count() (int, error) {
	value, err := GetProperty[float64]((*CadesObject)(info), "Count")
	return int(value), err
}
func (info *CCspInformations) ItemByIndex(index int) (*CCspInformation, error) {
	param := ValueToParam(index)
	_, err := CallMethod((*CadesObject)(info), "ItemByIndex", []CadesParam{*param})
	if err != nil {
		return &CCspInformation{}, err
	}

	info.Cades.ObjId++
	infoObj := CCspInformation{
		Cades: info.Cades,
		ObjId: info.Cades.ObjId,
	}
	return &infoObj, nil
}
func (info *CCspInformations) ItemByName(providerName string) (*CCspInformation, error) {
	param := ValueToParam(providerName)
	_, err := CallMethod((*CadesObject)(info), "ItemByName", []CadesParam{*param})
	if err != nil {
		return &CCspInformation{}, err
	}

	info.Cades.ObjId++
	infoObj := CCspInformation{
		Cades: info.Cades,
		ObjId: info.Cades.ObjId,
	}
	return &infoObj, nil
}

func (info *CCspInformations) GetCspStatusFromProviderName(name string, keySpecFlag int) (*CCSPStatus, error) {
	params := ArgumentsToParams(2, []any{name, keySpecFlag})
	_, err := CallMethod((*CadesObject)(info), "GetCspStatusFromProviderName", params)
	if err != nil {
		return &CCSPStatus{}, err
	}

	info.Cades.ObjId++
	obj := CCSPStatus{
		Cades: info.Cades,
		ObjId: info.Cades.ObjId,
	}
	return &obj, nil
}

type CCSPStatus CadesObject

func (status *CCSPStatus) CspAlgorithm() (*CspAlgorithm, error) {
	obj, err := GetPropertyWithObject((*CadesObject)(status), "CspAlgorithm")

	if err != nil {
		return &CspAlgorithm{}, err
	}

	return (*CspAlgorithm)(obj), err
}

func (status *CCSPStatus) CspInformation() (*CCspInformation, error) {
	obj, err := GetPropertyWithObject((*CadesObject)(status), "CspInformation")

	if err != nil {
		return &CCspInformation{}, err
	}

	return (*CCspInformation)(obj), err
}

type CspAlgorithm CadesObject

func (alg *CspAlgorithm) DefaultLength() (int, error) {
	value, err := GetProperty[float64]((*CadesObject)(alg), "DefaultLength")
	return int(value), err
}

func (alg *CspAlgorithm) Type() (int, error) {
	value, err := GetProperty[float64]((*CadesObject)(alg), "Type")
	return int(value), err
}

func (alg *CspAlgorithm) GetAlgorithmOid(long int, algFlags int) (*CObjectId, error) {
	params := ArgumentsToParams(2, []any{long, algFlags})
	_, err := CallMethod((*CadesObject)(alg), "GetAlgorithmOid", params)
	if err != nil {
		return &CObjectId{}, err
	}

	alg.Cades.ObjId++
	obj := CObjectId{
		Cades: alg.Cades,
		ObjId: alg.Cades.ObjId,
	}
	return &obj, nil
}

func (x509 *X509EnrollmentRoot) CCspInformations() (*CCspInformations, error) {
	body := &CadesRequestBody{
		Tabid: x509.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   x509.Cades.RequestId,
			Destination: "nmcades",
			Method:      "CreateObject",
			Params: []CadesParam{
				{Type: "string", Value: "X509Enrollment.CCspInformations"},
			},
		},
	}

	_, err := x509.Cades.SendRequest(body)

	if err != nil {
		return &CCspInformations{}, err
	}

	x509.Cades.ObjId++
	informations := CCspInformations{
		Cades: x509.Cades,
		ObjId: x509.Cades.ObjId,
	}
	return &informations, nil
}

type X509Enrollment CadesObject

func (en *X509Enrollment) InitializeFromRequest(obj *CX509CertificateRequestPkcs10) error {
	param := ValueToParam(*(*CadesObject)(obj))
	return CallVoidMethod((*CadesObject)(en), "InitializeFromRequest", []CadesParam{*param})
}

func (en *X509Enrollment) InstallResponse(args ...any) error {
	params := ArgumentsToParams(4, args)
	return CallVoidMethod((*CadesObject)(en), "InstallResponse", params)
}

func (en *X509Enrollment) Initialize(context int) error {
	param := ValueToParam(context)
	return CallVoidMethod((*CadesObject)(en), "Initialize", []CadesParam{*param})
}

func (en *X509Enrollment) CreateRequest(arg int) (string, error) {
	param := ValueToParam(arg)
	data, err := CallMethod((*CadesObject)(en), "CreateRequest", []CadesParam{*param})
	if err != nil {
		return "", err
	}

	if csr, ok := data.ReturnValue.Value.(string); ok {
		return csr, nil
	}

	return "", ErrEmpty
}

func (x509 *X509EnrollmentRoot) CX509Enrollment() (*X509Enrollment, error) {
	body := &CadesRequestBody{
		Tabid: x509.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   x509.Cades.RequestId,
			Destination: "nmcades",
			Method:      "CreateObject",
			Params: []CadesParam{
				{Type: "string", Value: "X509Enrollment.CX509Enrollment"},
			},
		},
	}

	_, err := x509.Cades.SendRequest(body)

	if err != nil {
		return &X509Enrollment{}, err
	}

	x509.Cades.ObjId++
	enrollment := X509Enrollment{
		Cades: x509.Cades,
		ObjId: x509.Cades.ObjId,
	}
	return &enrollment, nil
}

type CX509Extension CadesObject

func (ext *CX509Extension) Initialize(data string, args ...any) error {
	params := ArgumentsToParams(3, args)
	params = append(params, CadesParam{
		Type:  "string",
		Value: data,
	})
	return CallVoidMethod((*CadesObject)(ext), "Initialize", params)
}

func (x509 *X509EnrollmentRoot) CX509Extension() (*CX509Extension, error) {
	body := &CadesRequestBody{
		Tabid: x509.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   x509.Cades.RequestId,
			Destination: "nmcades",
			Method:      "CreateObject",
			Params: []CadesParam{
				{Type: "string", Value: "X509Enrollment.CX509Extension"},
			},
		},
	}

	_, err := x509.Cades.SendRequest(body)

	if err != nil {
		return &CX509Extension{}, err
	}

	x509.Cades.ObjId++
	ext := CX509Extension{
		Cades: x509.Cades,
		ObjId: x509.Cades.ObjId,
	}
	return &ext, nil
}

type X509Extensions CadesObject

func (ext *X509Extensions) Add(obj *CX509Extension) error {
	param := ValueToParam(*(*CadesObject)(obj))
	return CallVoidMethod((*CadesObject)(ext), "Add", []CadesParam{*param})
}

type CX509CertificateRequestPkcs10 CadesObject

func (pkcs10 *CX509CertificateRequestPkcs10) SetSubject(value *CX500DistinguishedName) (bool, error) {
	param := ValueToParam(*(*CadesObject)(value))
	return SetProperty((*CadesObject)(pkcs10), "Subject", []CadesParam{*param})
}

func (pkcs10 *CX509CertificateRequestPkcs10) SetHashAlgorithm(value *CObjectId) (bool, error) {
	param := ValueToParam(*(*CadesObject)(value))
	return SetProperty((*CadesObject)(pkcs10), "HashAlgorithm", []CadesParam{*param})
}

func (pkcs10 *CX509CertificateRequestPkcs10) X509Extensions() (*X509Extensions, error) {
	obj, err := GetPropertyWithObject((*CadesObject)(pkcs10), "X509Extensions")

	if err != nil {
		return &X509Extensions{}, err
	}

	return (*X509Extensions)(obj), err
}
func (pkcs10 *CX509CertificateRequestPkcs10) InitializeFromPrivateKey(args ...any) error {
	params := ArgumentsToParams(3, args)
	return CallVoidMethod((*CadesObject)(pkcs10), "InitializeFromPrivateKey", params)
}

func (x509 *X509EnrollmentRoot) CX509CertificateRequestPkcs10() (*CX509CertificateRequestPkcs10, error) {
	body := &CadesRequestBody{
		Tabid: x509.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   x509.Cades.RequestId,
			Destination: "nmcades",
			Method:      "CreateObject",
			Params: []CadesParam{
				{Type: "string", Value: "X509Enrollment.CX509CertificateRequestPkcs10"},
			},
		},
	}

	_, err := x509.Cades.SendRequest(body)

	if err != nil {
		return &CX509CertificateRequestPkcs10{}, err
	}

	x509.Cades.ObjId++
	pkcs10 := CX509CertificateRequestPkcs10{
		Cades: x509.Cades,
		ObjId: x509.Cades.ObjId,
	}
	return &pkcs10, nil
}

type CX509PrivateKey CadesObject

func (pk *CX509PrivateKey) SetKeySpec(value int) (bool, error) {
	param := ValueToParam(value)
	return SetProperty((*CadesObject)(pk), "KeySpec", []CadesParam{*param})
}
func (pk *CX509PrivateKey) SetProviderName(value string) (bool, error) {
	param := ValueToParam(value)
	return SetProperty((*CadesObject)(pk), "ProviderName", []CadesParam{*param})
}
func (pk *CX509PrivateKey) SetPin(value string) (bool, error) {
	param := ValueToParam(value)
	return SetProperty((*CadesObject)(pk), "Pin", []CadesParam{*param})
}
func (pk *CX509PrivateKey) SetProviderType(value int) (bool, error) {
	param := ValueToParam(value)
	return SetProperty((*CadesObject)(pk), "ProviderType", []CadesParam{*param})
}
func (pk *CX509PrivateKey) SetKeyProtection(value int) (bool, error) {
	param := ValueToParam(value)
	return SetProperty((*CadesObject)(pk), "KeyProtection", []CadesParam{*param})
}
func (pk *CX509PrivateKey) SetLength(value int) (bool, error) {
	param := ValueToParam(value)
	return SetProperty((*CadesObject)(pk), "Length", []CadesParam{*param})
}
func (pk *CX509PrivateKey) SetExportPolicy(value int) (bool, error) {
	param := ValueToParam(value)
	return SetProperty((*CadesObject)(pk), "ExportPolicy", []CadesParam{*param})
}
func (pk *CX509PrivateKey) SetExisting(value bool) (bool, error) {
	param := ValueToParam(value)
	return SetProperty((*CadesObject)(pk), "Existing", []CadesParam{*param})
}
func (pk *CX509PrivateKey) SetMachineContext(value bool) (bool, error) {
	param := ValueToParam(value)
	return SetProperty((*CadesObject)(pk), "MachineContext", []CadesParam{*param})
}
func (pk *CX509PrivateKey) SetContainerName(value string) (bool, error) {
	param := ValueToParam(value)
	return SetProperty((*CadesObject)(pk), "ContainerName", []CadesParam{*param})
}

func (x509 *X509EnrollmentRoot) CX509PrivateKey() (*CX509PrivateKey, error) {
	body := &CadesRequestBody{
		Tabid: x509.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   x509.Cades.RequestId,
			Destination: "nmcades",
			Method:      "CreateObject",
			Params: []CadesParam{
				{Type: "string", Value: "X509Enrollment.CX509PrivateKey"},
			},
		},
	}

	_, err := x509.Cades.SendRequest(body)

	if err != nil {
		return &CX509PrivateKey{}, err
	}

	x509.Cades.ObjId++
	pk := CX509PrivateKey{
		Cades: x509.Cades,
		ObjId: x509.Cades.ObjId,
	}
	return &pk, nil
}

type CX509ExtensionKeyUsage CadesObject

func (keyUsage *CX509ExtensionKeyUsage) InitializeEncode(flags int) error {
	param := ValueToParam(flags)
	return CallVoidMethod((*CadesObject)(keyUsage), "InitializeEncode", []CadesParam{*param})
}

func (x509 *X509EnrollmentRoot) CX509ExtensionKeyUsage() (*CX509ExtensionKeyUsage, error) {
	body := &CadesRequestBody{
		Tabid: x509.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   x509.Cades.RequestId,
			Destination: "nmcades",
			Method:      "CreateObject",
			Params: []CadesParam{
				{Type: "string", Value: "X509Enrollment.CX509ExtensionKeyUsage"},
			},
		},
	}

	_, err := x509.Cades.SendRequest(body)

	if err != nil {
		return &CX509ExtensionKeyUsage{}, err
	}

	x509.Cades.ObjId++
	keyUsage := CX509ExtensionKeyUsage{
		Cades: x509.Cades,
		ObjId: x509.Cades.ObjId,
	}
	return &keyUsage, nil
}

type CX509ExtensionEnhancedKeyUsage CadesObject

func (eKey *CX509ExtensionEnhancedKeyUsage) InitializeEncode(obj *CObjectIds) error {
	param := ValueToParam(*(*CadesObject)(obj))
	return CallVoidMethod((*CadesObject)(eKey), "InitializeEncode", []CadesParam{*param})
}

func (x509 *X509EnrollmentRoot) CX509ExtensionEnhancedKeyUsage() (*CX509ExtensionEnhancedKeyUsage, error) {
	body := &CadesRequestBody{
		Tabid: x509.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   x509.Cades.RequestId,
			Destination: "nmcades",
			Method:      "CreateObject",
			Params: []CadesParam{
				{Type: "string", Value: "X509Enrollment.CX509ExtensionEnhancedKeyUsage"},
			},
		},
	}

	_, err := x509.Cades.SendRequest(body)

	if err != nil {
		return &CX509ExtensionEnhancedKeyUsage{}, err
	}

	x509.Cades.ObjId++
	enhancedKeyUsage := CX509ExtensionEnhancedKeyUsage{
		Cades: x509.Cades,
		ObjId: x509.Cades.ObjId,
	}
	return &enhancedKeyUsage, nil
}

type CObjectId CadesObject

func (id *CObjectId) InitializeFromValue(oid string) error {
	param := ValueToParam(oid)
	return CallVoidMethod((*CadesObject)(id), "InitializeFromValue", []CadesParam{*param})
}

func (id *CObjectId) Value() (string, error) {
	return GetProperty[string]((*CadesObject)(id), "Value")
}

func (id *CObjectId) FriendlyName() (string, error) {
	return GetProperty[string]((*CadesObject)(id), "FriendlyName")
}

func (x509 *X509EnrollmentRoot) CObjectId() (*CObjectId, error) {
	body := &CadesRequestBody{
		Tabid: x509.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   x509.Cades.RequestId,
			Destination: "nmcades",
			Method:      "CreateObject",
			Params: []CadesParam{
				{Type: "string", Value: "X509Enrollment.CObjectId"},
			},
		},
	}

	_, err := x509.Cades.SendRequest(body)

	if err != nil {
		return &CObjectId{}, err
	}

	x509.Cades.ObjId++
	id := CObjectId{
		Cades: x509.Cades,
		ObjId: x509.Cades.ObjId,
	}
	return &id, nil
}

type CObjectIds CadesObject

func (ids *CObjectIds) Add(obj *CObjectId) error {
	param := ValueToParam(*(*CadesObject)(obj))
	return CallVoidMethod((*CadesObject)(ids), "Add", []CadesParam{*param})
}

func (x509 *X509EnrollmentRoot) CObjectIds() (*CObjectIds, error) {
	body := &CadesRequestBody{
		Tabid: x509.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   x509.Cades.RequestId,
			Destination: "nmcades",
			Method:      "CreateObject",
			Params: []CadesParam{
				{Type: "string", Value: "X509Enrollment.CObjectIds"},
			},
		},
	}

	_, err := x509.Cades.SendRequest(body)

	if err != nil {
		return &CObjectIds{}, err
	}

	x509.Cades.ObjId++
	ids := CObjectIds{
		Cades: x509.Cades,
		ObjId: x509.Cades.ObjId,
	}
	return &ids, nil
}

type CX500DistinguishedName CadesObject

func (dn *CX500DistinguishedName) Encode(args ...any) error {
	params := ArgumentsToParams(2, args)
	return CallVoidMethod((*CadesObject)(dn), "Encode", params)
}

func (x509 *X509EnrollmentRoot) CX500DistinguishedName() (*CX500DistinguishedName, error) {
	body := &CadesRequestBody{
		Tabid: x509.Cades.Id,
		Data: &CadesRequestData{
			RequestId:   x509.Cades.RequestId,
			Destination: "nmcades",
			Method:      "CreateObject",
			Params: []CadesParam{
				{Type: "string", Value: "X509Enrollment.CX500DistinguishedName"},
			},
		},
	}

	_, err := x509.Cades.SendRequest(body)

	if err != nil {
		return &CX500DistinguishedName{}, err
	}

	x509.Cades.ObjId++
	dn := CX500DistinguishedName{
		Cades: x509.Cades,
		ObjId: x509.Cades.ObjId,
	}
	return &dn, nil
}
