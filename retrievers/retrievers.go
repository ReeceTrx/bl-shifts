package retrievers

// Retriever interface defines the GetCodes method
type Retriever interface {
	GetCodes() ([]string, float64, string, error)
}
