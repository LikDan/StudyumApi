package schedule

type Lesson struct {
	Action      string `json:"action"`
	Date        string `json:"new_date"`
	Description string `json:"pair_disc"`
	PairType    int    `json:"pair_type"`
	SubjectID   string `json:"subject_id"`
	GroupID     string `json:"group_id"`
	ID          string `json:"reset_id"`
}
