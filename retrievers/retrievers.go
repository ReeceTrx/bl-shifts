package retrievers

type Retriever interface {
	GetCodes() ([]string, error)
}
