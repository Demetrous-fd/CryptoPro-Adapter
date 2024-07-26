package cades

type Certificates struct {
	Cades *Cades
	ObjId uint32
}

func (certificates *Certificates) Count() (uint16, error) {
	value, err := GetProperty[float64]((*CadesObject)(certificates), "Count")
	if err != nil {
		return 0, err
	}
	return uint16(value), nil
}

func (certificates *Certificates) Item(index uint16) (*Certificate, error) {
	param := ValueToParam(index)
	_, err := CallMethod((*CadesObject)(certificates), "Item", []CadesParam{*param})
	if err != nil {
		return &Certificate{}, err
	}

	certificates.Cades.ObjId++
	certificate := Certificate{
		Cades: certificates.Cades,
		ObjId: certificates.Cades.ObjId,
	}
	return &certificate, nil
}

// Arguments: (FindType, varCriteria, bFindValidOnly)
// https://learn.microsoft.com/en-us/windows/win32/seccrypto/certificates-find
func (certificates *Certificates) Find(args ...any) (*Certificates, error) {
	params := ArgumentsToParams(3, args)
	_, err := CallMethod((*CadesObject)(certificates), "Find", params)
	if err != nil {
		return &Certificates{}, err
	}

	certificates.Cades.ObjId++
	newCertificates := Certificates{
		Cades: certificates.Cades,
		ObjId: certificates.Cades.ObjId,
	}
	return &newCertificates, nil
}
