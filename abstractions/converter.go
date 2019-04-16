package abstractions

// Converter defines a generic blob converter
type Converter interface {
	Apply(blob []byte, spec *OutputSpec) ([]byte, error)
	Destroy()
}
