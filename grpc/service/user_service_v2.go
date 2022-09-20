package service

import (
	"context"
	"fmt"
	"regexp"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"

	"github.com/saidamir98/udevs_pkg/logger"
	"github.com/saidamir98/udevs_pkg/security"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *userService) V2CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CommonMessage, error) {
	s.log.Info("---CreateUser--->", logger.Any("req", req))

	if len(req.Password) < 6 {
		err := fmt.Errorf("password must not be less than 6 characters")
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, err
	}

	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	req.Password = hashedPassword

	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	email := emailRegex.MatchString(req.Email)
	if !email {
		err = fmt.Errorf("email is not valid")
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, err
	}

	phoneRegex := regexp.MustCompile(`^[+]?(\d{1,2})?[\s.-]?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}$`)
	phone := phoneRegex.MatchString(req.Phone)
	if !phone {
		err = fmt.Errorf("phone number is not valid")
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, err
	}

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
		TableSlug: "user",
		Data:      structData,
	})
	if err != nil {
		s.log.Error("!!!CreateUser.ObjectBuilderService.Create--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *userService) V2GetUserByID(ctx context.Context, req *pb.UserPrimaryKey) (*pb.CommonMessage, error) {
	s.log.Info("---GetUserByID--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!GetUserByID--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.services.ObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
		TableSlug: "user",
		Data:      structData,
	})
	if err != nil {
		s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *userService) V2GetUserList(ctx context.Context, req *pb.GetUserListRequest) (*pb.CommonMessage, error) {
	s.log.Info("---GetUserList--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!GetUserList--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.services.ObjectBuilderService().GetList(ctx, &pbObject.CommonMessage{
		TableSlug: "user",
		Data:      structData,
	})
	if err != nil {
		s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil

}

func (s *userService) V2UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.CommonMessage, error) {
	s.log.Info("---UpdateUser--->", logger.Any("req", req))

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!UpdateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = s.services.ObjectBuilderService().Update(ctx, &pbObject.CommonMessage{
		TableSlug: "user",
		Data:      structData,
	})
	if err != nil {
		s.log.Error("!!!UpdateUser.ObjectBuilderService.Update--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	email := emailRegex.MatchString(req.Email)
	if !email {
		err = fmt.Errorf("email is not valid")
		s.log.Error("!!!UpdateUser--->", logger.Error(err))
		return nil, err
	}

	phoneRegex := regexp.MustCompile(`^[+]?(\d{1,2})?[\s.-]?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}$`)
	phone := phoneRegex.MatchString(req.Phone)
	if !phone {
		err = fmt.Errorf("phone number is not valid")
		s.log.Error("!!!UpdateUser--->", logger.Error(err))
		return nil, err
	}

	if err != nil {
		s.log.Error("!!!UpdateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.services.ObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
		TableSlug: "user",
		Data:      structData,
	})
	if err != nil {
		s.log.Error("!!!UpdateUser.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonMessage{
		TableSlug: result.TableSlug,
		Data:      result.Data,
	}, nil
}

func (s *userService) V2DeleteUser(ctx context.Context, req *pb.UserPrimaryKey) (*emptypb.Empty, error) {
	s.log.Info("---DeleteUser--->", logger.Any("req", req))

	res := &emptypb.Empty{}

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!DeleteUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = s.services.ObjectBuilderService().Delete(ctx, &pbObject.CommonMessage{
		TableSlug: "user",
		Data:      structData,
	})
	if err != nil {
		s.log.Error("!!!DeleteUser.ObjectBuilderService.Delete--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res, nil
}
