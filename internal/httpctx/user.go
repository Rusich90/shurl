package httpctx

import "github.com/gin-gonic/gin"

func GetUserID(c *gin.Context) (string, bool) {
	v, ok := c.Get("userID")
	if !ok {
		return "", false
	}

	userID, ok := v.(string)
	return userID, ok
}
