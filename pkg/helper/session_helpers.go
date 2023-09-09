package helper

import (
	"log"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
)

func ConvertPbToAnotherPb(data *pbObject.V2LoginResponse) *pb.V2LoginResponse {

	log.Printf("ConvertPbToAnotherPb: %v", data)

	res := &pb.V2LoginResponse{}
	res.UserId = data.GetUserId()
	res.LoginTableSlug = data.GetLoginTableSlug()
	tables := make([]*pb.Table, 0, len(data.GetClientType().GetTables()))
	for _, v := range data.GetClientType().GetTables() {
		table := &pb.Table{}
		table.Data = v.GetData()
		table.Icon = v.GetIcon()
		table.Label = v.GetLabel()
		table.Slug = v.GetSlug()
		table.ViewLabel = v.GetViewLabel()
		table.ViewSlug = v.GetViewSlug()
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
		permission.Automation = v.GetAutomation()
		permission.LanguageBtn = v.GetLanguageBtn()
		permission.Settings = v.GetSettings()
		permission.ShareModal = v.GetShareModal()
		permission.ViewCreate = v.GetViewCreate()
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

	res.GlobalPermission = &pb.GlobalPermission{
		Id:                    data.GetGlobalPermission().GetId(),
		MenuButton:            data.GetGlobalPermission().GetMenuButton(),
		Chat:                  data.GetGlobalPermission().GetChat(),
		SettingsButton:        data.GetGlobalPermission().GetSettingsButton(),
		ProjectSettingsButton: data.GetGlobalPermission().GetProjectSettingsButton(),
		ProfileSettingsButton: data.GetGlobalPermission().GetProfileSettingsButton(),
		MenuSettingButton:     data.GetGlobalPermission().GetMenuSettingButton(),
		RedirectsButton:       data.GetGlobalPermission().GetRedirectsButton(),
		ApiKeysButton:         data.GetGlobalPermission().GetApiKeysButton(),
		EnvironmentsButton:    data.GetGlobalPermission().GetEnvironmentsButton(),
		ProjectsButton:        data.GetGlobalPermission().GetProjectsButton(),
		VersionButton:         data.GetGlobalPermission().GetVersionButton(),
		EnvironmentButton:     data.GetGlobalPermission().GetEnvironmentButton(),
		ProjectButton:         data.GetGlobalPermission().GetProjectButton(),
		SmsButton:             data.GetGlobalPermission().GetSmsButton(),
	}

	res.UserData = data.GetUserData()
	return res
}
