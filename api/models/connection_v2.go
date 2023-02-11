package models

type CreateConnectionRequest struct {
	TableSlug    string `json:"table_slug"`
	ViewSlug     string `json:"view_slug"`
	ClientTypeId string `json:"client_type_id"`
	Name         string `json:"name"`
	ProjectId    string `json:"project_id"`
}

type Connection struct {
	Guid         string `json:"guid"`
	TableSlug    string `json:"table_slug"`
	ViewSlug     string `json:"view_slug"`
	ClientTypeId string `json:"client_type_id"`
	Name         string `json:"name"`
	ProjectId    string `json:"project_id"`
}