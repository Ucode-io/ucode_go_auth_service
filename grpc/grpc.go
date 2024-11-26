package grpc

import (
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/grpc/service"
	"ucode/ucode_go_auth_service/storage"

	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"github.com/saidamir98/udevs_pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func SetUpServer(cfg config.BaseConfig, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI, projectServiceNodes service.ServiceNodesI) (grpcServer *grpc.Server) {
	grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer())),
		grpc.StreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(opentracing.GlobalTracer())),
	)

	auth_service.RegisterClientServiceServer(grpcServer, service.NewClientService(cfg, log, strg, svcs, projectServiceNodes))
	auth_service.RegisterPermissionServiceServer(grpcServer, service.NewPermissionService(cfg, log, strg, svcs, projectServiceNodes))
	auth_service.RegisterUserServiceServer(grpcServer, service.NewUserService(cfg, log, strg, svcs, projectServiceNodes))
	auth_service.RegisterSessionServiceServer(grpcServer, service.NewSessionService(cfg, log, strg, svcs, projectServiceNodes))
	auth_service.RegisterCompanyServiceServer(grpcServer, service.NewCompanyService(cfg, log, strg, svcs, projectServiceNodes))
	auth_service.RegisterProjectServiceServer(grpcServer, service.NewProjectService(cfg, log, strg, svcs, projectServiceNodes))
	auth_service.RegisterApiKeysServer(grpcServer, service.NewApiKeysService(cfg, log, strg, svcs, projectServiceNodes))
	auth_service.RegisterAppleIdLoginServiceServer(grpcServer, service.NewAppleSettingsService(cfg, log, strg, svcs, projectServiceNodes))
	auth_service.RegisterRegisterServiceServer(grpcServer, service.NewRegisterService(cfg, log, strg, svcs, projectServiceNodes))
	auth_service.RegisterSyncUserServiceServer(grpcServer, service.NewSyncUserService(cfg, log, strg, svcs, projectServiceNodes))
	reflection.Register(grpcServer)
	auth_service.RegisterAuthPingServiceServer(grpcServer, service.NewPingService(log, projectServiceNodes))
	auth_service.RegisterApiKeyUsageServiceServer(grpcServer, service.NewApiKeyUsageService(cfg, log, strg, svcs, projectServiceNodes))

	return
}
