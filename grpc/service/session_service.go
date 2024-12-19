package service

import (
	"context"
	"errors"

	"time"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/grpc/client"
	span "ucode/ucode_go_auth_service/pkg/jaeger"
	"ucode/ucode_go_auth_service/pkg/security"
	"ucode/ucode_go_auth_service/storage"

	"github.com/saidamir98/udevs_pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type sessionService struct {
	cfg         config.BaseConfig
	log         logger.LoggerI
	strg        storage.StorageI
	services    client.ServiceManagerI
	serviceNode ServiceNodesI
	pb.UnimplementedSessionServiceServer
}

func NewSessionService(cfg config.BaseConfig, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI, projectServiceNodes ServiceNodesI) *sessionService {
	return &sessionService{
		cfg:         cfg,
		log:         log,
		strg:        strg,
		services:    svcs,
		serviceNode: projectServiceNodes,
	}
}

func (s *sessionService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	res := &pb.LoginResponse{}

	if len(req.Username) < 6 {
		err := errors.New("invalid username")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(req.Password) < 6 {
		err := errors.New("invalid password")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.strg.User().GetByUsername(ctx, req.Username)
	if err != nil {
		err := errors.New("invalid username or password")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	hashType := user.GetHashType()
	if config.HashTypes[hashType] == 1 {
		match, err := security.ComparePassword(user.GetPassword(), req.Password)
		if err != nil {
			s.log.Error("!!!Login-->ComparePasswordArgon", logger.Error(err))
			return nil, err
		}
		if !match {
			err := errors.New("username or password is wrong")
			s.log.Error("!!!Login-->Wrong", logger.Error(err))
			return nil, err
		}

		go func() {
			hashedPassword, err := security.HashPasswordBcrypt(req.Password)
			if err != nil {
				s.log.Error("!!!V2UserResetPassword--->HashPasswordBcryptGo", logger.Error(err))
				return
			}
			err = s.strg.User().UpdatePassword(context.Background(), user.Id, hashedPassword)
			if err != nil {
				s.log.Error("!!!V2UserResetPassword--->UpdatePassword", logger.Error(err))
				return
			}
		}()
	} else if config.HashTypes[hashType] == 2 {
		match, err := security.ComparePasswordBcrypt(user.GetPassword(), req.Password)
		if err != nil {
			s.log.Error("!!!Login-->ComparePasswordBcrypt", logger.Error(err))
			return nil, err
		}
		if !match {
			err := errors.New("username or password is wrong")
			s.log.Error("!!!Login--->", logger.Error(err))
			return nil, err
		}
	} else {
		err := errors.New("hash type is not supported")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if user.Active < 0 {
		err := errors.New("user is not active")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if user.Active == 0 {
		err := errors.New("user hasn't been activated yet")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	expiresAt, err := time.Parse(config.DatabaseTimeLayout, user.ExpiresAt)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if expiresAt.Unix() < time.Now().Unix() {
		err := errors.New("user has been expired")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res.UserFound = true
	res.User = user

	companies, err := s.services.CompanyService().GetList(ctx, &pb.GetComapnyListRequest{UserId: user.Id})
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res.Companies = companies.Companies

	// TODO - Delete all old sessions & refresh token has this function too
	_, err = s.strg.Session().DeleteExpiredUserSessions(ctx, user.Id)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userSessionList, err := s.strg.Session().GetSessionListByUserID(ctx, user.Id)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res.Sessions = userSessionList.Sessions

	return res, nil
}

func (s *sessionService) GetList(ctx context.Context, req *pb.GetSessionListRequest) (*pb.GetSessionListResponse, error) {
	s.log.Info("!!!SessionGetList--->", logger.Any("req", req))

	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_session.SessionGetList", req)
	defer dbSpan.Finish()

	resp, err := s.strg.Session().GetList(ctx, req)
	if err != nil {
		s.log.Error("!!!GetList--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return resp, nil
}

func (s *sessionService) Delete(ctx context.Context, req *pb.SessionPrimaryKey) (*emptypb.Empty, error) {
	s.log.Info("!!!SessionDelete--->", logger.Any("req", req))

	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_session.Delete", req)
	defer dbSpan.Finish()

	_, err := s.strg.Session().Delete(ctx, req)
	if err != nil {
		s.log.Error("!!!Delete--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *sessionService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	var res = &pb.RefreshTokenResponse{}
	tokenInfo, err := security.ParseClaims(req.RefreshToken, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	session, err := s.strg.Session().GetByPK(ctx, &pb.SessionPrimaryKey{Id: tokenInfo.ID})
	if err != nil {
		s.log.Error("!!!RefreshToken session getbypk--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.ClientTypeId != "" {
		session.ClientTypeId = req.ClientTypeId
	}
	if req.ProjectId != "" {
		session.ProjectId = req.ProjectId
	}
	if req.RoleId != "" {
		session.RoleId = req.RoleId
	}
	if req.EnvId != "" {
		session.EnvId = req.EnvId
	}

	_, err = s.strg.Session().Update(ctx, &pb.UpdateSessionRequest{
		Id:               session.Id,
		ProjectId:        session.ProjectId,
		ClientPlatformId: session.ClientPlatformId,
		ClientTypeId:     session.ClientTypeId,
		UserId:           session.UserId,
		RoleId:           session.RoleId,
		Ip:               session.Ip,
		Data:             session.Data,
		ExpiresAt:        session.ExpiresAt,
		IsChanged:        session.IsChanged,
		EnvId:            session.EnvId,
	})
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{Id: session.UserId})
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if user.Active < 0 {
		err := errors.New("user is not active")
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// TODO - wrap in a function
	m := map[string]interface{}{
		"id":                 session.Id,
		"project_id":         session.ProjectId,
		"client_platform_id": session.ClientPlatformId,
		"client_type_id":     session.ClientTypeId,
		"user_id":            session.UserId,
		"role_id":            session.RoleId,
		"ip":                 session.Data,
		"data":               session.Data,
		"env_id":             session.EnvId,
	}

	accessToken, err := security.GenerateJWT(m, config.AccessTokenExpiresInTime, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	refreshToken, err := security.GenerateJWT(m, config.RefreshTokenExpiresInTime, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	res.Token = &pb.Token{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
		ExpiresAt:        session.ExpiresAt,
		RefreshInSeconds: int32(config.AccessTokenExpiresInTime.Seconds()),
	}

	return res, nil
}
