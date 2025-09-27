package retrievers

// Retriever is an interface for fetching shift codes from any source
type Retriever interface {
	// GetCodes returns the latest codes, the post timestamp, and an error if any
	GetCodes() ([]string, float64, error)
}
