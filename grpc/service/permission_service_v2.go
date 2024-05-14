package service

import (
	"context"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	nobs "ucode/ucode_go_auth_service/genproto/new_object_builder_service"
	"ucode/ucode_go_auth_service/genproto/object_builder_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"

	"github.com/saidamir98/udevs_pkg/logger"
	"github.com/spf13/cast"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	services, err := s.serviceNode.GetByNodeType(
		req.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	switch req.ResourceType {
	case 1:
		result, err = services.GetObjectBuilderServiceByType(req.NodeType).Create(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			s.log.Error("!!!AddRole.ObjectBuilderService.Create--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		roleData, _ := helper.ConvertStructToResponse(result.Data)
		roleDataData := cast.ToStringMap(roleData["data"])
		_, err = services.BuilderPermissionService().CreateDefaultPermission(
			ctx,
			&object_builder_service.CreateDefaultPermissionRequest{
				ProjectId: req.GetProjectId(),
				RoleId:    cast.ToString(roleDataData["guid"]),
			},
		)
		if err != nil {
			s.log.Error("!!!AddRole.ObjectBuilderService.CreateDefaultPermission--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

	case 3:
		result, err := services.GoItemService().Create(ctx, &nobs.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			s.log.Error("!!!AddRole.PostgresObjectBuilderService.Create--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		roleData, _ := helper.ConvertStructToResponse(result.Data)
		roleDataData := cast.ToStringMap(roleData["data"])
		_, err = services.GoObjectBuilderPermissionService().CreateDefaultPermission(
			ctx,
			&nobs.CreateDefaultPermissionRequest{
				ProjectId: req.GetProjectId(),
				RoleId:    cast.ToString(roleDataData["guid"]),
			},
		)
		if err != nil {
			s.log.Error("!!!AddRole.ObjectBuilderService.CreateDefaultPermission--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &pb.CommonMessage{
			TableSlug: result.TableSlug,
			Data:      result.Data,
		}, nil
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

	services, err := s.serviceNode.GetByNodeType(
		req.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	switch req.ResourceType {
	case 1:
		result, err = services.GetObjectBuilderServiceByType(req.NodeType).GetSingle(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetRoleById.ObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		result, err = services.PostgresObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
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

	services, err := s.serviceNode.GetByNodeType(
		req.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	switch req.ResourceType {
	case 1:
		result, err = services.GetObjectBuilderServiceByType(req.NodeType).GetList(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetRolesList.ObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		result, err := services.GoObjectBuilderService().GetList2(ctx, &nobs.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetRolesList.GoObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &pb.CommonMessage{
			TableSlug: result.TableSlug,
			Data:      result.Data,
		}, nil
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

	services, err := s.serviceNode.GetByNodeType(
		req.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	req.ProjectId = req.GetDbProjectId()

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!UpdateRole--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	switch req.ResourceType {
	case 1:
		result, err = services.GetObjectBuilderServiceByType(req.NodeType).Update(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!UpdateRole.ObjectBuilderService.Update--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		result, err = services.PostgresObjectBuilderService().Update(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
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

	services, err := s.serviceNode.GetByNodeType(
		req.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	switch req.ResourceType {
	case 1:
		result, err = services.GetObjectBuilderServiceByType(req.NodeType).GetSingle(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetRoleById.ObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		_, err = services.GetObjectBuilderServiceByType(req.NodeType).Delete(ctx, &pbObject.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetRoleById.ObjectBuilderService.Delete--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		result, err := services.GoItemService().GetSingle(ctx, &nobs.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetRoleById.PostgresObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		_, err = services.GoItemService().Delete(ctx, &nobs.CommonMessage{
			TableSlug: "role",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetRoleById.PostgresObjectBuilderService.Delete--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &pb.CommonMessage{
			TableSlug: result.TableSlug,
			Data:      result.Data,
		}, nil
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}
