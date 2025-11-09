package models

type DeleteData struct {
	Table string                 `json:"table"`
	Where map[string]interface{} `json:"where"`
}

type UpsertData struct {
	Table  string                 `json:"table"`
	Fields []string               `json:"fields"`
	Data   map[string]interface{} `json:"data"`
}
