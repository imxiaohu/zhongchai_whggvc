package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// GetSchoolServerStatus 获取学校服务器状态
func GetSchoolServerStatus(c *gin.Context) {
	healthService := services.GetSchoolHealthCheckService()
	status := healthService.GetStatus()

	response := map[string]interface{}{
		"isAlive":      status.IsAlive,
		"lastCheck":    status.LastCheck.Format("2006-01-02 15:04:05"),
		"lastAlive":    status.LastAlive.Format("2006-01-02 15:04:05"),
		"errorCount":   status.ErrorCount,
		"errorMsg":     status.ErrorMsg,
		"responseTime": status.ResponseTime,
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(response))
}

// ForceCheckSchoolServer 强制检查学校服务器状态
func ForceCheckSchoolServer(c *gin.Context) {
	healthService := services.GetSchoolHealthCheckService()
	healthService.ForceCheck()

	c.JSON(http.StatusOK, utils.NewSuccessResponse("健康检查已触发"))
}

// GetMaintenanceInfo 获取维护信息
func GetMaintenanceInfo(c *gin.Context) {
	healthService := services.GetSchoolHealthCheckService()
	maintenanceInfo := healthService.GetMaintenanceMessage()

	c.JSON(http.StatusOK, utils.NewSuccessResponse(maintenanceInfo))
}
