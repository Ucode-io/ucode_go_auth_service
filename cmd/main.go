package main

import (
	"context"
	"net"
	"ucode/ucode_go_auth_service/api"
	"ucode/ucode_go_auth_service/api/handlers"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/grpc"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/grpc/service"
	"ucode/ucode_go_auth_service/storage/postgres"

	"github.com/saidamir98/udevs_pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	baseCfg := config.BaseLoad()

	loggerLevel := logger.LevelDebug

	switch baseCfg.Environment {
	case config.DebugMode:
		loggerLevel = logger.LevelDebug
		gin.SetMode(gin.DebugMode)
	case config.TestMode:
		loggerLevel = logger.LevelDebug
		gin.SetMode(gin.TestMode)
	default:
		loggerLevel = logger.LevelInfo
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.NewLogger(baseCfg.ServiceName, loggerLevel)
	defer logger.Cleanup(log)

	pgStore, err := postgres.NewPostgres(context.Background(), baseCfg)
	if err != nil {
		log.Panic("postgres.NewPostgres", logger.Error(err))
	}
	defer pgStore.CloseDB()

	// connection with auth and company
	baseSvcs, err := client.NewGrpcClients(baseCfg)
	if err != nil {
		log.Panic("--- U-code auth service and company service grpc client error: ", logger.Error(err))
	}

	serviceNodes := service.NewServiceNodes()

	// connection with shared services
	uConf := config.Load()
	grpcSvcs, err := client.NewSharedGrpcClients(uConf)
	if err != nil {
		log.Error("Error adding grpc client with base config. NewGrpcClients", logger.Error(err))
		return
	}
	err = serviceNodes.Add(grpcSvcs, baseCfg.UcodeNamespace)
	if err != nil {
		log.Error("Error adding company grpc client to serviceNode. ServiceNode", logger.Error(err))
		return
	}
	log.Info(" --- U-code company services --- added to serviceNodes")

	projectServiceNodes, mapProjectConfs, err := service.EnterPriceProjectsGrpcSvcs(ctx, serviceNodes, baseSvcs, log)
	if err != nil {
		log.Error("Error maping company enter price projects to serviceNode. ServiceNode", logger.Error(err))
		return
	}
	mapProjectConfs[baseCfg.UcodeNamespace] = uConf
	projectServiceNodes.SetConfigs(mapProjectConfs)

	grpcServer := grpc.SetUpServer(baseCfg, log, pgStore, baseSvcs, projectServiceNodes)
	// log.Info(" --- U-code auth service and company service grpc client done --- ")

	go func() {
		lis, err := net.Listen("tcp", baseCfg.AuthGRPCPort)
		if err != nil {
			log.Panic("net.Listen", logger.Error(err))
		}

		log.Info("GRPC: Server being started...", logger.String("port", baseCfg.AuthGRPCPort))

		if err := grpcServer.Serve(lis); err != nil {
			log.Panic("grpcServer.Serve", logger.Error(err))
		}
	}()
	h := handlers.NewHandler(baseCfg, log, baseSvcs, projectServiceNodes)

	r := api.SetUpRouter(h, baseCfg)

	r.Run(baseCfg.HTTPPort)
}
