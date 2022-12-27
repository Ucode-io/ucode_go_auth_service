package models

type GetUserProjects struct {
	Companies []Companie `json:"companies"`
}

type Companie struct {
	Id       string   `json:"id"`
	Projects []string `json:"projects"`
}
