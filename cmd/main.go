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
	cronjob "ucode/ucode_go_auth_service/pkg/cron"
	"ucode/ucode_go_auth_service/storage/postgres"

	"github.com/gin-gonic/gin"
	"github.com/golanguzb70/ratelimiter"
	"github.com/opentracing/opentracing-go"
	"github.com/saidamir98/udevs_pkg/logger"
	"github.com/uber/jaeger-client-go"
	jaeger_config "github.com/uber/jaeger-client-go/config"
)

func main() {
	var (
		loggerLevel string
		baseCfg     = config.BaseLoad()
	)

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

	jaegerCfg := &jaeger_config.Configuration{
		ServiceName: baseCfg.ServiceName,
		Sampler: &jaeger_config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaeger_config.ReporterConfig{
			LogSpans:           false,
			LocalAgentHostPort: baseCfg.JaegerHostPort,
		},
	}

	log := logger.NewLogger(baseCfg.ServiceName, loggerLevel)
	defer func() {
		_ = logger.Cleanup(log)
	}()

	tracer, closer, err := jaegerCfg.NewTracer(jaeger_config.Logger(jaeger.StdLogger))
	if err != nil {
		log.Error("ERROR: cannot init Jaeger", logger.Error(err))
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &ratelimiter.Config{
		RedisHost:    baseCfg.GetRequestRedisHost,
		RedisPort:    baseCfg.GetRequestRedisPort,
		JwtSignInKey: "jwt_sign_in_key",
		LeakyBuckets: config.RateLimitCfg,
	}

	limiter, err := ratelimiter.NewRateLimiter(cfg)
	if err != nil {
		log.Panic("Error creating rate limiter", logger.Error(err))
	}

	pgStore, err := postgres.NewPostgres(context.Background(), baseCfg)
	if err != nil {
		log.Panic("postgres.NewPostgres", logger.Error(err))
	}
	defer pgStore.CloseDB()

	// connection with auth and company
	baseSvcs, err := client.NewGrpcClients(ctx, baseCfg)
	if err != nil {
		log.Panic("--- U-code auth service and company service grpc client error: ", logger.Error(err))
	}

	serviceNodes := service.NewServiceNodes()

	// connection with shared services
	uConf := config.Load()
	grpcSvcs, err := client.NewSharedGrpcClients(ctx, uConf)
	if err != nil {
		log.Error("Error adding grpc client with base config. NewGrpcClients", logger.Error(err))
		return
	}

	err = serviceNodes.Add(grpcSvcs, baseCfg.UcodeNamespace)
	if err != nil {
		log.Error("Error adding company grpc client to serviceNode. ServiceNode", logger.Error(err))
		return
	}

	log.Info(" --- U-code company services --- added to serviceNodes!")

	projectServiceNodes, mapProjectConfs, err := service.EnterPriceProjectsGrpcSvcs(ctx, serviceNodes, baseSvcs, log)
	if err != nil {
		log.Error("Error maping company enter price projects to serviceNode. ServiceNode", logger.Error(err))
	}

	if projectServiceNodes == nil {
		projectServiceNodes = serviceNodes
	}

	if mapProjectConfs == nil {
		mapProjectConfs = make(map[string]config.Config)
	}

	mapProjectConfs[baseCfg.UcodeNamespace] = uConf
	projectServiceNodes.SetConfigs(mapProjectConfs)

	grpcServer := grpc.SetUpServer(baseCfg, log, pgStore, baseSvcs, projectServiceNodes)
	go func() {
		_ = cronjob.New(uConf, log, pgStore).RunJobs(context.Background())
	}()

	go func() {
		lis, err := net.Listen("tcp", baseCfg.AuthGRPCPort)
		if err != nil {
			log.Panic("net.Listen", logger.Error(err))
		}

		log.Info("GRPC: Server being started....", logger.String("port", baseCfg.AuthGRPCPort))

		if err := grpcServer.Serve(lis); err != nil {
			log.Panic("grpcServer.Serve", logger.Error(err))
		}
	}()

	h := handlers.NewHandler(baseCfg, log, baseSvcs, projectServiceNodes)

	r := api.SetUpRouter(h, baseCfg, tracer, limiter)

	_ = r.Run(baseCfg.HTTPPort)
}
