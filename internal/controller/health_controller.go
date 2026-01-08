package controller

import (
	"gin-app-start/internal/common"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthCheck godoc
//
//	@Summary		Health check
//	@Description	Check if the service is running
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=object{status=string,message=string}}
//	@Router			/health [get]
func (ctrl *HealthController) HealthCheck() common.HandlerFunc {
	return func(c common.Context) {
		c.Payload(gin.H{
			"status":  "ok",
			"message": "service is running",
		})
	}
}
