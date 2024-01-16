package service

import (
	"context"
	"fmt"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"

	"github.com/saidamir98/udevs_pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *clientService) V2CreateClientType(ctx context.Context, req *pb.V2CreateClientTypeRequest) (*pb.CommonMessage, error) {
	s.log.Info("---CreateClientType--->", logger.Any("req", req))
	var (
		result *pbObject.CommonMessage
	)

	requestToObjBuilderService := &pb.CreateClientTypeRequestToObjService{
		Name:         req.Name,
		ConfirmBy:    req.ConfirmBy.String(),
		SelfRegister: req.SelfRegister,
		SelfRecover:  req.SelfRecover,
		ProjectId:    req.DbProjectId,
		TableSlug:    req.GetTableSlug(),
	}

	structData, err := helper.ConvertRequestToSturct(requestToObjBuilderService)
	if err != nil {
		s.log.Error("!!!CreateClientType--->", logger.Error(err))
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
		result, err = services.GetObjectBuilderServiceByType(req.NodeType).Create(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.ResourceEnvrironmentId,
			})

		if err != nil {
			s.log.Error("!!!CreateClientType.ObjectBuilderService.Create--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	case 3:
		result, err = services.PostgresObjectBuilderService().Create(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetProjectId(),
			})

		if err != nil {
			s.log.Error("!!!CreateClientType.PostgresObjectBuilderService.Create--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2GetClientTypeByID(ctx context.Context, req *pb.V2ClientTypePrimaryKey) (*pb.CommonMessage, error) {
	s.log.Info("---GetClientTypeByID--->", logger.Any("req", req))

	var (
		result *pbObject.CommonMessage
	)
	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!GetClientTypeByID--->", logger.Error(err))
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
		result, err = services.GetObjectBuilderServiceByType(req.NodeType).GetSingle(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetResourceEnvrironmentId(),
			})

		if err != nil {
			s.log.Error("!!!GetClientTypeByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	case 3:
		result, err = services.PostgresObjectBuilderService().GetSingle(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetResourceEnvrironmentId(),
			})

		if err != nil {
			s.log.Error("!!!GetClientTypeByID.PostgresObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2GetClientTypeList(ctx context.Context, req *pb.V2GetClientTypeListRequest) (*pb.CommonMessage, error) {
	s.log.Info("---GetClientTypeList--->", logger.Any("req", req))
	result := &pbObject.CommonMessage{}

	if req.Limit == 0 {
		req.Limit = 1000
	}

	// @TODO limit offset error should fix
	if req.Limit == 0 {
		req.Limit = 1000
	}
	structReq := map[string]interface{}{
		"limit":  req.GetLimit(),
		"offset": req.GetOffset(),
		//"search": req.GetSearch(),
	}

	if req.Guids != nil {
		structReq["guid"] = req.Guids
	}

	structData, err := helper.ConvertRequestToSturct(structReq)

	if err != nil {
		s.log.Error("!!!GetClientTypeList--->", logger.Error(err))
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
		// c3c74e92-044c-4c30-ad51-b50eec3f49fa - staging
		// dc3b8f74-aa46-4101-b255-d6b82ac0db2d - production

		result, err = services.ObjectBuilderService().GetListSlim(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.ResourceEnvrironmentId,
			})
		if err != nil {
			s.log.Error("!!!----2-----GetClientTypeList.ObjectBuilderService.GetList--->", logger.Error(err))
			return &pb.CommonMessage{}, nil
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		fmt.Println("\n\n\n\n RESPONSE 222 -----> ", result)

		result, err = services.ObjectBuilderService().GetListSlim(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetResourceEnvrironmentId(),
			})
		if err != nil {
			s.log.Error("!!!GetClientTypeList.ObjectBuilderService.GetList--->", logger.Error(err))
			return &pb.CommonMessage{}, nil
			// return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	case 3:

		result, err = services.PostgresObjectBuilderService().GetList(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetResourceEnvrironmentId(),
			})

		if err != nil {
			s.log.Error("!!!GetClientTypeList.PostgresObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2UpdateClientType(ctx context.Context, req *pb.V2UpdateClientTypeRequest) (*pb.CommonMessage, error) {
	s.log.Info("---UpdateClientType--->", logger.Any("req", req))

	var (
		result *pbObject.CommonMessage
	)
	requestToObjBuilderService := map[string]interface{}{
		"name":            req.Name,
		"confirm_by":      req.ConfirmBy.String(),
		"self_register":   req.SelfRegister,
		"self_recorder":   req.SelfRecover,
		"project_id":      req.DbProjectId,
		"guid":            req.Guid,
		"client_type_ids": req.ClientPlatformIds,
		"table_slug":      req.TableSlug,
		"id":              req.Guid,
		"default_page":    req.DefaultPage,
	}

	structData, err := helper.ConvertRequestToSturct(requestToObjBuilderService)
	if err != nil {
		s.log.Error("!!!GetClientTypeList--->", logger.Error(err))
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
		result, err = services.GetObjectBuilderServiceByType(req.NodeType).Update(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetResourceEnvrironmentId(),
			})
		if err != nil {
			s.log.Error("!!!UpdateClientType.ObjectBuilderService.Update--->", logger.Error(err))
			return nil, status.Error(codes.NotFound, err.Error())
		}
	case 3:
		result, err = services.PostgresObjectBuilderService().Update(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetResourceEnvrironmentId(),
			})
		if err != nil {
			s.log.Error("!!!UpdateClientType.PostgresObjectBuilderService.Update--->", logger.Error(err))
			return nil, status.Error(codes.NotFound, err.Error())
		}
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2DeleteClientType(ctx context.Context, req *pb.V2ClientTypePrimaryKey) (*emptypb.Empty, error) {
	s.log.Info("---DeleteClientType--->", logger.Any("req", req))

	res := &emptypb.Empty{}

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!DeleteClientType--->", logger.Error(err))
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
		_, err = services.GetObjectBuilderServiceByType(req.NodeType).Delete(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetResourceEnvrironmentId(),
			})

		if err != nil {
			s.log.Error("!!!DeleteClientType.ObjectBuilderService.Delete--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		_, err = services.PostgresObjectBuilderService().Delete(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetResourceEnvrironmentId(),
			})

		if err != nil {
			s.log.Error("!!!DeleteClientType.PostgresObjectBuilderService.Delete--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return res, nil
}

// func (s *clientService) V2CreateClientPlatform(ctx context.Context, req *pb.CreateClientPlatformRequest) (*pb.CommonMessage, error) {
// 	s.log.Info("---CreateClientPlatform--->", logger.Any("req", req))

// 	structData, err := helper.ConvertRequestToSturct(req)
// 	if err != nil {
// 		s.log.Error("!!!CreateClientPlatform--->", logger.Error(err))
// 		return nil, status.Error(codes.InvalidArgument, err.Error())
// 	}
// 	var result *pbObject.CommonMessage
// 	switch req.ResourceType {
// 	case 1:
// 		result, err = s.services.ObjectBuilderService().Create(ctx,
// 			&pbObject.CommonMessage{
// 				TableSlug: "client_platform",
// 				Data:      structData,
// 				ProjectId: req.ProjectId,
// 			})

// 		if err != nil {
// 			s.log.Error("!!!CreateClientPlatform.ObjectBuilderService.Create--->", logger.Error(err))
// 			return nil, status.Error(codes.InvalidArgument, err.Error())
// 		}
// 	case 3:
// 		result, err = s.services.ObjectBuilderService().Create(ctx,
// 			&pbObject.CommonMessage{
// 				TableSlug: "client_platform",
// 				Data:      structData,
// 				ProjectId: req.ProjectId,
// 			})

// 		if err != nil {
// 			s.log.Error("!!!CreateClientPlatform.ObjectBuilderService.Create--->", logger.Error(err))
// 			return nil, status.Error(codes.InvalidArgument, err.Error())
// 		}
// 	}

// 	return &pb.CommonMessage{
// 		TableSlug: result.TableSlug,
// 		Data:      result.Data,
// 	}, nil
// }

// func (s *clientService) V2GetClientPlatformByID(ctx context.Context, req *pb.ClientPlatformPrimaryKey) (*pb.CommonMessage, error) {
// 	s.log.Info("---GetClientPlatformByID--->", logger.Any("req", req))

// 	structData, err := helper.ConvertRequestToSturct(req)
// 	if err != nil {
// 		s.log.Error("!!!GetClientPlatformById--->", logger.Error(err))
// 		return nil, status.Error(codes.InvalidArgument, err.Error())
// 	}
// 	var result *pbObject.CommonMessage
// 	switch req.ResourceType {
// 	case 1:
// 		result, err = s.services.ObjectBuilderService().GetSingle(ctx,
// 			&pbObject.CommonMessage{
// 				TableSlug: "client_platform",
// 				Data:      structData,
// 				ProjectId: req.ProjectId,
// 			})

// 		if err != nil {
// 			s.log.Error("!!!GetClientPlatformByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
// 			return nil, status.Error(codes.NotFound, err.Error())
// 		}
// 	case 3:
// 		result, err = s.services.PostgresObjectBuilderService().GetSingle(ctx,
// 			&pbObject.CommonMessage{
// 				TableSlug: "client_platform",
// 				Data:      structData,
// 				ProjectId: req.ProjectId,
// 			})

// 		if err != nil {
// 			s.log.Error("!!!GetClientPlatformByID.PostgresObjectBuilderService.GetSingle--->", logger.Error(err))
// 			return nil, status.Error(codes.NotFound, err.Error())
// 		}

// 	}

// 	return &pb.CommonMessage{
// 		TableSlug: result.TableSlug,
// 		Data:      result.Data,
// 	}, nil
// }

// func (s *clientService) V2GetClientPlatformByIDDetailed(ctx context.Context, req *pb.ClientPlatformPrimaryKey) (*pb.CommonMessage, error) {
// 	s.log.Info("---GetClientPlatformByID--->", logger.Any("req", req))

// 	structData, err := helper.ConvertRequestToSturct(req)
// 	if err != nil {
// 		s.log.Error("!!!GetClientPlatformByIDDetailed--->", logger.Error(err))
// 		return nil, status.Error(codes.InvalidArgument, err.Error())
// 	}
// 	var result *pbObject.CommonMessage
// 	switch req.ResourceType {
// 	case 1:
// 		result, err = s.services.ObjectBuilderService().GetSingle(ctx,
// 			&pbObject.CommonMessage{
// 				TableSlug: "client_platform",
// 				Data:      structData,
// 				ProjectId: config.UcodeDefaultProjectID,
// 			})

// 		if err != nil {
// 			s.log.Error("!!!GetClientPlatformByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
// 			return nil, status.Error(codes.Internal, err.Error())
// 		}
// 	case 3:
// 		result, err = s.services.ObjectBuilderService().GetSingle(ctx,
// 			&pbObject.CommonMessage{
// 				TableSlug: "client_platform",
// 				Data:      structData,
// 				ProjectId: config.UcodeDefaultProjectID,
// 			})

// 		if err != nil {
// 			s.log.Error("!!!GetClientPlatformByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
// 			return nil, status.Error(codes.Internal, err.Error())
// 		}
// 	}

// 	return &pb.CommonMessage{
// 		TableSlug: result.TableSlug,
// 		Data:      result.Data,
// 	}, nil
// }

// func (s *clientService) V2GetClientPlatformList(ctx context.Context, req *pb.GetClientPlatformListRequest) (*pb.CommonMessage, error) {
// 	s.log.Info("---GetClientPlatformList--->", logger.Any("req", req))

// 	// structData, err := helper.ConvertRequestToSturct(map[string]interface{}{
// 	// 	// "offset": req.Offset,
// 	// 	// "limit":  req.Limit,
// 	// 	// "search": req.Search,
// 	// })
// 	// if err != nil {
// 	// 	s.log.Error("!!!ClientPlatform--->", logger.Error(err))
// 	// }
// 	// var result *pbObject.CommonMessage
// 	// switch req.ResourceType {
// 	// case 1:
// 	// 	result, err = s.services.ObjectBuilderService().GetList(ctx,
// 	// 		&pbObject.CommonMessage{
// 	// 			TableSlug: "client_platform",
// 	// 			Data:      structData,
// 	// 			ProjectId: req.ProjectId,
// 	// 		})

// 	// 	if err != nil {
// 	// 		s.log.Error("!!!GetClientPlatformList.ObjectBuilderService.GetList--->", logger.Error(err))
// 	// 		return nil, status.Error(codes.Internal, err.Error())
// 	// 	}
// 	// case 3:
// 	// 	result, err = s.services.PostgresObjectBuilderService().GetList(ctx,
// 	// 		&pbObject.CommonMessage{
// 	// 			TableSlug: "client_platform",
// 	// 			Data:      structData,
// 	// 			ProjectId: req.ProjectId,
// 	// 		})

// 	// 	if err != nil {
// 	// 		s.log.Error("!!!GetClientPlatformList.PostgresObjectBuilderService.GetList--->", logger.Error(err))
// 	// 		return nil, status.Error(codes.Internal, err.Error())
// 	// 	}

// 	// }

// 	return &pb.CommonMessage{
// 		TableSlug: "client_platform",
// 		Data:      &structpb.Struct{},
// 	}, nil
// }

// func (s *clientService) V2UpdateClientPlatform(ctx context.Context, req *pb.UpdateClientPlatformRequest) (*pb.CommonMessage, error) {
// 	s.log.Info("---UpdateClientPlatform--->", logger.Any("req", req))

// 	structData, err := helper.ConvertRequestToSturct(req)
// 	if err != nil {
// 		s.log.Error("!!!UpdateClientPlatform--->", logger.Error(err))
// 		return nil, status.Error(codes.InvalidArgument, err.Error())
// 	}
// 	var result *pbObject.CommonMessage
// 	switch req.ResourceType {
// 	case 1:
// 		result, err = s.services.ObjectBuilderService().Update(ctx,
// 			&pbObject.CommonMessage{
// 				TableSlug: "client_platform",
// 				Data:      structData,
// 				ProjectId: config.UcodeDefaultProjectID,
// 			})

// 		if err != nil {
// 			s.log.Error("!!!UpdateClientPlatform.ObjectBuilderService.Update--->", logger.Error(err))
// 			return nil, status.Error(codes.InvalidArgument, err.Error())
// 		}
// 		_, err = s.services.ObjectBuilderService().ManyToManyAppend(ctx,
// 			&pbObject.ManyToManyMessage{
// 				TableFrom: "client_platform",
// 				TableTo:   "client_type",
// 				IdFrom:    req.Id,
// 				IdTo:      req.ClientTypeIds,
// 				ProjectId: config.UcodeDefaultProjectID,
// 			})
// 		if err != nil {
// 			s.log.Error("!!!UpdateClientType.ObjectBuilderService.ManyToManyAppend--->", logger.Error(err))
// 			return nil, status.Error(codes.NotFound, err.Error())
// 		}
// 	case 3:
// 		result, err = s.services.PostgresObjectBuilderService().Update(ctx,
// 			&pbObject.CommonMessage{
// 				TableSlug: "client_platform",
// 				Data:      structData,
// 				ProjectId: config.UcodeDefaultProjectID,
// 			})

// 		if err != nil {
// 			s.log.Error("!!!UpdateClientPlatform.PostgresObjectBuilderService.Update--->", logger.Error(err))
// 			return nil, status.Error(codes.InvalidArgument, err.Error())
// 		}
// 		_, err = s.services.ObjectBuilderService().ManyToManyAppend(ctx,
// 			&pbObject.ManyToManyMessage{
// 				TableFrom: "client_platform",
// 				TableTo:   "client_type",
// 				IdFrom:    req.Id,
// 				IdTo:      req.ClientTypeIds,
// 				ProjectId: config.UcodeDefaultProjectID,
// 			})
// 		if err != nil {
// 			s.log.Error("!!!UpdateClientType.ObjectBuilderService.ManyToManyAppend--->", logger.Error(err))
// 			return nil, status.Error(codes.NotFound, err.Error())
// 		}

// 	}

// 	_, err = s.services.ObjectBuilderService().ManyToManyAppend(ctx,
// 		&pbObject.ManyToManyMessage{
// 			TableFrom: "client_platform",
// 			TableTo:   "client_type",
// 			IdFrom:    req.Id,
// 			IdTo:      req.ClientTypeIds,
// 			ProjectId: config.UcodeDefaultProjectID,
// 		})
// 	if err != nil {
// 		s.log.Error("!!!UpdateClientType.ObjectBuilderService.ManyToManyAppend--->", logger.Error(err))
// 		return nil, status.Error(codes.NotFound, err.Error())
// 	}

// 	return &pb.CommonMessage{
// 		TableSlug: result.TableSlug,
// 		Data:      result.Data,
// 	}, nil
// }

// func (s *clientService) V2DeleteClientPlatform(ctx context.Context, req *pb.ClientPlatformPrimaryKey) (*emptypb.Empty, error) {
// 	s.log.Info("---DeleteClientPlatform--->", logger.Any("req", req))

// 	res := &emptypb.Empty{}
// 	structData, err := helper.ConvertRequestToSturct(req)
// 	if err != nil {
// 		s.log.Error("!!!DeleteClientPlatform--->", logger.Error(err))
// 		return nil, status.Error(codes.InvalidArgument, err.Error())
// 	}
// 	switch req.ResourceType {
// 	case 1:
// 		_, err = s.services.ObjectBuilderService().Delete(ctx,
// 			&pbObject.CommonMessage{
// 				TableSlug: "client_platform",
// 				Data:      structData,
// 				ProjectId: config.UcodeDefaultProjectID,
// 			})

// 		if err != nil {
// 			s.log.Error("!!!DeleteClientPlatform.ObjectBuilderService.Delete--->", logger.Error(err))
// 			return nil, status.Error(codes.Internal, err.Error())
// 		}
// 	case 3:

// 		_, err = s.services.PostgresObjectBuilderService().Delete(ctx,
// 			&pbObject.CommonMessage{
// 				TableSlug: "client_platform",
// 				Data:      structData,
// 				ProjectId: config.UcodeDefaultProjectID,
// 			})

// 		if err != nil {
// 			s.log.Error("!!!DeleteClientPlatform.PostgresObjectBuilderService.Delete--->", logger.Error(err))
// 			return nil, status.Error(codes.Internal, err.Error())
// 		}

// 	}

// 	return res, nil
// }
