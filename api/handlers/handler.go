package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/grpc/service"
	"ucode/ucode_go_auth_service/pkg/util"

	"github.com/google/uuid"
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
	limitStr := c.DefaultQuery("limit", h.cfg.DefaultLimit)
	return strconv.Atoi(limitStr)
}

func (h *Handler) versionHistory(c *gin.Context, req *models.CreateVersionHistoryRequest) error {
	var (
		current  = map[string]interface{}{"data": req.Current}
		previous = map[string]interface{}{"data": req.Previous}
		request  = map[string]interface{}{"data": req.Request}
		response = map[string]interface{}{"data": req.Response}
		user     = req.UserInfo
	)

	sharedServiceManager, err := h.serviceNode.GetByNodeType(req.ProjectId, req.NodeType)
	if err != nil {
		h.log.Info("Error getting shared service manager", logger.Error(err))
		return err
	}

	if req.Current == nil {
		current["data"] = make(map[string]interface{})
	}
	if req.Previous == nil {
		previous["data"] = make(map[string]interface{})
	}
	if req.Request == nil {
		request["data"] = make(map[string]interface{})
	}
	if req.Response == nil {
		response["data"] = make(map[string]interface{})
	}

	if util.IsValidUUID(req.UserInfo) {
		info, err := h.services.UserService().GetUserByID(context.Background(), &auth_service.UserPrimaryKey{
			Id: req.UserInfo,
		})
		if err == nil {
			if info.Login != "" {
				user = info.Login
			} else {
				user = info.Phone
			}
		}
	}

	_, err = sharedServiceManager.VersionHistoryService().Create(
		context.Background(),
		&object_builder_service.CreateVersionHistoryRequest{
			Id:                uuid.NewString(),
			ProjectId:         req.ProjectId,
			ActionSource:      req.ActionSource,
			ActionType:        req.ActionType,
			Previus:           fromMapToString(previous),
			Current:           fromMapToString(current),
			UsedEnvrironments: req.UsedEnvironments,
			Date:              time.Now().Format("2006-01-02T15:04:05.000Z"),
			UserInfo:          user,
			Request:           fromMapToString(request),
			Response:          fromMapToString(response),
			ApiKey:            req.ApiKey,
			Type:              req.Type,
			TableSlug:         req.TableSlug,
		},
	)
	if err != nil {
		fmt.Println("=======================================================")
		log.Println(err)
		fmt.Println("=======================================================")
		return err
	}
	return nil
}

func fromMapToString(req map[string]interface{}) string {
	reqString, err := json.Marshal(req)
	if err != nil {
		return ""
	}
	return string(reqString)
}
