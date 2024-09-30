package service

import (
	"context"
	"fmt"
	"regexp"
	"runtime"

	"ucode/ucode_go_auth_service/api/models"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	pbCompany "ucode/ucode_go_auth_service/genproto/company_service"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/security"
	"ucode/ucode_go_auth_service/storage"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opentracing/opentracing-go"
	"github.com/saidamir98/udevs_pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type syncUserService struct {
	cfg         config.BaseConfig
	log         logger.LoggerI
	strg        storage.StorageI
	services    client.ServiceManagerI
	serviceNode ServiceNodesI
	pb.UnimplementedSyncUserServiceServer
}

func NewSyncUserService(cfg config.BaseConfig, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI, projectServiceNodes ServiceNodesI) *syncUserService {
	return &syncUserService{
		cfg:         cfg,
		log:         log,
		strg:        strg,
		services:    svcs,
		serviceNode: projectServiceNodes,
	}
}

func (sus *syncUserService) CreateUser(ctx context.Context, req *pb.CreateSyncUserRequest) (*pb.SyncUserResponse, error) {
	sus.log.Info("---CreateSyncUser--->", logger.Any("req", req))

	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_sync_user.CreateUser")
	defer dbSpan.Finish()
	var (
		response = pb.SyncUserResponse{}
		before   runtime.MemStats
		user     *pb.User
		err      error
		username string
		skip     bool
	)

	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		sus.log.Info("Memory used by the SyncUserCreateUser", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			sus.log.Info("Memory used over 300 mb", logger.Any("SyncUserCreateUser", memoryUsed))
		}
	}()

	for _, loginStrategy := range req.GetLoginStrategy() {
		if loginStrategy == "login" {
			username = req.GetLogin()
			skip = true
		} else if loginStrategy == "email" {
			username = req.GetEmail()
		} else if loginStrategy == "phone" {
			username = req.GetPhone()
		}
		if username != "" {
			user, err = sus.strg.User().GetByUsername(context.Background(), username)
			if err != nil {
				sus.log.Error("!!!CreateUser-->UserGetByUsername", logger.Error(err))
				return nil, err
			}

			if skip {
				if len(user.GetId()) > 0 {
					sus.log.Error("!!!CreateSyncUser-->LoginCheck", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, "user with this login already exists")
				}
			}
		}
	}

	userId := user.GetId()
	project, err := sus.services.ProjectServiceClient().GetById(context.Background(), &pbCompany.GetProjectByIdRequest{
		ProjectId: req.GetProjectId(),
	})
	if err != nil {
		sus.log.Error("!!!CreateUser-->ProjectGetById", logger.Error(err))
		return nil, err
	}

	if userId == "" {
		if req.GetPassword() != "" {
			hashedPassword, err := security.HashPasswordBcrypt(req.GetPassword())
			if err != nil {
				sus.log.Error("!!!CreateUser-->HashPassword", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			req.Password = hashedPassword
		}

		user, err := sus.strg.User().Create(ctx, &pb.CreateUserRequest{
			Login:     req.GetLogin(),
			Password:  req.GetPassword(),
			Email:     req.GetEmail(),
			Phone:     req.GetPhone(),
			CompanyId: project.GetCompanyId(),
		})
		if err != nil {
			sus.log.Error("!!!CreateUser--->UserCreate", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userId = user.GetId()
		_, err = sus.strg.User().AddUserToProject(context.Background(), &pb.AddUserToProjectReq{
			UserId:       userId,
			CompanyId:    project.GetCompanyId(),
			RoleId:       req.GetRoleId(),
			ProjectId:    req.GetProjectId(),
			ClientTypeId: req.GetClientTypeId(),
			EnvId:        req.GetEnvironmentId(),
		})
		if err != nil {
			sus.log.Error("!!!CreateUser--->AddUserToProject", logger.Error(err))
			return nil, err
		}
	} else {
		exists, err := sus.strg.User().GetUserProjectByAllFields(context.Background(), models.GetUserProjectByAllFieldsReq{
			ClientTypeId: req.GetClientTypeId(),
			RoleId:       req.GetRoleId(),
			UserId:       userId,
			CompanyId:    project.GetCompanyId(),
			ProjectId:    req.GetProjectId(),
			EnvId:        req.GetEnvironmentId(),
		})
		if err != nil {
			sus.log.Error("!!!CreateUser--->GetUserProjectByAllFields", logger.Error(err))
			return nil, err
		}
		if !exists {
			_, err = sus.strg.User().AddUserToProject(context.Background(), &pb.AddUserToProjectReq{
				UserId:       userId,
				CompanyId:    project.GetCompanyId(),
				RoleId:       req.GetRoleId(),
				ProjectId:    req.GetProjectId(),
				ClientTypeId: req.GetClientTypeId(),
				EnvId:        req.GetEnvironmentId(),
			})
			if err != nil {
				sus.log.Error("!!!CreateUser--->AddUserToProjectExists", logger.Error(err))
				return nil, err
			}
		}
	}

	if req.GetInvite() {
		emailSettings, err := sus.strg.Email().GetListEmailSettings(ctx, &pb.GetListEmailSettingsRequest{
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			sus.log.Error("!!!CreateUser--->AddUserToProjectExists", logger.Error(err))
			return nil, err
		}
		var devEmail string
		var devEmailPassword string
		if len(emailSettings.GetItems()) > 0 {
			devEmail = emailSettings.GetItems()[0].GetEmail()
			devEmailPassword = emailSettings.GetItems()[0].GetPassword()
		}
		err = helper.SendInviteMessageToEmail(helper.SendMessageToEmailRequest{
			Subject:       "Invite message",
			To:            req.GetEmail(),
			UserId:        userId,
			Email:         devEmail,
			Password:      devEmailPassword,
			Username:      req.GetLogin(),
			TempPassword:  req.GetPassword(),
			EnvironmentId: req.GetEnvironmentId(),
			ClientTypeId:  req.GetClientTypeId(),
			ProjectId:     req.GetProjectId(),
		},
		)
		if err != nil {
			sus.log.Error("Error while sending message to invite")
			sus.log.Error(err.Error())
		}
	}
	response.UserId = userId
	return &response, nil
}

func (sus *syncUserService) DeleteUser(ctx context.Context, req *pb.DeleteSyncUserRequest) (*empty.Empty, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_sync_user.DeleteUser")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		sus.log.Info("Memory used by the SyncUserDeleteUser", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			sus.log.Info("Memory used over 300 mb", logger.Any("SyncUserDeleteUser", memoryUsed))
		}
	}()

	_, err := sus.strg.User().DeleteUserFromProject(ctx, &pb.DeleteSyncUserRequest{
		UserId:        req.GetUserId(),
		RoleId:        req.GetRoleId(),
		ProjectId:     req.GetProjectId(),
		ClientTypeId:  req.GetClientTypeId(),
		EnvironmentId: req.GetEnvironmentId(),
	})
	if err != nil {
		sus.log.Info("!!!DeleteSyncUser-->DeleteUserFromProject", logger.Error(err))
		return nil, err
	}

	// if req.GetProjectId() != "42ab0799-deff-4f8c-bf3f-64bf9665d304" {
	// 	_, _ = sus.strg.User().Delete(context.Background(), &pb.UserPrimaryKey{
	// 		Id: req.GetUserId(),
	// 	})
	// }

	return &empty.Empty{}, nil
}

func (sus *syncUserService) UpdateUser(ctx context.Context, req *pb.UpdateSyncUserRequest) (*pb.SyncUserResponse, error) {
	sus.log.Info("---UpdateSyncUser--->", logger.Any("req", req))

	var (
		before                        runtime.MemStats
		hashedPassword                string
		err                           error
		syncUser                      = &pb.SyncUserResponse{}
		hasPassword, hasLoginStrategy bool
	)

	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_sync_user.UpdateUser")
	defer dbSpan.Finish()

	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		sus.log.Info("Memory used by the SyncUserUpdateUser", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			sus.log.Info("Memory used over 300 mb", logger.Any("SyncUserUpdateUser", memoryUsed))
		}
	}()

	project, err := sus.services.ProjectServiceClient().GetById(ctx,
		&pbCompany.GetProjectByIdRequest{ProjectId: req.GetProjectId()})
	if err != nil {
		sus.log.Error("!!!UpdateSyncUser-->ProjectGetById", logger.Error(err))
		return nil, err
	}

	req.CompanyId = project.GetCompanyId()

	if len(req.Password) > 0 {
		if len(req.Password) < 6 {
			err = fmt.Errorf("password must not be less than 6 characters")
			sus.log.Error("!!!UpdateSyncUser--->CheckPassword", logger.Error(err))
			return nil, err
		}

		hashedPassword, err = security.HashPasswordBcrypt(req.Password)
		if err != nil {
			sus.log.Error("!!!UpdateSyncUser--->HashPassword", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		hasPassword = true
		req.Password = hashedPassword
	}

	if len(req.GetLogin()) > 0 {
		if len(req.Login) < 6 {
			err = fmt.Errorf("login must not be less than 6 characters")
			sus.log.Error("!!!UpdateSyncUser--->CheckLogin", logger.Error(err))
			return nil, err
		}

		syncUser, err = sus.strg.User().UpdateSyncUser(ctx, req, "login")
		if err != nil {
			sus.log.Error("!!!UpdateSyncUser--->UpdateSyncUserLogin", logger.Error(err))
			return nil, err
		}

		hasLoginStrategy = true
	}

	if len(req.GetEmail()) > 0 {
		if !IsValidEmailNew(req.Email) {
			err = fmt.Errorf("email is not valid")
			sus.log.Error("!!!UpdateSyncUser--->CheckValidEmail", logger.Error(err))
			return nil, err
		}

		syncUser, err = sus.strg.User().UpdateSyncUser(ctx, req, "email")
		if err != nil {
			sus.log.Error("!!!UpdateSyncUser--->UpdateSyncUserEmail", logger.Error(err))
			return nil, err
		}

		hasLoginStrategy = true
	}

	if len(req.GetPhone()) > 0 {
		syncUser, err = sus.strg.User().UpdateSyncUser(ctx, req, "phone")
		if err != nil {
			sus.log.Error("!!!UpdateSyncUser--->UpdateSyncUserPhone", logger.Error(err))
			return nil, err
		}

		hasLoginStrategy = true
	}

	if hasPassword && !hasLoginStrategy {
		_, err = sus.strg.User().ResetPassword(ctx, &pb.ResetPasswordRequest{
			UserId:   req.GetGuid(),
			Password: hashedPassword}, nil)
		if err != nil {
			sus.log.Error("!!!UpdateSyncUser--->UpdateSyncUserPassword", logger.Error(err))
			return nil, err
		}

		syncUser.UserId = req.GetGuid()
	}

	if !hasLoginStrategy && !hasPassword {
		syncUser.UserId = req.GetGuid()
	}

	return syncUser, nil
}

func (sus *syncUserService) DeleteManyUser(ctx context.Context, req *pb.DeleteManyUserRequest) (*empty.Empty, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_sync_user.DeleteManyUser")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		sus.log.Info("Memory used by the SyncDeleteManyUser", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			sus.log.Info("Memory used over 300 mb", logger.Any("SyncDeleteManyUser", memoryUsed))
		}
	}()

	project, err := sus.services.ProjectServiceClient().GetById(ctx, &pbCompany.GetProjectByIdRequest{
		ProjectId: req.GetProjectId(),
	})
	if err != nil {
		sus.log.Error("!!!DeleteManyUser--->GetProjectById", logger.Error(err))
		return nil, err
	}
	req.CompanyId = project.CompanyId

	_, err = sus.strg.User().DeleteUsersFromProject(ctx, req)
	if err != nil {
		sus.log.Error("!!!DeleteManyUser--->DeleteUsersFromProject", logger.Error(err))
		return nil, err
	}

	// if req.GetProjectId() != "42ab0799-deff-4f8c-bf3f-64bf9665d304" {
	// 	for _, v := range req.Users {
	// 		_, _ = sus.strg.User().Delete(context.Background(), &pb.UserPrimaryKey{
	// 			Id: v.GetUserId(),
	// 		})
	// 	}
	// }

	return &empty.Empty{}, nil
}

func (sus *syncUserService) CreateUsers(ctx context.Context, in *pb.CreateSyncUsersRequest) (*pb.SyncUsersResponse, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_sync_user.CreateUsers")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		sus.log.Info("Memory used by the SyncCreateUsers", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			sus.log.Info("Memory used over 300 mb", logger.Any("SyncCreateUsers", memoryUsed))
		}
	}()

	var (
		response = pb.SyncUsersResponse{}
		user_ids = make([]string, 0, len(in.Users))
		err      error
	)
	for _, req := range in.Users {
		var (
			user     *pb.User
			username string
		)
		for _, loginStrategy := range req.GetLoginStrategy() {
			if loginStrategy == "login" {
				username = req.GetLogin()
			} else if loginStrategy == "email" {
				username = req.GetEmail()
			} else if loginStrategy == "phone" {
				username = req.GetPhone()
			}
			if username != "" {
				user, err = sus.strg.User().GetByUsername(context.Background(), username)
				if err != nil {
					sus.log.Error("!!!CreateSyncUsers--->GetUserByUsername", logger.Error(err))
					return nil, err
				}
			}
		}

		userId := user.GetId()
		project, err := sus.services.ProjectServiceClient().GetById(context.Background(), &pbCompany.GetProjectByIdRequest{
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			sus.log.Error("!!!CreateSyncUsers--->GetProjectById", logger.Error(err))
			return nil, err
		}

		if userId == "" {
			if req.GetPassword() != "" {
				hashedPassword, err := security.HashPasswordBcrypt(req.GetPassword())
				if err != nil {
					sus.log.Error("!!!CreateSyncUsers--->HashPasswordBcrypt", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
				req.Password = hashedPassword
			}

			user, err := sus.strg.User().Create(ctx, &pb.CreateUserRequest{
				Login:     req.GetLogin(),
				Password:  req.GetPassword(),
				Email:     req.GetEmail(),
				Phone:     req.GetPhone(),
				CompanyId: project.GetCompanyId(),
			})
			if err != nil {
				sus.log.Error("!!!CreateSyncUsers--->CreateUser", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			userId = user.GetId()
			_, err = sus.strg.User().AddUserToProject(context.Background(), &pb.AddUserToProjectReq{
				UserId:       userId,
				CompanyId:    project.GetCompanyId(),
				RoleId:       req.GetRoleId(),
				ProjectId:    req.GetProjectId(),
				ClientTypeId: req.GetClientTypeId(),
				EnvId:        req.GetEnvironmentId(),
			})
			if err != nil {
				sus.log.Error("!!!CreateSyncUsers-->AddUserToProject", logger.Error(err))
				return nil, err
			}
		} else {
			exists, err := sus.strg.User().GetUserProjectByAllFields(context.Background(), models.GetUserProjectByAllFieldsReq{
				ClientTypeId: req.GetClientTypeId(),
				RoleId:       req.GetRoleId(),
				UserId:       userId,
				CompanyId:    project.GetCompanyId(),
				ProjectId:    req.GetProjectId(),
				EnvId:        req.GetEnvironmentId(),
			})
			if err != nil {
				sus.log.Error("!!!CreateSyncUsers-->GetUserProjectByAllFields", logger.Error(err))
				return nil, err
			}
			if !exists {
				_, err = sus.strg.User().AddUserToProject(context.Background(), &pb.AddUserToProjectReq{
					UserId:       userId,
					CompanyId:    project.GetCompanyId(),
					RoleId:       req.GetRoleId(),
					ProjectId:    req.GetProjectId(),
					ClientTypeId: req.GetClientTypeId(),
					EnvId:        req.GetEnvironmentId(),
				})
				if err != nil {
					sus.log.Error("!!!CreateSyncUsers-->AddUserToProjectExists", logger.Error(err))
					return nil, err
				}
			}
		}

		if req.GetInvite() {
			emailSettings, err := sus.strg.Email().GetListEmailSettings(ctx, &pb.GetListEmailSettingsRequest{
				ProjectId: req.GetProjectId(),
			})
			if err != nil {
				return nil, err
			}
			var devEmail, devEmailPassword string

			if len(emailSettings.GetItems()) > 0 {
				devEmail = emailSettings.GetItems()[0].GetEmail()
				devEmailPassword = emailSettings.GetItems()[0].GetPassword()
			}
			err = helper.SendInviteMessageToEmail(helper.SendMessageToEmailRequest{
				Subject:       "Invite message",
				To:            req.GetEmail(),
				UserId:        userId,
				Email:         devEmail,
				Password:      devEmailPassword,
				Username:      req.GetLogin(),
				TempPassword:  req.GetPassword(),
				EnvironmentId: req.GetEnvironmentId(),
				ClientTypeId:  req.GetClientTypeId(),
				ProjectId:     req.GetProjectId(),
			},
			)
			if err != nil {
				sus.log.Error("Error while sending message to invite")
				sus.log.Error(err.Error())
			}
		}
		user_ids = append(user_ids, userId)
	}

	response.UserIds = user_ids

	return &response, nil
}

func IsValidEmailNew(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	re := regexp.MustCompile(emailRegex)

	return re.MatchString(email)
}
