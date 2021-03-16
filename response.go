package paperboy

// SummariesResponse ...
type SummariesResponse struct {
	LastID    string `json:"LastId"`
	Summaries []*Summary
}
