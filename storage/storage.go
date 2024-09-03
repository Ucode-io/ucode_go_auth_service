package storage

import (
	"context"
	"errors"
	"ucode/ucode_go_auth_service/api/models"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"

	// "github.com/jackc/pgconn/internal/ctxwatch"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrorTheSameId = errors.New("cannot use the same uuid for 'id' and 'parent_id' fields")
var ErrorProjectId = errors.New("not valid 'project_id'")

type StorageI interface {
	CloseDB()
	ClientPlatform() ClientPlatformRepoI
	ClientType() ClientTypeRepoI
	Client() ClientRepoI
	Relation() RelationRepoI
	UserInfoField() UserInfoFieldRepoI
	Role() RoleRepoI
	Permission() PermissionRepoI
	Scope() ScopeRepoI
	PermissionScope() PermissionScopeRepoI
	RolePermission() RolePermissionRepoI
	User() UserRepoI
	Integration() IntegrationRepoI
	UserRelation() UserRelationRepoI
	UserInfo() UserInfoRepoI
	Session() SessionRepoI
	Email() EmailRepoI
	Company() CompanyRepoI
	Project() ProjectRepoI
	ApiKeys() ApiKeysRepoI
	AppleSettings() AppleSettingsI
	LoginStrategy() LoginStrategyI
	LoginPlatformType() LoginPlatformType
	SmsOtpSettings() SmsOtpSettingsRepoI
	ApiKeyUsage() ApiKeyUsageRepoI
}

type ClientPlatformRepoI interface {
	Create(ctx context.Context, entity *pb.CreateClientPlatformRequest) (pKey *pb.ClientPlatformPrimaryKey, err error)
	GetList(ctx context.Context, queryParam *pb.GetClientPlatformListRequest) (res *pb.GetClientPlatformListResponse, err error)
	GetByPK(ctx context.Context, pKey *pb.ClientPlatformPrimaryKey) (res *pb.ClientPlatform, err error)
	GetByPKDetailed(ctx context.Context, pKey *pb.ClientPlatformPrimaryKey) (res *pb.ClientPlatformDetailedResponse, err error)
	Update(ctx context.Context, entity *pb.UpdateClientPlatformRequest) (rowsAffected int64, err error)
	Delete(ctx context.Context, pKey *pb.ClientPlatformPrimaryKey) (rowsAffected int64, err error)
}

type ClientTypeRepoI interface {
	Create(ctx context.Context, entity *pb.CreateClientTypeRequest) (pKey *pb.ClientTypePrimaryKey, err error)
	GetList(ctx context.Context, queryParam *pb.GetClientTypeListRequest) (res *pb.GetClientTypeListResponse, err error)
	GetByPK(ctx context.Context, pKey *pb.ClientTypePrimaryKey) (res *pb.ClientType, err error)
	Update(ctx context.Context, entity *pb.UpdateClientTypeRequest) (rowsAffected int64, err error)
	Delete(ctx context.Context, pKey *pb.ClientTypePrimaryKey) (rowsAffected int64, err error)
	GetCompleteByPK(ctx context.Context, pKey *pb.ClientTypePrimaryKey) (res *pb.CompleteClientType, err error)
}

type ClientRepoI interface {
	Add(ctx context.Context, projectID string, entity *pb.AddClientRequest) (err error)
	GetByPK(ctx context.Context, entity *pb.ClientPrimaryKey) (res *pb.Client, err error)
	Update(ctx context.Context, entity *pb.UpdateClientRequest) (rowsAffected int64, err error)
	Remove(ctx context.Context, entity *pb.ClientPrimaryKey) (rowsAffected int64, err error)
	GetList(ctx context.Context, queryParam *pb.GetClientListRequest) (res *pb.GetClientListResponse, err error)
	GetMatrix(ctx context.Context, req *pb.GetClientMatrixRequest) (res *pb.GetClientMatrixResponse, err error)
}

type RelationRepoI interface {
	Add(ctx context.Context, entity *pb.AddRelationRequest) (pKey *pb.RelationPrimaryKey, err error)
	GetByPK(ctx context.Context, entity *pb.RelationPrimaryKey) (res *pb.Relation, err error)
	Update(ctx context.Context, entity *pb.UpdateRelationRequest) (rowsAffected int64, err error)
	Remove(ctx context.Context, entity *pb.RelationPrimaryKey) (rowsAffected int64, err error)
}

type UserInfoFieldRepoI interface {
	Add(ctx context.Context, entity *pb.AddUserInfoFieldRequest) (pKey *pb.UserInfoFieldPrimaryKey, err error)
	GetByPK(ctx context.Context, entity *pb.UserInfoFieldPrimaryKey) (res *pb.UserInfoField, err error)
	Update(ctx context.Context, entity *pb.UpdateUserInfoFieldRequest) (rowsAffected int64, err error)
	Remove(ctx context.Context, entity *pb.UserInfoFieldPrimaryKey) (rowsAffected int64, err error)
}

type RoleRepoI interface {
	Add(ctx context.Context, entity *pb.AddRoleRequest) (pKey *pb.RolePrimaryKey, err error)
	GetByPK(ctx context.Context, entity *pb.RolePrimaryKey) (res *pb.Role, err error)
	GetList(ctx context.Context, entity *pb.GetRolesListRequest) (res *pb.GetRolesResponse, err error)
	GetRoleByIdDetailed(ctx context.Context, entity *pb.RolePrimaryKey) (res *pb.GetRoleByIdResponse, err error)
	Update(ctx context.Context, entity *pb.UpdateRoleRequest) (rowsAffected int64, err error)
	Remove(ctx context.Context, entity *pb.RolePrimaryKey) (rowsAffected int64, err error)
}

type PermissionRepoI interface {
	Create(ctx context.Context, entity *pb.CreatePermissionRequest) (pKey *pb.PermissionPrimaryKey, err error)
	GetList(ctx context.Context, queryParam *pb.GetPermissionListRequest) (res *pb.GetPermissionListResponse, err error)
	GetByPK(ctx context.Context, pKey *pb.PermissionPrimaryKey) (res *pb.GetPermissionByIDResponse, err error)
	Update(ctx context.Context, entity *pb.UpdatePermissionRequest) (rowsAffected int64, err error)
	Delete(ctx context.Context, pKey *pb.PermissionPrimaryKey) (rowsAffected int64, err error)
	GetListByClientPlatformId(ctx context.Context, clientPlatformID string) (res []*pb.Permission, err error)
	GetListByRoleId(ctx context.Context, roleID string) (res []*pb.Permission, err error)
}

type ScopeRepoI interface {
	Upsert(ctx context.Context, entity *pb.UpsertScopeRequest) (res *pb.ScopePrimaryKey, err error)
	GetByPK(ctx context.Context, pKey *pb.ScopePrimaryKey) (res *pb.Scope, err error)
	GetList(ctx context.Context, queryParam *pb.GetScopeListRequest) (res *pb.GetScopesResponse, err error)
}

type PermissionScopeRepoI interface {
	Add(ctx context.Context, entity *pb.AddPermissionScopeRequest) (res *pb.PermissionScopePrimaryKey, err error)
	Remove(ctx context.Context, entity *pb.PermissionScopePrimaryKey) (rowsAffected int64, err error)
	GetByPK(ctx context.Context, pKey *pb.PermissionScopePrimaryKey) (res *pb.PermissionScope, err error)
	HasAccess(ctx context.Context, roleID, clientPlatformID, path, method string) (hasAccess bool, err error)
}

type RolePermissionRepoI interface {
	Add(ctx context.Context, entity *pb.AddRolePermissionRequest) (res *pb.RolePermissionPrimaryKey, err error)
	AddMultiple(ctx context.Context, entity *pb.AddRolePermissionsRequest) (rowsAffected int64, err error)
	Remove(ctx context.Context, entity *pb.RolePermissionPrimaryKey) (rowsAffected int64, err error)
	GetByPK(ctx context.Context, pKey *pb.RolePermissionPrimaryKey) (res *pb.RolePermission, err error)
}

type UserRepoI interface {
	GetListByPKs(ctx context.Context, pKeys *pb.UserPrimaryKeyList) (res *pb.GetUserListResponse, err error)
	Create(ctx context.Context, entity *pb.CreateUserRequest) (pKey *pb.UserPrimaryKey, err error)
	GetList(ctx context.Context, queryParam *pb.GetUserListRequest) (res *pb.GetUserListResponse, err error)
	GetByPK(ctx context.Context, pKey *pb.UserPrimaryKey) (res *pb.User, err error)
	Update(ctx context.Context, entity *pb.UpdateUserRequest) (rowsAffected int64, err error)
	Delete(ctx context.Context, pKey *pb.UserPrimaryKey) (rowsAffected int64, err error)
	GetByUsername(ctx context.Context, username string) (res *pb.User, err error)
	ResetPassword(ctx context.Context, user *pb.ResetPasswordRequest) (rowsAffected int64, err error)
	GetUserProjects(ctx context.Context, userId string) (*pb.GetUserProjectsRes, error)
	GetUserProjectClientTypes(ctx context.Context, req *models.UserProjectClientTypeRequest) (*models.UserProjectClientTypeResponse, error)
	AddUserToProject(ctx context.Context, req *pb.AddUserToProjectReq) (*pb.AddUserToProjectRes, error)
	UpdateUserToProject(ctx context.Context, req *pb.AddUserToProjectReq) (*pb.AddUserToProjectRes, error)
	GetProjectsByUserId(ctx context.Context, req *pb.GetProjectsByUserIdReq) (*pb.GetProjectsByUserIdRes, error)
	GetUserIds(ctx context.Context, req *pb.GetUserListRequest) (*[]string, error)
	GetUserByLoginType(ctx context.Context, req *pb.GetUserByLoginTypesRequest) (*pb.GetUserByLoginTypesResponse, error)
	GetListTimezone(ctx context.Context, in *pb.GetListSettingReq) (*models.ListTimezone, error)
	GetListLanguage(ctx context.Context, in *pb.GetListSettingReq) (*models.ListLanguage, error)
	V2ResetPassword(ctx context.Context, req *pb.V2ResetPasswordRequest) (int64, error)
	GetUserProjectByAllFields(ctx context.Context, req models.GetUserProjectByAllFieldsReq) (bool, error)
	DeleteUserFromProject(ctx context.Context, req *pb.DeleteSyncUserRequest) (*emptypb.Empty, error)
	DeleteUsersFromProject(ctx context.Context, req *pb.DeleteManyUserRequest) (*emptypb.Empty, error)
	GetAllUserProjects(ctx context.Context) ([]string, error)
	UpdateUserProjects(ctx context.Context, envId, projectId string) (*emptypb.Empty, error)
	GetUserEnvProjects(ctx context.Context, userId string) (*models.GetUserEnvProjectRes, error)
	V2GetByUsername(ctx context.Context, id, projectId string) (res *pb.User, err error)
}

type IntegrationRepoI interface {
	GetListByPKs(ctx context.Context, pKeys *pb.IntegrationPrimaryKeyList) (res *pb.GetIntegrationListResponse, err error)
	Create(ctx context.Context, entity *pb.CreateIntegrationRequest) (pKey *pb.IntegrationPrimaryKey, err error)
	GetList(ctx context.Context, queryParam *pb.GetIntegrationListRequest) (res *pb.GetIntegrationListResponse, err error)
	CreateSession(ctx context.Context, entity *pb.CreateSessionRequest) (pKey *pb.SessionPrimaryKey, err error)
	GetByPK(ctx context.Context, pKey *pb.IntegrationPrimaryKey) (res *pb.Integration, err error)
	Update(ctx context.Context, entity *pb.UpdateIntegrationRequest) (rowsAffected int64, err error)
	Delete(ctx context.Context, pKey *pb.IntegrationPrimaryKey) (rowsAffected int64, err error)
	GetIntegrationSessions(ctx context.Context, pKey *pb.IntegrationPrimaryKey) (res *pb.GetIntegrationSessionsResponse, err error)
	DeleteSession(ctx context.Context, pKey *pb.GetIntegrationTokenRequest) (rowsAffected int64, err error)
	GetIntegrationSession(ctx context.Context, req *pb.GetIntegrationTokenRequest) (res *pb.Session, err error)
}

type UserRelationRepoI interface {
	Add(ctx context.Context, entity *pb.AddUserRelationRequest) (res *pb.UserRelationPrimaryKey, err error)
	Remove(ctx context.Context, entity *pb.UserRelationPrimaryKey) (rowsAffected int64, err error)
	GetByPK(ctx context.Context, pKey *pb.UserRelationPrimaryKey) (res *pb.UserRelation, err error)
}

type UserInfoRepoI interface {
	Upsert(ctx context.Context, entity *pb.UpsertUserInfoRequest) (res *pb.UserInfoPrimaryKey, err error)
	GetByPK(ctx context.Context, pKey *pb.UserInfoPrimaryKey) (res *pb.UserInfo, err error)
}

type SessionRepoI interface {
	Create(ctx context.Context, entity *pb.CreateSessionRequest) (pKey *pb.SessionPrimaryKey, err error)
	GetList(ctx context.Context, queryParam *pb.GetSessionListRequest) (res *pb.GetSessionListResponse, err error)
	GetByPK(ctx context.Context, pKey *pb.SessionPrimaryKey) (res *pb.Session, err error)
	Update(ctx context.Context, entity *pb.UpdateSessionRequest) (rowsAffected int64, err error)
	Delete(ctx context.Context, pKey *pb.SessionPrimaryKey) (rowsAffected int64, err error)
	DeleteExpiredUserSessions(ctx context.Context, userID string) (rowsAffected int64, err error)
	DeleteExpiredIntegrationSessions(ctx context.Context, userID string) (rowsAffected int64, err error)
	GetSessionListByUserID(ctx context.Context, userID string) (res *pb.GetSessionListResponse, err error)
	GetSessionListByIntegrationID(ctx context.Context, userID string) (res *pb.GetSessionListResponse, err error)
	UpdateByRoleId(ctx context.Context, entity *pb.UpdateSessionByRoleIdRequest) (rowsAffected int64, err error)
	ExpireSessions(ctx context.Context, entity *pb.ExpireSessionsRequest) (err error)
}

type EmailRepoI interface {
	Create(ctx context.Context, input *pb.Email) (*pb.Email, error)
	GetByPK(ctx context.Context, input *pb.EmailOtpPrimaryKey) (*pb.Email, error)
	CreateEmailSettings(ctx context.Context, input *pb.EmailSettings) (*pb.EmailSettings, error)
	UpdateEmailSettings(ctx context.Context, input *pb.UpdateEmailSettingsRequest) (*pb.EmailSettings, error)
	GetListEmailSettings(ctx context.Context, input *pb.GetListEmailSettingsRequest) (*pb.UpdateEmailSettingsResponse, error)
	DeleteEmailSettings(ctx context.Context, input *pb.EmailSettingsPrimaryKey) (*emptypb.Empty, error)
}

type CompanyRepoI interface {
	Register(ctx context.Context, entity *pb.RegisterCompanyRequest) (pKey *pb.CompanyPrimaryKey, err error)
	Update(ctx context.Context, entity *pb.UpdateCompanyRequest) (rowsAffected int64, err error)
	Remove(ctx context.Context, pKey *pb.CompanyPrimaryKey) (rowsAffected int64, err error)
	GetList(ctx context.Context, queryParam *pb.GetComapnyListRequest) (*pb.GetListCompanyResponse, error)
	GetByID(ctx context.Context, pKey *pb.CompanyPrimaryKey) (*pb.Company, error)
	TransferOwnership(ctx context.Context, companyID, ownerID string) (rowsAffected int64, err error)
}

type ProjectRepoI interface {
	Create(ctx context.Context, entity *pb.CreateProjectRequest) (pKey *pb.ProjectPrimaryKey, err error)
	GetList(ctx context.Context, queryParam *pb.GetProjectListRequest) (res *pb.GetProjectListResponse, err error)
	GetByPK(ctx context.Context, pKey *pb.ProjectPrimaryKey) (res *pb.Project, err error)
	Update(ctx context.Context, entity *pb.UpdateProjectRequest) (rowsAffected int64, err error)
	Delete(ctx context.Context, pKey *pb.ProjectPrimaryKey) (rowsAffected int64, err error)
}

type ApiKeysRepoI interface {
	Create(ctx context.Context, req *pb.CreateReq, appSecret, appId, id string) (res *pb.CreateRes, err error)
	GetList(ctx context.Context, req *pb.GetListReq) (res *pb.GetListRes, err error)
	Get(ctx context.Context, req *pb.GetReq) (res *pb.GetRes, err error)
	Update(ctx context.Context, req *pb.UpdateReq) (rowsAffected int64, err error)
	Delete(ctx context.Context, req *pb.DeleteReq) (rowsAffected int64, err error)
	GetByAppId(ctx context.Context, appId string) (*pb.GetRes, error)
	GetEnvID(ctx context.Context, req *pb.GetReq) (*pb.GetRes, error)
	UpdateIsMonthlyLimitReached(ctx context.Context) error
}

type AppleSettingsI interface {
	Create(ctx context.Context, input *pb.AppleIdSettings) (*pb.AppleIdSettings, error)
	GetByPK(ctx context.Context, pKey *pb.AppleIdSettingsPrimaryKey) (res *pb.AppleIdSettings, err error)
	UpdateAppleSettings(ctx context.Context, input *pb.AppleIdSettings) (string, error)
	GetListAppleSettings(ctx context.Context, input *pb.GetListAppleIdSettingsRequest) (*pb.GetListAppleIdSettingsResponse, error)
	DeleteAppleSettings(ctx context.Context, input *pb.AppleIdSettingsPrimaryKey) (*emptypb.Empty, error)
}

type LoginStrategyI interface {
	GetList(ctx context.Context, req *pb.GetListRequest) (res *pb.GetListResponse, err error)
	GetByID(ctx context.Context, req *pb.LoginStrategyPrimaryKey) (res *pb.LoginStrategy, err error)
	Upsert(ctx context.Context, req *pb.UpdateRequest) (res *pb.UpdateResponse, err error)
}

// LoginPlatformType
type LoginPlatformType interface {
	CreateLoginPlatformType(ctx context.Context, input *pb.LoginPlatform) (*pb.LoginPlatform, error)
	GetLoginPlatformType(ctx context.Context, pKey *pb.LoginPlatformTypePrimaryKey) (res *pb.LoginPlatform, err error)
	UpdateLoginPlatformType(ctx context.Context, input *pb.UpdateLoginPlatformTypeRequest, types string) (string, error)
	GetListLoginPlatformType(ctx context.Context, input *pb.GetListLoginPlatformTypeRequest) (*pb.GetListLoginPlatformTypeResponse, error)
	DeleteLoginSettings(ctx context.Context, input *pb.LoginPlatformTypePrimaryKey) (*emptypb.Empty, error)
}

// sms otp settings repo is used to save otp creds for each project and environment
type SmsOtpSettingsRepoI interface {
	Create(context.Context, *pb.CreateSmsOtpSettingsRequest) (*pb.SmsOtpSettings, error)
	Update(context.Context, *pb.SmsOtpSettings) (int64, error)
	GetById(context.Context, *pb.SmsOtpSettingsPrimaryKey) (*pb.SmsOtpSettings, error)
	GetList(context.Context, *pb.GetListSmsOtpSettingsRequest) (*pb.SmsOtpSettingsResponse, error)
	Delete(context.Context, *pb.SmsOtpSettingsPrimaryKey) (int64, error)
}

type ApiKeyUsageRepoI interface {
	CheckLimit(ctx context.Context, req *pb.CheckLimitRequest) (res *pb.CheckLimitResponse, err error)
	Create(ctx context.Context, req *pb.ApiKeyUsage) error
	Upsert(ctx context.Context, req *pb.ApiKeyUsage) error
	UpdateMonthlyLimit(ctx context.Context) error
}
