package retrievers

// Retriever interface now returns codes, post timestamp, and error
type Retriever interface {
	GetCodes() ([]string, float64, error)
}
