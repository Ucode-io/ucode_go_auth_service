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
		Id:           data.GetClientType().Guid,
		Name:         data.GetClientType().Name,
		ConfirmBy:    pb.ConfirmStrategies(data.ClientType.ConfirmBy),
		SelfRegister: data.GetClientType().SelfRegister,
		SelfRecover:  data.GetClientType().SelfRecover,
		ProjectId:    data.GetClientType().ProjectId,
		Tables:       tables,
	}

	res.ClientPlatform = &pb.ClientPlatform{
		Id:        data.GetClientPlatform().GetGuid(),
		Name:      data.GetClientPlatform().GetName(),
		ProjectId: data.GetClientPlatform().GetProjectId(),
		Subdomain: data.GetClientPlatform().GetSubdomain(),
	}
	permissions := make([]*pb.RecordPermission, 0, len(data.Permissions))
	for _, v := range data.Permissions {
		permission := &pb.RecordPermission{}
		permission.ClientTypeId = v.ClientTypeId
		permission.Id = v.Guid
		permission.Read = v.Read
		permission.Write = v.Write
		permission.Delete = v.Delete
		permission.Update = v.Update
		permission.RoleId = v.RoleId
		permission.TableSlug = v.TableSlug
		permissions = append(permissions, permission)
	}

	appPermissions := make([]*pb.RecordPermission, 0, len(data.Permissions))
	for _, v := range data.AppPermissions {
		appPermission := &pb.RecordPermission{}
		appPermission.ClientTypeId = v.ClientTypeId
		appPermission.Id = v.Guid
		appPermission.Read = v.Read
		appPermission.Write = v.Write
		appPermission.Delete = v.Delete
		appPermission.Update = v.Update
		appPermission.RoleId = v.RoleId
		appPermission.TableSlug = v.TableSlug
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
