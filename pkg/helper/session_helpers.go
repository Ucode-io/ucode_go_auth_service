package helper

import (
	"log"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
)

func ConvertPbToAnotherPb(data *pbObject.V2LoginResponse) *pb.V2LoginResponse {
	log.Printf("---INFO-->ConvertPbToAnotherPb---> %+v", data)

	res := &pb.V2LoginResponse{}
	res.UserId = data.UserId
	res.LoginTableSlug = data.LoginTableSlug
	tables := make([]*pb.Table, 0, len(data.GetClientType().GetTables()))
	for _, v := range data.GetClientType().GetTables() {
		table := &pb.Table{}
		table.Data = v.Data
		table.Icon = v.Icon
		table.Label = v.Label
		table.Slug = v.Slug
		table.ViewLabel = v.ViewLabel
		table.ViewSlug = v.ViewSlug
		tables = append(tables, table)
	}

	res.ClientType = &pb.ClientType{
		Id:           data.GetClientType().GetGuid(),
		Name:         data.GetClientType().GetName(),
		ConfirmBy:    pb.ConfirmStrategies(data.GetClientType().GetConfirmBy()),
		SelfRegister: data.GetClientType().GetSelfRegister(),
		SelfRecover:  data.GetClientType().GetSelfRecover(),
		ProjectId:    data.GetClientType().GetProjectId(),
		Tables:       tables,
	}

	res.ClientPlatform = &pb.ClientPlatform{
		Id:        data.GetClientPlatform().GetGuid(),
		Name:      data.GetClientPlatform().GetName(),
		ProjectId: data.GetClientPlatform().GetProjectId(),
		Subdomain: data.GetClientPlatform().GetSubdomain(),
	}
	permissions := make([]*pb.RecordPermission, 0, len(data.GetPermissions()))
	for _, v := range data.GetPermissions() {
		permission := &pb.RecordPermission{}
		permission.ClientTypeId = v.GetClientTypeId()
		permission.Id = v.GetGuid()
		permission.Read = v.GetRead()
		permission.Write = v.GetWrite()
		permission.Delete = v.GetDelete()
		permission.Update = v.GetUpdate()
		permission.RoleId = v.GetRoleId()
		permission.TableSlug = v.GetTableSlug()
		permissions = append(permissions, permission)
	}

	appPermissions := make([]*pb.RecordPermission, 0, len(data.GetPermissions()))
	for _, v := range data.GetAppPermissions() {
		appPermission := &pb.RecordPermission{}
		appPermission.ClientTypeId = v.GetClientTypeId()
		appPermission.Id = v.GetGuid()
		appPermission.Read = v.GetRead()
		appPermission.Write = v.GetWrite()
		appPermission.Delete = v.GetDelete()
		appPermission.Update = v.GetUpdate()
		appPermission.RoleId = v.GetRoleId()
		appPermission.TableSlug = v.GetTableSlug()
		appPermissions = append(appPermissions, appPermission)
	}
	res.Permissions = permissions
	res.AppPermissions = appPermissions
	res.Role = &pb.Role{
		Id:               data.GetRole().GetGuid(),
		ClientTypeId:     data.GetRole().GetClientTypeId(),
		Name:             data.GetRole().GetName(),
		ClientPlatformId: data.GetRole().GetClientPlatformId(),
		ProjectId:        data.GetRole().GetProjectId(),
	}
	return res
}
