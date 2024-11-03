package cades

type About CadesObject

func NewAbout(cades *Cades) (*About, error) {
	body := &CadesRequestBody{
		Tabid: cades.Id,
		Data: &CadesRequestData{
			RequestId:   cades.RequestId,
			Destination: "nmcades",
			Method:      "CreateObject",
			Params: []CadesParam{
				{Type: "string", Value: "CAdESCOM.About"},
			},
		},
	}

	_, err := cades.SendRequest(body)

	if err != nil {
		return &About{}, err
	}

	cades.ObjId++
	About := About{
		Cades: cades,
		ObjId: cades.ObjId,
	}
	return &About, nil
}

func (about *About) MajorVersion() (int, error) {
	value, err := GetProperty[float64]((*CadesObject)(about), "MajorVersion")
	return int(value), err
}

func (about *About) MinorVersion() (int, error) {
	value, err := GetProperty[float64]((*CadesObject)(about), "MinorVersion")
	return int(value), err
}

func (about *About) BuildVersion() (int, error) {
	value, err := GetProperty[float64]((*CadesObject)(about), "BuildVersion")
	return int(value), err
}

func (about *About) Version() (string, error) {
	return GetProperty[string]((*CadesObject)(about), "Version")
}

func (about *About) PluginVersion() (*Version, error) {
	obj, err := GetPropertyWithObject((*CadesObject)(about), "PluginVersion")

	if err != nil {
		return &Version{}, err
	}

	return (*Version)(obj), err
}

type Version CadesObject

func (version *Version) MajorVersion() (int, error) {
	value, err := GetProperty[float64]((*CadesObject)(version), "MajorVersion")
	return int(value), err
}

func (version *Version) MinorVersion() (int, error) {
	value, err := GetProperty[float64]((*CadesObject)(version), "MinorVersion")
	return int(value), err
}

func (version *Version) BuildVersion() (int, error) {
	value, err := GetProperty[float64]((*CadesObject)(version), "BuildVersion")
	return int(value), err
}

func (version *Version) ToString() (string, error) {
	data, err := CallMethod((*CadesObject)(version), "toString", []CadesParam{})
	if err != nil {
		return "", err
	}

	if value, ok := data.ReturnValue.Value.(string); ok {
		return value, nil
	}

	return "", ErrEmpty
}
