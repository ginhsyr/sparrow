package middleware

import (
	"Sparrow/internal/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RequireRole(allowedRoles ...model.RoleType) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "roles not found"})
			c.Abort()
			return
		}

		userRoles := roleVal.([]model.RoleType)
		roleMap := make(map[model.RoleType]bool)
		for _, r := range userRoles {
			roleMap[r] = true
		}

		for _, allowed := range allowedRoles {
			if roleMap[allowed] {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient role permissions"})
		c.Abort()
	}
}
