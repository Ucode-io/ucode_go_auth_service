package service

import (
	"context"
	"fmt"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"

	"github.com/saidamir98/udevs_pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *clientService) V2CreateClientPlatform(ctx context.Context, req *pb.CreateClientPlatformRequest) (*pb.CommonMessage, error) {
	s.log.Info("---CreateClientPlatform--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!CreateClientPlatform--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var result *pbObject.CommonMessage
	switch req.ResourceType {
	case 1:
		result, err = s.services.ObjectBuilderService().Create(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_platform",
				Data:      structData,
				ProjectId: req.ProjectId,
			})

		if err != nil {
			s.log.Error("!!!CreateClientPlatform.ObjectBuilderService.Create--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	case 3:
		result, err = s.services.ObjectBuilderService().Create(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_platform",
				Data:      structData,
				ProjectId: req.ProjectId,
			})

		if err != nil {
			s.log.Error("!!!CreateClientPlatform.ObjectBuilderService.Create--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2GetClientPlatformByID(ctx context.Context, req *pb.ClientPlatformPrimaryKey) (*pb.CommonMessage, error) {
	s.log.Info("---GetClientPlatformByID--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!GetClientPlatformById--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var result *pbObject.CommonMessage
	switch req.ResourceType {
	case 1:
		result, err = s.services.ObjectBuilderService().GetSingle(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_platform",
				Data:      structData,
				ProjectId: req.ProjectId,
			})

		if err != nil {
			s.log.Error("!!!GetClientPlatformByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.NotFound, err.Error())
		}
	case 3:
		result, err = s.services.PostgresObjectBuilderService().GetSingle(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_platform",
				Data:      structData,
				ProjectId: req.ProjectId,
			})

		if err != nil {
			s.log.Error("!!!GetClientPlatformByID.PostgresObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.NotFound, err.Error())
		}

	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2GetClientPlatformByIDDetailed(ctx context.Context, req *pb.ClientPlatformPrimaryKey) (*pb.CommonMessage, error) {
	s.log.Info("---GetClientPlatformByID--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!GetClientPlatformByIDDetailed--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var result *pbObject.CommonMessage
	switch req.ResourceType {
	case 1:
		result, err = s.services.ObjectBuilderService().GetSingle(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_platform",
				Data:      structData,
				ProjectId: config.UcodeDefaultProjectID,
			})

		if err != nil {
			s.log.Error("!!!GetClientPlatformByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		result, err = s.services.ObjectBuilderService().GetSingle(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_platform",
				Data:      structData,
				ProjectId: config.UcodeDefaultProjectID,
			})

		if err != nil {
			s.log.Error("!!!GetClientPlatformByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2GetClientPlatformList(ctx context.Context, req *pb.GetClientPlatformListRequest) (*pb.CommonMessage, error) {
	s.log.Info("---GetClientPlatformList--->", logger.Any("req", req))

	// structData, err := helper.ConvertRequestToSturct(map[string]interface{}{
	// 	// "offset": req.Offset,
	// 	// "limit":  req.Limit,
	// 	// "search": req.Search,
	// })
	// if err != nil {
	// 	s.log.Error("!!!ClientPlatform--->", logger.Error(err))
	// }
	// var result *pbObject.CommonMessage
	// switch req.ResourceType {
	// case 1:
	// 	result, err = s.services.ObjectBuilderService().GetList(ctx,
	// 		&pbObject.CommonMessage{
	// 			TableSlug: "client_platform",
	// 			Data:      structData,
	// 			ProjectId: req.ProjectId,
	// 		})

	// 	if err != nil {
	// 		s.log.Error("!!!GetClientPlatformList.ObjectBuilderService.GetList--->", logger.Error(err))
	// 		return nil, status.Error(codes.Internal, err.Error())
	// 	}
	// case 3:
	// 	result, err = s.services.PostgresObjectBuilderService().GetList(ctx,
	// 		&pbObject.CommonMessage{
	// 			TableSlug: "client_platform",
	// 			Data:      structData,
	// 			ProjectId: req.ProjectId,
	// 		})

	// 	if err != nil {
	// 		s.log.Error("!!!GetClientPlatformList.PostgresObjectBuilderService.GetList--->", logger.Error(err))
	// 		return nil, status.Error(codes.Internal, err.Error())
	// 	}

	// }

	return &pb.CommonMessage{
		TableSlug: "client_platform",
		Data:      &structpb.Struct{},
	}, nil
}

func (s *clientService) V2UpdateClientPlatform(ctx context.Context, req *pb.UpdateClientPlatformRequest) (*pb.CommonMessage, error) {
	s.log.Info("---UpdateClientPlatform--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!UpdateClientPlatform--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var result *pbObject.CommonMessage
	switch req.ResourceType {
	case 1:
		result, err = s.services.ObjectBuilderService().Update(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_platform",
				Data:      structData,
				ProjectId: config.UcodeDefaultProjectID,
			})

		if err != nil {
			s.log.Error("!!!UpdateClientPlatform.ObjectBuilderService.Update--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		_, err = s.services.ObjectBuilderService().ManyToManyAppend(ctx,
			&pbObject.ManyToManyMessage{
				TableFrom: "client_platform",
				TableTo:   "client_type",
				IdFrom:    req.Id,
				IdTo:      req.ClientTypeIds,
				ProjectId: config.UcodeDefaultProjectID,
			})
		if err != nil {
			s.log.Error("!!!UpdateClientType.ObjectBuilderService.ManyToManyAppend--->", logger.Error(err))
			return nil, status.Error(codes.NotFound, err.Error())
		}
	case 3:
		result, err = s.services.PostgresObjectBuilderService().Update(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_platform",
				Data:      structData,
				ProjectId: config.UcodeDefaultProjectID,
			})

		if err != nil {
			s.log.Error("!!!UpdateClientPlatform.PostgresObjectBuilderService.Update--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		_, err = s.services.ObjectBuilderService().ManyToManyAppend(ctx,
			&pbObject.ManyToManyMessage{
				TableFrom: "client_platform",
				TableTo:   "client_type",
				IdFrom:    req.Id,
				IdTo:      req.ClientTypeIds,
				ProjectId: config.UcodeDefaultProjectID,
			})
		if err != nil {
			s.log.Error("!!!UpdateClientType.ObjectBuilderService.ManyToManyAppend--->", logger.Error(err))
			return nil, status.Error(codes.NotFound, err.Error())
		}

	}

	_, err = s.services.ObjectBuilderService().ManyToManyAppend(ctx,
		&pbObject.ManyToManyMessage{
			TableFrom: "client_platform",
			TableTo:   "client_type",
			IdFrom:    req.Id,
			IdTo:      req.ClientTypeIds,
			ProjectId: config.UcodeDefaultProjectID,
		})
	if err != nil {
		s.log.Error("!!!UpdateClientType.ObjectBuilderService.ManyToManyAppend--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2DeleteClientPlatform(ctx context.Context, req *pb.ClientPlatformPrimaryKey) (*emptypb.Empty, error) {
	s.log.Info("---DeleteClientPlatform--->", logger.Any("req", req))

	res := &emptypb.Empty{}
	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!DeleteClientPlatform--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	switch req.ResourceType {
	case 1:
		_, err = s.services.ObjectBuilderService().Delete(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_platform",
				Data:      structData,
				ProjectId: config.UcodeDefaultProjectID,
			})

		if err != nil {
			s.log.Error("!!!DeleteClientPlatform.ObjectBuilderService.Delete--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:

		_, err = s.services.PostgresObjectBuilderService().Delete(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_platform",
				Data:      structData,
				ProjectId: config.UcodeDefaultProjectID,
			})

		if err != nil {
			s.log.Error("!!!DeleteClientPlatform.PostgresObjectBuilderService.Delete--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

	}

	return res, nil
}

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
	switch req.ResourceType {
	case 1:
		result, err = s.services.ObjectBuilderService().Create(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetProjectId(),
			})

		if err != nil {
			s.log.Error("!!!CreateClientType.ObjectBuilderService.Create--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	case 3:
		result, err = s.services.PostgresObjectBuilderService().Create(ctx,
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

	switch req.ResourceType {
	case 1:
		result, err = s.services.ObjectBuilderService().GetSingle(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetProjectId(),
			})

		if err != nil {
			s.log.Error("!!!GetClientTypeByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	case 3:
		result, err = s.services.PostgresObjectBuilderService().GetSingle(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetProjectId(),
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

//
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
	switch req.ResourceType {
	case 1:

		result, err = s.services.ObjectBuilderService().GetListSlim(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetProjectId(),
			})
		if err != nil {
			s.log.Error("!!!GetClientTypeList.ObjectBuilderService.GetList--->", logger.Error(err))
			return &pb.CommonMessage{}, nil
			// return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	case 3:

		result, err = s.services.PostgresObjectBuilderService().GetList(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetProjectId(),
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
		"mame":            req.Name,
		"confirm_by":      req.ConfirmBy.String(),
		"self_register":   req.SelfRegister,
		"self_recorder":   req.SelfRecover,
		"project_id":      req.DbProjectId,
		"guid":            req.Guid,
		"client_type_ids": req.ClientPlatformIds,
		"table_slug":      req.TableSlug,
	}

	structData, err := helper.ConvertRequestToSturct(requestToObjBuilderService)
	if err != nil {
		s.log.Error("!!!GetClientTypeList--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	switch req.ResourceType {
	case 1:
		result, err = s.services.ObjectBuilderService().Update(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetProjectId(),
			})
		if err != nil {
			s.log.Error("!!!UpdateClientType.ObjectBuilderService.Update--->", logger.Error(err))
			return nil, status.Error(codes.NotFound, err.Error())
		}
	case 3:
		result, err = s.services.PostgresObjectBuilderService().Update(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetProjectId(),
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
	switch req.ResourceType {
	case 1:
		_, err = s.services.ObjectBuilderService().Delete(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetProjectId(),
			})

		if err != nil {
			s.log.Error("!!!DeleteClientType.ObjectBuilderService.Delete--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		_, err = s.services.PostgresObjectBuilderService().Delete(ctx,
			&pbObject.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetProjectId(),
			})

		if err != nil {
			s.log.Error("!!!DeleteClientType.PostgresObjectBuilderService.Delete--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return res, nil
}

func (s *clientService) V2AddClient(ctx context.Context, req *pb.AddClientRequest) (*pb.CommonMessage, error) {
	s.log.Info("---AddClient--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!AddClient--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.services.ObjectBuilderService().Create(ctx,
		&pbObject.CommonMessage{
			TableSlug: "client",
			Data:      structData,
			ProjectId: config.UcodeDefaultProjectID,
		})

	if err != nil {
		s.log.Error("!!!AddClient.ObjectBuilderService.Create--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	convertedResult, err := helper.ConvertStructToResponse(result.Data)
	if err != nil {
		s.log.Error("!!!AddClient.ConvertStructToResponse--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	clientId := convertedResult["data"].(map[string]interface{})["guid"]

	loginRequest := &pb.LoginTableRequest{
		ObjectId:   req.LoginTable.ObjectId,
		TableSlug:  req.LoginTable.TableSlug,
		ViewFields: req.LoginTable.ViewFields,
		ClientId:   fmt.Sprintf("%v", clientId),
	}

	structLoginCreate, err := helper.ConvertRequestToSturct(loginRequest)
	if err != nil {
		s.log.Error("!!!AddClient.ConvertRequestToSturct--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = s.services.ObjectBuilderService().Create(ctx,
		&pbObject.CommonMessage{
			TableSlug: "login_table",
			Data:      structLoginCreate,
			ProjectId: config.UcodeDefaultProjectID,
		})
	if err != nil {
		s.log.Error("!!!AddClient.ObjectBuilderService.Create--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2GetClientList(ctx context.Context, req *pb.GetClientListRequest) (*pb.CommonMessage, error) {
	s.log.Info("---GetClientList--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!GetClientList--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.services.ObjectBuilderService().GetList(ctx,
		&pbObject.CommonMessage{
			TableSlug: "client",
			Data:      structData,
			ProjectId: config.UcodeDefaultProjectID,
		})

	if err != nil {
		s.log.Error("!!!GetClientList.ObjectBuilderService.Create--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2GetClientMatrix(ctx context.Context, req *pb.GetClientMatrixRequest) (*pb.CommonMessage, error) {
	s.log.Info("---GetClientMatrix--->", logger.Any("req", req))
	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!GetClientList--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result, err := s.services.ObjectBuilderService().GetList(ctx,
		&pbObject.CommonMessage{
			TableSlug: "client",
			Data:      structData,
			ProjectId: config.UcodeDefaultProjectID,
		})

	if err != nil {
		s.log.Error("!!!GetClientMatrix--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2UpdateClient(ctx context.Context, req *pb.UpdateClientRequest) (*pb.CommonMessage, error) {
	s.log.Info("---UpdateClient--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!UpdateClient--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.services.ObjectBuilderService().Update(ctx,
		&pbObject.CommonMessage{
			TableSlug: "client",
			Data:      structData,
			ProjectId: config.UcodeDefaultProjectID,
		})

	if err != nil {
		s.log.Error("!!!UpdateClient.ObjectBuilderService.Update--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2RemoveClient(ctx context.Context, req *pb.ClientPrimaryKey) (*pb.CommonMessage, error) {
	s.log.Info("---RemoveClient--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!RemoveClient--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resultClient, err := s.services.ObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
		TableSlug: "client",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!RemoveClient.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	_, err = s.services.ObjectBuilderService().Delete(ctx, &pbObject.CommonMessage{
		TableSlug: "client",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})

	if err != nil {
		s.log.Error("!!!RemoveClient.ObjectBuilderService.Delete--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: resultClient.TableSlug,
		Data:      resultClient.Data,
	}, nil

}

func (s *clientService) V2AddRelation(ctx context.Context, req *pb.AddRelationRequest) (*pb.Relation, error) {
	s.log.Info("---AddRelation--->", logger.Any("req", req))

	pKey, err := s.strg.Relation().Add(ctx, req)
	if err != nil {
		s.log.Error("!!!AddRelation--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return s.strg.Relation().GetByPK(ctx, pKey)
}

func (s *clientService) V2UpdateRelation(ctx context.Context, req *pb.UpdateRelationRequest) (*pb.Relation, error) {
	s.log.Info("---UpdateRelation--->", logger.Any("req", req))

	rowsAffected, err := s.strg.Relation().Update(ctx, req)

	if err != nil {
		s.log.Error("!!!UpdateRelation--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if rowsAffected <= 0 {
		return nil, status.Error(codes.InvalidArgument, "no rows were affected")
	}

	res, err := s.strg.Relation().GetByPK(ctx, &pb.RelationPrimaryKey{
		Id: req.Id,
	})

	if err != nil {
		s.log.Error("!!!UpdateRelation--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return res, err
}

func (s *clientService) V2RemoveRelation(ctx context.Context, req *pb.RelationPrimaryKey) (*pb.Relation, error) {
	s.log.Info("---RemoveRelation--->", logger.Any("req", req))

	res, err := s.strg.Relation().GetByPK(ctx, req)

	if err != nil {
		s.log.Error("!!!GetRelationPlatformByID--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	rowsAffected, err := s.strg.Relation().Remove(ctx, req)

	if err != nil {
		s.log.Error("!!!RemoveRelation--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if rowsAffected <= 0 {
		return nil, status.Error(codes.InvalidArgument, "no rows were affected")
	}

	return res, nil
}

func (s *clientService) V2AddUserInfoField(ctx context.Context, req *pb.AddUserInfoFieldRequest) (*pb.CommonMessage, error) {
	s.log.Info("---AddUserInfoField--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!AddUserInfoField--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result, err := s.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
		TableSlug: "user_info_field",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!AddUserInfoField.ObjectBuilderService.Creaete--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2UpdateUserInfoField(ctx context.Context, req *pb.UpdateUserInfoFieldRequest) (*pb.CommonMessage, error) {
	s.log.Info("---UpdateUserInfoField--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!UpdateUserInfoField--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result, err := s.services.ObjectBuilderService().Update(ctx, &pbObject.CommonMessage{
		TableSlug: "user_info_field",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!UpdateUserInfoField.ObjectBuilderService.Update--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2RemoveUserInfoField(ctx context.Context, req *pb.UserInfoFieldPrimaryKey) (*pb.CommonMessage, error) {
	s.log.Info("---RemoveUserInfoField--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!RemoveUserInfoField--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resultUserInfoField, err := s.services.ObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
		TableSlug: "user_info_field",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})
	if err != nil {
		s.log.Error("!!!RemoveUserInfoField.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	_, err = s.services.ObjectBuilderService().Delete(ctx, &pbObject.CommonMessage{
		TableSlug: "user_info_field",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})

	if err != nil {
		s.log.Error("!!!RemoveUserInfoField.ObjectBuilderService.Delete--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: resultUserInfoField.TableSlug,
		Data:      resultUserInfoField.Data,
	}, nil
}
