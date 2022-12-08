package handlers

import (
	"ucode/ucode_go_auth_service/api/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Saidakbar(c *gin.Context) {

	h.handleResponse(c, http.ServiceUnavailable, `{"hello": "world"}`)
}
