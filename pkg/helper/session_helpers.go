package helper

import (
	"fmt"
	"runtime"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
)

var mB = 1024 * 1024

func ConvertPbToAnotherPb(data *pbObject.V2LoginResponse) *pb.V2LoginResponse {
	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	tables := make([]*pb.Table, 0, len(data.GetClientType().GetTables()))
	permissions := make([]*pb.RecordPermission, 0, len(data.GetPermissions()))
	appPermissions := make([]*pb.RecordPermission, 0, len(data.GetPermissions()))

	for _, v := range data.GetClientType().GetTables() {
		tables = append(tables, &pb.Table{
			Data:      v.GetData(),
			Icon:      v.GetIcon(),
			Label:     v.GetLabel(),
			Slug:      v.GetSlug(),
			ViewLabel: v.GetViewLabel(),
			ViewSlug:  v.GetViewSlug(),
		})
	}

	for _, v := range data.GetPermissions() {
		permissions = append(permissions, &pb.RecordPermission{
			ClientTypeId: v.GetClientTypeId(),
			Id:           v.GetGuid(),
			Read:         v.GetRead(),
			Write:        v.GetWrite(),
			Delete:       v.GetDelete(),
			Update:       v.GetUpdate(),
			RoleId:       v.GetRoleId(),
			TableSlug:    v.GetTableSlug(),
			Automation:   v.GetAutomation(),
			LanguageBtn:  v.GetLanguageBtn(),
			Settings:     v.GetSettings(),
			ShareModal:   v.GetShareModal(),
			ViewCreate:   v.GetViewCreate(),
			AddField:     v.GetAddField(),
			PdfAction:    v.GetPdfAction(),
			AddFilter:    v.GetAddFilter(),
			FieldFilter:  v.GetFieldFilter(),
			FixColumn:    v.GetFixColumn(),
			Group:        v.GetGroup(),
			ExcelMenu:    v.GetExcelMenu(),
			TabGroup:     v.GetTabGroup(),
			SearchButton: v.GetSearchButton(),
			Columns:      v.GetColumns(),
		})

		appPermissions = append(appPermissions, &pb.RecordPermission{
			ClientTypeId: v.GetClientTypeId(),
			Id:           v.GetGuid(),
			Read:         v.GetRead(),
			Write:        v.GetWrite(),
			Delete:       v.GetDelete(),
			Update:       v.GetUpdate(),
			RoleId:       v.GetRoleId(),
			TableSlug:    v.GetTableSlug(),
		})

	}

	res := &pb.V2LoginResponse{
		UserId:         data.GetUserId(),
		LoginTableSlug: data.GetLoginTableSlug(),
		ClientType: &pb.ClientType{
			Id:           data.GetClientType().GetGuid(),
			Name:         data.GetClientType().GetName(),
			ConfirmBy:    pb.ConfirmStrategies(data.GetClientType().GetConfirmBy()),
			SelfRegister: data.GetClientType().GetSelfRegister(),
			SelfRecover:  data.GetClientType().GetSelfRecover(),
			ProjectId:    data.GetClientType().GetProjectId(),
			Tables:       tables,
			DefaultPage:  data.GetClientType().GetDefaultPage(),
		},
		ClientPlatform: &pb.ClientPlatform{
			Id:        data.GetClientPlatform().GetGuid(),
			Name:      data.GetClientPlatform().GetName(),
			ProjectId: data.GetClientPlatform().GetProjectId(),
			Subdomain: data.GetClientPlatform().GetSubdomain(),
		},
		Permissions:    permissions,
		AppPermissions: appPermissions,
		Role: &pb.Role{
			Id:               data.GetRole().GetGuid(),
			ClientTypeId:     data.GetRole().GetClientTypeId(),
			Name:             data.GetRole().GetName(),
			ClientPlatformId: data.GetRole().GetClientPlatformId(),
			ProjectId:        data.GetRole().GetProjectId(),
		},
		UserData: data.GetUserData(),
		GlobalPermission: &pb.GlobalPermission{
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
		},
	}

	runtime.ReadMemStats(&memAfter)
	fmt.Println("<<<<<<ConvertPbToAnotherPb Memory Usage>>>>>>")
	fmt.Printf("Alloc: %d bytes\n", (memAfter.Alloc-memBefore.Alloc)/uint64(mB))
	fmt.Printf("TotalAlloc: %d bytes\n", (memAfter.TotalAlloc-memBefore.TotalAlloc)/uint64(mB))
	fmt.Printf("HeapAlloc: %d bytes\n", (memAfter.HeapAlloc-memBefore.HeapAlloc)/uint64(mB))
	fmt.Printf("Mallocs: %d\n", (memAfter.Mallocs-memBefore.Mallocs)/uint64(mB))
	return res
}
