package helper

import (
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
)

func ConvertPbToAnotherPb(data *pbObject.V2LoginResponse) *pb.V2LoginResponse {
	res := &pb.V2LoginResponse{}
	res.UserId = data.UserId
	tables := make([]*pb.Table, 0, len(data.ClientType.Tables))
	for _, v := range data.ClientType.Tables {
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
		Id:           data.ClientType.Guid,
		Name:         data.ClientType.Name,
		ConfirmBy:    pb.ConfirmStrategies(data.ClientType.ConfirmBy),
		SelfRegister: data.ClientType.SelfRegister,
		SelfRecover:  data.ClientType.SelfRecover,
		ProjectId:    data.ClientType.ProjectId,
		Tables:       tables,
	}

	res.ClientPlatform = &pb.ClientPlatform{
		Id:        data.ClientPlatform.Guid,
		Name:      data.ClientPlatform.Name,
		ProjectId: data.ClientPlatform.ProjectId,
		Subdomain: data.ClientPlatform.Subdomain,
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
	res.Permissions = permissions
	res.Role = &pb.Role{
		Id:               data.Role.Guid,
		ClientTypeId:     data.Role.ClientTypeId,
		Name:             data.Role.Name,
		ClientPlatformId: data.Role.ClientPlatformId,
		ProjectId:        data.Role.ProjectId,
	}
	return res
}

func TokenGenerator(input *pbObject.V2LoginResponse) (*pb.V2LoginResponse, error) {
	return nil, nil
}
