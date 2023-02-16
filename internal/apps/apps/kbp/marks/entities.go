package marks

type Mark struct {
	Action    string `json:"action"`
	Value     string `json:"value"`
	MarkID    string `json:"mark_id"`
	LessonID  string `json:"pair_id"`
	StudentID string `json:"student_id"`
}
