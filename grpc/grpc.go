package grpc

import (
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/grpc/service"
	"ucode/ucode_go_auth_service/storage"

	"github.com/saidamir98/udevs_pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func SetUpServer(cfg config.Config, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI) (grpcServer *grpc.Server) {
	grpcServer = grpc.NewServer()

	auth_service.RegisterClientServiceServer(grpcServer, service.NewClientService(cfg, log, strg, svcs))
	auth_service.RegisterPermissionServiceServer(grpcServer, service.NewPermissionService(cfg, log, strg, svcs))
	auth_service.RegisterUserServiceServer(grpcServer, service.NewUserService(cfg, log, strg, svcs))
	auth_service.RegisterSessionServiceServer(grpcServer, service.NewSessionService(cfg, log, strg, svcs))
	auth_service.RegisterIntegrationServiceServer(grpcServer, service.NewIntegrationService(cfg, log, strg, svcs))
	auth_service.RegisterEmailOtpServiceServer(grpcServer, service.NewEmailService(cfg, log, strg, svcs))
	reflection.Register(grpcServer)
	return
}
