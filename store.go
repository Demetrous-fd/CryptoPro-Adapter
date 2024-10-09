package cades

type Store CadesObject

func NewStore(cades *Cades) (*Store, error) {
	body := &CadesRequestBody{
		Tabid: cades.Id,
		Data: &CadesRequestData{
			RequestId:   cades.RequestId,
			Destination: "nmcades",
			Method:      "CreateObject",
			Params: []CadesParam{
				{Type: "string", Value: "CAdESCOM.Store"},
			},
		},
	}

	_, err := cades.SendRequest(body)

	if err != nil {
		return &Store{}, err
	}

	cades.ObjId++
	store := Store{
		Cades: cades,
		ObjId: cades.ObjId,
	}
	return &store, nil
}

func (store *Store) Open(args ...any) error {
	params := ArgumentsToParams(3, args)
	err := CallVoidMethod((*CadesObject)(store), "Open", params)
	return err
}

func (store *Store) Close() error {
	params := []CadesParam{}
	err := CallVoidMethod((*CadesObject)(store), "Close", params)
	return err
}

func (store *Store) Add(obj *Certificate) error {
	param := ValueToParam(*(*CadesObject)(obj))
	err := CallVoidMethod((*CadesObject)(store), "Add", []CadesParam{*param})
	return err
}

func (store *Store) Certificates() (*Certificates, error) {
	obj, err := GetPropertyWithObject((*CadesObject)(store), "Certificates")
	if err != nil {
		return &Certificates{}, err
	}

	certificates := Certificates{
		Cades: obj.Cades,
		ObjId: obj.ObjId,
	}
	return &certificates, nil
}
