package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// GetEoaNewsListByTypeId 根据类型ID获取新闻列表
func GetEoaNewsListByTypeId(c *gin.Context) {
	// 获取请求参数
	typeIdStr := c.Query("typeId")
	currentStr := c.Query("current")
	sizeStr := c.Query("size")

	// 参数验证
	if typeIdStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("类型ID不能为空", 400))
		return
	}

	// 转换参数
	typeId, err := strconv.Atoi(typeIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("类型ID格式错误", 400))
		return
	}

	// 设置默认分页参数
	current := 1
	size := 10

	// 解析分页参数
	if currentStr != "" {
		current, err = strconv.Atoi(currentStr)
		if err != nil || current < 1 {
			current = 1
		}
	}

	if sizeStr != "" {
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 1 {
			size = 10
		}
	}

	// 获取新闻列表
	news, total, err := models.GetNewsByTypeID(typeId, current, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取新闻列表失败", 500))
		return
	}

	// 构建响应数据
	var newsList []gin.H
	for _, item := range news {
		newsList = append(newsList, gin.H{
			"id":        item.ID,
			"title":     item.Title,
			"content":   item.Content,
			"summary":   item.Summary,
			"cover":     item.Cover,
			"author":    item.Author,
			"source":    item.Source,
			"typeId":    item.TypeID,
			"typeName":  item.TypeName,
			"publishAt": item.PublishAt.Format("2006-01-02 15:04:05"),
			"isTop":     item.IsTop,
			"viewCount": item.ViewCount,
		})
	}

	// 返回响应
	response := gin.H{
		"success": true,
		"message": "获取成功",
		"code":    0,
		"result": gin.H{
			"records":     newsList,
			"total":       total,
			"size":        size,
			"current":     current,
			"pages":       (total + int64(size) - 1) / int64(size),
			"searchCount": true,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetNewsDetail 获取新闻详情
func GetNewsDetail(c *gin.Context) {
	// 获取请求参数
	idStr := c.Param("id")

	// 参数验证
	if idStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("新闻ID不能为空", 400))
		return
	}

	// 转换参数
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("新闻ID格式错误", 400))
		return
	}

	// 获取新闻详情
	news, err := models.GetNewsDetail(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("新闻不存在", 404))
		return
	}

	// 构建响应数据
	response := gin.H{
		"success": true,
		"message": "获取成功",
		"code":    0,
		"result": gin.H{
			"id":        news.ID,
			"title":     news.Title,
			"content":   news.Content,
			"summary":   news.Summary,
			"cover":     news.Cover,
			"author":    news.Author,
			"source":    news.Source,
			"typeId":    news.TypeID,
			"typeName":  news.TypeName,
			"publishAt": news.PublishAt.Format("2006-01-02 15:04:05"),
			"isTop":     news.IsTop,
			"viewCount": news.ViewCount,
		},
	}

	c.JSON(http.StatusOK, response)
}
