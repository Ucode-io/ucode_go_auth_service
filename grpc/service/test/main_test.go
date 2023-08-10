package service_test

import (
	"context"
	"log"
	"net"
	"os"
	"testing"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/grpc"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/pkg/logger"
	"ucode/ucode_go_auth_service/storage/postgres"

	"github.com/manveru/faker"
)

var (
	conf     config.Config
	fakeData *faker.Faker
)

func TestMain(m *testing.M) {
	conf = config.Load()
	newLogger := logger.NewLogger(conf.ServiceName, logger.LevelDebug)
	fakeData, _ = faker.New("en")

	pgStore, err := postgres.NewPostgres(context.Background(), conf)
	if err != nil {
		log.Fatal(err)
	}
	defer pgStore.CloseDB()

	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.SetUpServer(conf, newLogger, pgStore, svcs)

	go func() {
		lis, err := net.Listen("tcp", conf.AuthGRPCPort)
		if err != nil {
			log.Fatal(err)
		}

		newLogger.Info("GRPC: Server being started...", logger.String("port", conf.AuthGRPCPort))

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	os.Exit(m.Run())
}
