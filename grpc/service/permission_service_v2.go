package service

import (
	"context"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"

	"github.com/saidamir98/udevs_pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *permissionService) V2AddRole(ctx context.Context, req *pb.V2AddRoleRequest) (*pb.CommonMessage, error) {
	s.log.Info("---AddRole--->", logger.Any("req", req))
	var (
		result *pbObject.CommonMessage
	)

	// pKey, err := s.strg.Role().Add(ctx, req)
	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!AddRole--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	switch req.ResourceType {
	case 1:
		result, err = s.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			s.log.Error("!!!AddRole.ObjectBuilderService.Create--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		result, err = s.services.PostgresObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			s.log.Error("!!!AddRole.PostgresObjectBuilderService.Create--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *permissionService) V2GetRoleById(ctx context.Context, req *pb.V2RolePrimaryKey) (*pb.CommonMessage, error) {
	s.log.Info("---GetRoleById--->", logger.Any("req", req))

	var (
		result *pbObject.CommonMessage
	)

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!GetRoleById--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	switch req.ResourceType {
	case 1:
		result, err = s.services.ObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			s.log.Error("!!!GetRoleById.ObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		result, err = s.services.PostgresObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			s.log.Error("!!!GetRoleById.PostgresObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *permissionService) V2GetRolesList(ctx context.Context, req *pb.V2GetRolesListRequest) (*pb.CommonMessage, error) {
	s.log.Info("---GetRolesList--->", logger.Any("req", req))

	var (
		result *pbObject.CommonMessage
	)
	structData, err := helper.ConvertRequestToSturct(map[string]interface{}{
		// "client_platform_id": req.GetClientPlatformId(),
		"client_type_id": req.GetClientTypeId(),
	})
	if err != nil {
		s.log.Error("!!!GetRolesList--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	switch req.ResourceType {
	case 1:
		result, err = s.services.ObjectBuilderService().GetList(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			s.log.Error("!!!GetRolesList.ObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		result, err = s.services.PostgresObjectBuilderService().GetList(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			s.log.Error("!!!GetRolesList.PostgresObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *permissionService) V2UpdateRole(ctx context.Context, req *pb.V2UpdateRoleRequest) (*pb.CommonMessage, error) {
	s.log.Info("---UpdateRole--->", logger.Any("req", req))

	var (
		result *pbObject.CommonMessage
	)

	projectId := req.GetProjectId()
	req.ProjectId = req.GetDbProjectId()

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!UpdateRole--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	switch req.ResourceType {
	case 1:
		result, err = s.services.ObjectBuilderService().Update(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: projectId,
		})
		if err != nil {
			s.log.Error("!!!UpdateRole.ObjectBuilderService.Update--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		result, err = s.services.PostgresObjectBuilderService().Update(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: projectId,
		})
		if err != nil {
			s.log.Error("!!!UpdateRole.PostgresObjectBuilderService.Update--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil

}

func (s *permissionService) V2RemoveRole(ctx context.Context, req *pb.V2RolePrimaryKey) (*pb.CommonMessage, error) {
	s.log.Info("---RemoveRole--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!GetRoleById--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var (
		result *pbObject.CommonMessage
	)
	switch req.ResourceType {
	case 1:
		result, err = s.services.ObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			s.log.Error("!!!GetRoleById.ObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		_, err = s.services.ObjectBuilderService().Delete(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			s.log.Error("!!!GetRoleById.ObjectBuilderService.Delete--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		result, err = s.services.PostgresObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			s.log.Error("!!!GetRoleById.PostgresObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		_, err = s.services.PostgresObjectBuilderService().Delete(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			s.log.Error("!!!GetRoleById.PostgresObjectBuilderService.Delete--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *permissionService) V2CreatePermission(ctx context.Context, req *pb.CreatePermissionRequest) (*pb.CommonMessage, error) {
	s.log.Info("---CreatePermission--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!CreatePermission--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
		TableSlug: "permission",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!CreatePermission.ObjectBuilderService.Create--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *permissionService) V2GetPermissionByID(ctx context.Context, req *pb.PermissionPrimaryKey) (*pb.CommonMessage, error) {
	s.log.Info("---GetPermissionByID--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!GetPermissionByID--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.services.ObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
		TableSlug: "permission",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!GetPermissionByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *permissionService) V2GetPermissionList(ctx context.Context, req *pb.GetPermissionListRequest) (*pb.CommonMessage, error) {
	s.log.Info("---GetPermissionList--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!GetPermissionList--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.services.ObjectBuilderService().GetList(ctx, &pbObject.CommonMessage{
		TableSlug: "permission",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!GetPermissionList.ObjectBuilderService.GetList--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *permissionService) V2UpdatePermission(ctx context.Context, req *pb.UpdatePermissionRequest) (*pb.CommonMessage, error) {
	s.log.Info("---UpdatePermission--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!UpdatePermission--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.services.ObjectBuilderService().Update(ctx, &pbObject.CommonMessage{
		TableSlug: "permission",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!UpdatePermission.ObjectBuilderService.Update--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil

}

func (s *permissionService) V2DeletePermission(ctx context.Context, req *pb.PermissionPrimaryKey) (*emptypb.Empty, error) {
	s.log.Info("---DeletePermission--->", logger.Any("req", req))

	res := &emptypb.Empty{}
	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!DeletePermission--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = s.services.ObjectBuilderService().Delete(ctx, &pbObject.CommonMessage{
		TableSlug: "permission",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!DeletePermission.ObjectBuilderService.Delete--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res, nil
}

func (s *permissionService) V2GetScopesList(ctx context.Context, req *pb.GetScopeListRequest) (*pb.CommonMessage, error) {
	s.log.Info("---GetScopesList--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!GetScopesList--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result, err := s.services.ObjectBuilderService().GetList(ctx, &pbObject.CommonMessage{
		TableSlug: "scope",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!GetScopesList..ObjectBuilderService.GetList--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil

}

func (s *permissionService) V2AddPermissionScope(ctx context.Context, req *pb.AddPermissionScopeRequest) (*pb.CommonMessage, error) {
	s.log.Info("---AddPermissionScope--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!AddPermissionScope--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result, err := s.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
		TableSlug: "permission_scope",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!AddPermissionScope..ObjectBuilderService.Create--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *permissionService) V2RemovePermissionScope(ctx context.Context, req *pb.PermissionScopePrimaryKey) (*pb.CommonMessage, error) {
	s.log.Info("---RemovePermissionScope--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!RemovePermissionScope--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result, err := s.services.ObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
		TableSlug: "permission_scope",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!RemovePermissionScope..ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	_, err = s.services.ObjectBuilderService().Delete(ctx, &pbObject.CommonMessage{
		TableSlug: "permission_scope",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!RemovePermissionScope..ObjectBuilderService.Create--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *permissionService) V2AddRolePermission(ctx context.Context, req *pb.AddRolePermissionRequest) (*pb.CommonMessage, error) {
	s.log.Info("---AddRolePermission--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!AddRolePermission--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result, err := s.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
		TableSlug: "role_permission",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!AddRolePermission..ObjectBuilderService.Create--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *permissionService) V2RemoveRolePermission(ctx context.Context, req *pb.RolePermissionPrimaryKey) (*pb.CommonMessage, error) {
	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!RemoveRolePermission--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result, err := s.services.ObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
		TableSlug: "role_permission",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!RemoveRolePermission.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	_, err = s.services.ObjectBuilderService().Delete(ctx, &pbObject.CommonMessage{
		TableSlug: "roe_permission",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!RemoveRolePermission.ObjectBuilderService.Create--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil

}
