package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct {
}

func NewController() *Controller {
	return &Controller{}
}

type Health struct {
	Status string `json:"status" example:"OK"`
}

// @Summary      Responses with health status
// @Description  It responses with health status, if service is running
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  Health
// @Router       /health [get]
func (c *Controller) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Health{Status: "OK"})
}
