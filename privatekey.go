package cades

type PrivateKey CadesObject

func (pk *PrivateKey) ProviderName() (string, error) {
	return GetProperty[string]((*CadesObject)(pk), "ProviderName")
}
func (pk *PrivateKey) ProviderType() (int32, error) {
	value, err := GetProperty[float64]((*CadesObject)(pk), "ProviderType")
	if err != nil {
		return 0, err
	}
	return int32(value), nil
}
func (pk *PrivateKey) ContainerName() (string, error) {
	return GetProperty[string]((*CadesObject)(pk), "ContainerName")
}
func (pk *PrivateKey) UniqueContainerName() (string, error) {
	return GetProperty[string]((*CadesObject)(pk), "UniqueContainerName")
}
