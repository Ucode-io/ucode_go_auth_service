package service

// import (
// 	"context"
// 	"ucode/ucode_go_auth_service/config"
// 	pb "ucode/ucode_go_auth_service/genproto/auth_service"
// 	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
// 	"ucode/ucode_go_auth_service/pkg/helper"

// 	"github.com/saidamir98/udevs_pkg/logger"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// 	"google.golang.org/protobuf/types/known/emptypb"
// )

// func (s *clientService) V2CreateConnections(ctx context.Context, req *pb.CreateClientPlatformRequest) (*pb.CommonMessage, error) {
// 	s.log.Info("---V2CreateConnections--->", logger.Any("req", req))

// 	structData, err := helper.ConvertRequestToSturct(req)
// 	if err != nil {
// 		s.log.Error("!!!V2CreateConnections--->", logger.Error(err))
// 		return nil, status.Error(codes.InvalidArgument, err.Error())
// 	}

// 	result, err := s.services.ObjectBuilderService().Create(ctx,
// 		&pbObject.CommonMessage{
// 			TableSlug: "connection",
// 			Data:      structData,
// 			ProjectId: req.ProjectId,
// 		})

// 	if err != nil {
// 		s.log.Error("!!!V2CreateConnections.ObjectBuilderService.Create--->", logger.Error(err))
// 		return nil, status.Error(codes.InvalidArgument, err.Error())
// 	}

// 	return &pb.CommonMessage{
// 		TableSlug: result.TableSlug,
// 		Data:      result.Data,
// 	}, nil
// }

// func (s *clientService) V2GetConnection(ctx context.Context, req *pb.ClientPlatformPrimaryKey) (*pb.CommonMessage, error) {
// 	s.log.Info("---V2GetConnection--->", logger.Any("req", req))

// 	structData, err := helper.ConvertRequestToSturct(req)
// 	if err != nil {
// 		s.log.Error("!!!V2GetConnection--->", logger.Error(err))
// 		return nil, status.Error(codes.InvalidArgument, err.Error())
// 	}

// 	result, err := s.services.ObjectBuilderService().GetSingle(ctx,
// 		&pbObject.CommonMessage{
// 			TableSlug: "connection",
// 			Data:      structData,
// 			ProjectId: req.ProjectId,
// 		})

// 	if err != nil {
// 		s.log.Error("!!!V2GetConnection.ObjectBuilderService.GetSingle--->", logger.Error(err))
// 		return nil, status.Error(codes.NotFound, err.Error())
// 	}

// 	return &pb.CommonMessage{
// 		TableSlug: result.TableSlug,
// 		Data:      result.Data,
// 	}, nil
// }

// func (s *clientService) V2GetConnnectionList(ctx context.Context, req *pb.GetClientPlatformListRequest) (*pb.CommonMessage, error) {
// 	s.log.Info("---V2GetConnnectionList--->", logger.Any("req", req))

// 	structData, err := helper.ConvertRequestToSturct(map[string]interface{}{
// 		// "offset": req.Offset,
// 		// "limit":  req.Limit,
// 		// "search": req.Search,
// 	})
// 	if err != nil {
// 		s.log.Error("!!!V2GetConnnectionList--->", logger.Error(err))
// 	}

// 	result, err := s.services.ObjectBuilderService().GetList(ctx,
// 		&pbObject.CommonMessage{
// 			TableSlug: "connection",
// 			Data:      structData,
// 			ProjectId: req.ProjectId,
// 		})

// 	if err != nil {
// 		s.log.Error("!!!V2GetConnnectionList.ObjectBuilderService.GetList--->", logger.Error(err))
// 		return nil, status.Error(codes.Internal, err.Error())
// 	}

// 	return &pb.CommonMessage{
// 		TableSlug: result.TableSlug,
// 		Data:      result.Data,
// 	}, nil
// }

// func (s *clientService) V2UpdateConnection(ctx context.Context, req *pb.UpdateClientPlatformRequest) (*pb.CommonMessage, error) {
// 	s.log.Info("---V2UpdateConnection--->", logger.Any("req", req))

// 	structData, err := helper.ConvertRequestToSturct(req)
// 	if err != nil {
// 		s.log.Error("!!!V2UpdateConnection--->", logger.Error(err))
// 		return nil, status.Error(codes.InvalidArgument, err.Error())
// 	}

// 	result, err := s.services.ObjectBuilderService().Update(ctx,
// 		&pbObject.CommonMessage{
// 			TableSlug: "connection",
// 			Data:      structData,
// 			ProjectId: config.UcodeDefaultProjectID,
// 		})

// 	if err != nil {
// 		s.log.Error("!!!V2UpdateConnection.ObjectBuilderService.Update--->", logger.Error(err))
// 		return nil, status.Error(codes.InvalidArgument, err.Error())
// 	}

// 	return &pb.CommonMessage{
// 		TableSlug: result.TableSlug,
// 		Data:      result.Data,
// 	}, nil
// }

// func (s *clientService) V2DeleteConnection(ctx context.Context, req *pb.ClientPlatformPrimaryKey) (*emptypb.Empty, error) {
// 	s.log.Info("---V2DeleteConnection--->", logger.Any("req", req))

// 	res := &emptypb.Empty{}
// 	structData, err := helper.ConvertRequestToSturct(req)
// 	if err != nil {
// 		s.log.Error("!!!V2DeleteConnection--->", logger.Error(err))
// 		return nil, status.Error(codes.InvalidArgument, err.Error())
// 	}

// 	_, err = s.services.ObjectBuilderService().Delete(ctx,
// 		&pbObject.CommonMessage{
// 			TableSlug: "connection",
// 			Data:      structData,
// 			ProjectId: config.UcodeDefaultProjectID,
// 		})

// 	if err != nil {
// 		s.log.Error("!!!V2DeleteConnection.ObjectBuilderService.Delete--->", logger.Error(err))
// 		return nil, status.Error(codes.Internal, err.Error())
// 	}

// 	return res, nil
// }
