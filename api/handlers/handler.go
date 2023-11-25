package handlers

import (
	"strconv"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/grpc/service"

	"github.com/saidamir98/udevs_pkg/logger"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	cfg         config.BaseConfig
	log         logger.LoggerI
	services    client.ServiceManagerI
	serviceNode service.ServiceNodesI
}

func NewHandler(cfg config.BaseConfig, log logger.LoggerI, svcs client.ServiceManagerI, serviceNode service.ServiceNodesI) Handler {
	return Handler{
		cfg:         cfg,
		log:         log,
		services:    svcs,
		serviceNode: serviceNode,
	}
}

func (h *Handler) GetProjectSrvc(c *gin.Context, projectId string, nodeType string) (client.SharedServiceManagerI, error) {
	if nodeType == config.ENTER_PRICE_TYPE {
		srvc, err := h.serviceNode.Get(projectId)
		if err != nil {
			return nil, err
		}

		return srvc, nil
	} else {
		srvc, err := h.serviceNode.Get(h.cfg.UcodeNamespace)
		if err != nil {
			return nil, err
		}

		return srvc, nil
	}
}

func (h *Handler) handleResponse(c *gin.Context, status http.Status, data interface{}) {
	switch code := status.Code; {
	case code < 300:
		h.log.Info(
			"---Response--->",
			logger.Int("code", status.Code),
			logger.String("status", status.Status),
			logger.Any("description", status.Description),
			logger.Any("data", data),
		)
	case code < 400:
		h.log.Warn(
			"!!!Response--->",
			logger.Int("code", status.Code),
			logger.String("status", status.Status),
			logger.Any("description", status.Description),
			logger.Any("data", data),
		)
	default:
		h.log.Error(
			"!!!Response--->",
			logger.Int("code", status.Code),
			logger.String("status", status.Status),
			logger.Any("description", status.Description),
			logger.Any("data", data),
		)
	}

	c.JSON(status.Code, http.Response{
		Status:      status.Status,
		Description: status.Description,
		Data:        data,
	})
}

func (h *Handler) getOffsetParam(c *gin.Context) (offset int, err error) {
	offsetStr := c.DefaultQuery("offset", h.cfg.DefaultOffset)
	return strconv.Atoi(offsetStr)
}

func (h *Handler) getLimitParam(c *gin.Context) (offset int, err error) {
	offsetStr := c.DefaultQuery("limit", h.cfg.DefaultLimit)
	return strconv.Atoi(offsetStr)
}
