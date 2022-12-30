package entities

type GeneratedTable struct {
	Titles []string   `json:"titles" bson:"titles"`
	Rows   [][]string `json:"rows" bson:"rows"`
}
