package models

type CommonMessage struct {
	TableSlug string                 `json:"table_slug"`
	Data      map[string]interface{} `json:"data"`
}
