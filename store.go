package cades

type Store CadesObject

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
