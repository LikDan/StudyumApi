package entities

type I18nEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type I18n = map[string]string

type I18nWithHash struct {
	Hash        string `json:"hash"`
	Translation I18n   `json:"translation"`
}
