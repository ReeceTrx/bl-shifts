type Retriever interface {
	GetCodes() ([]string, float64, error)
}
