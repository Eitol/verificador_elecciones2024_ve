package results

type FindByDocIDResponse struct {
	Result    *Result
	IsCached  bool
	IsUpdated bool
}

type ResultRepo interface {
	FindByDocID(docID int) (*FindByDocIDResponse, error)
}

type ResultStore interface {
	StoreResult(result Result) error
	GetResultByURL(url string) (*Result, error)
}
