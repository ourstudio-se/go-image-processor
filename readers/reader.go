package readers

// Reader is a generic interface for supporting different types
// of blob readers
type Reader interface {
	ReadBlob() ([]byte, error)
}
