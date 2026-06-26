package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// getUintFromContext 从 gin.Context 中安全提取 uint 类型的值。
func getUintFromContext(c *gin.Context, key string) (uint, error) {
	val, ok := c.Get(key)
	if !ok {
		return 0, fmt.Errorf("missing context key: %s", key)
	}
	id, ok := val.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部错误"})
		return 0, fmt.Errorf("invalid type for context key: %s", key)
	}
	return id, nil
}

// parsePagination 从查询参数中解析分页参数，自动 clamp 到合法范围。
// offset 非负，limit 在 [1, maxLimit] 之间，非数字回退默认值。
func parsePagination(c *gin.Context, defaultLimit, maxLimit int) (offset, limit int) {
	offsetVal, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offsetVal < 0 {
		offsetVal = 0
	}

	limitVal, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limitVal < 1 {
		limitVal = defaultLimit
	}
	if limitVal > maxLimit {
		limitVal = maxLimit
	}

	return offsetVal, limitVal
}
