package models

type GetUserProjects struct {
	Companies []Companie `json:"companies"`
}

type Companie struct {
	Id       string   `json:"id"`
	Projects []string `json:"projects"`
}

type GetUserProjectByAllFieldsReq struct {
	ClientTypeId string `json:"client_type_id"`
	RoleId       string `json:"role_id"`
	UserId       string `json:"user_id"`
	CompanyId    string `json:"company_id"`
	ProjectId    string `json:"project_id"`
}
