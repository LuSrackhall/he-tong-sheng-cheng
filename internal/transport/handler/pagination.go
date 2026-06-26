package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

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
