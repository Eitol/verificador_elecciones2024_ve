package results

type Result struct {
	Url          string `json:"url"`
	Bytes        []byte `json:"bytes"`
	ExampleDocID string `json:"exampleDocID"`
}
