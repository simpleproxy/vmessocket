package securedload

var knownProtectedLoader map[string]ProtectedLoader

type ProtectedLoader interface {
	VerifyAndLoad(filename string) ([]byte, error)
}

func RegisterProtectedLoader(name string, sv ProtectedLoader) {
	if knownProtectedLoader == nil {
		knownProtectedLoader = map[string]ProtectedLoader{}
	}
	knownProtectedLoader[name] = sv
}
