package service

import (
	"context"
	"runtime"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	nobs "ucode/ucode_go_auth_service/genproto/new_object_builder_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"

	"github.com/opentracing/opentracing-go"
	"github.com/saidamir98/udevs_pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *clientService) V2CreateClientType(ctx context.Context, req *pb.V2CreateClientTypeRequest) (*pb.CommonMessage, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_client.V2CreateClientType")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2CreateClientType", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2CreateClientType", memoryUsed))
		}
	}()

	s.log.Info("---CreateClientType--->", logger.Any("req", req))
	var result *pbObject.CommonMessage

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

	services, err := s.serviceNode.GetByNodeType(req.ProjectId, req.NodeType)
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
		result, err := services.GoItemService().Create(ctx,
			&nobs.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetResourceEnvrironmentId(),
			})
		if err != nil {
			s.log.Error("!!!CreateClientType.PostgresObjectBuilderService.Create--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
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

func (s *clientService) V2GetClientTypeByID(ctx context.Context, req *pb.V2ClientTypePrimaryKey) (*pb.CommonMessage, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_client.V2GetClientTypeByID")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2GetClientTypeByID", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2GetClientTypeByID", memoryUsed))
		}
	}()

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
		result, err := services.GoItemService().GetSingle(ctx,
			&nobs.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetResourceEnvrironmentId(),
			})

		if err != nil {
			s.log.Error("!!!GetClientTypeByID.PostgresObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
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

func (s *clientService) V2GetClientTypeList(ctx context.Context, req *pb.V2GetClientTypeListRequest) (*pb.CommonMessage, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_client.V2GetClientTypeList")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2GetClientTypeList", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2GetClientTypeList", memoryUsed))
		}
	}()

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
		result, err = services.GetObjectBuilderServiceByType(req.NodeType).GetListSlim(ctx,
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
		result2, err := services.GoObjectBuilderService().GetList2(ctx,
			&nobs.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetResourceEnvrironmentId(),
			})

		if err != nil {
			s.log.Error("!!!GetClientTypeList.PostgresObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		result.Data = result2.Data
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *clientService) V2UpdateClientType(ctx context.Context, req *pb.V2UpdateClientTypeRequest) (*pb.CommonMessage, error) {
	s.log.Info("---UpdateClientType--->", logger.Any("req", req))

	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_client.V2UpdateClientType")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2UpdateClientType", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2UpdateClientType", memoryUsed))
		}
	}()

	var (
		result *pbObject.CommonMessage
	)
	requestToObjBuilderService := map[string]interface{}{
		"name":          req.Name,
		"confirm_by":    req.ConfirmBy.String(),
		"self_register": req.SelfRegister,
		"self_recorder": req.SelfRecover,
		// "project_id":      req.DbProjectId,
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
		result, err := services.GoItemService().Update(ctx,
			&nobs.CommonMessage{
				TableSlug: "client_type",
				Data:      structData,
				ProjectId: req.GetResourceEnvrironmentId(),
			})
		if err != nil {
			s.log.Error("!!!UpdateClientType.PostgresObjectBuilderService.Update--->", logger.Error(err))
			return nil, status.Error(codes.NotFound, err.Error())
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

func (s *clientService) V2DeleteClientType(ctx context.Context, req *pb.V2ClientTypePrimaryKey) (*emptypb.Empty, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_client.V2DeleteClientType")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2DeleteClientType", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2DeleteClientType", memoryUsed))
		}
	}()

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
		_, err = services.GoItemService().Delete(ctx,
			&nobs.CommonMessage{
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
