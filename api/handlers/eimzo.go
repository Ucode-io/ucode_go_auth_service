package handlers

import (
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/pkg/eimzo"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetChallenge(c *gin.Context) {
	resp, err := eimzo.GetChallenge(h.cfg)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}
