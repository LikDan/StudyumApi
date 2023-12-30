package entities

type CellResponse struct {
	Cell       JournalCell    `json:"cell"`
	Average    float32        `json:"average"`
	MarkAmount map[string]int `json:"markAmount"`
	RowColor   string         `json:"rowColor"`
}
