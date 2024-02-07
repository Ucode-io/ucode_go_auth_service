package service

import (
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/storage"

	"github.com/saidamir98/udevs_pkg/logger"
)

type versionHistoryService struct {
	cfg         config.BaseConfig
	log         logger.LoggerI
	strg        storage.StorageI
	services    client.ServiceManagerI
	serviceNode ServiceNodesI

}
