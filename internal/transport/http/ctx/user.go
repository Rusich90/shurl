package ctx

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetUserID(c *gin.Context) (*uuid.UUID, bool) {
	v, ok := c.Get("userID")
	if !ok || v == nil {
		return nil, false
	}
	uid, ok := v.(*uuid.UUID)
	return uid, ok
}
