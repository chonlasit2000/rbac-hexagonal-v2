package http

import (
	"strings"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/config"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/port"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Factory function เพื่อสร้าง Middleware
func NewRBACMiddleware(cfg *config.Config, rbacSvc port.RBACService) func(perm string) fiber.Handler {
	return func(requiredPerm string) fiber.Handler {
		return func(c *fiber.Ctx) error {
			// 1. ดึง Token จาก Header
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing Authorization header"})
			}
			tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

			// 2. Parse Token
			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				return []byte(cfg.Server.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Token"})
			}

			// 3. ดึง User ID จาก Claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Claims"})
			}
			userID := claims["user_id"].(string)

			// 4. เช็คสิทธิ์กับ RBAC Service
			allow, err := rbacSvc.CheckAccess(userID, requiredPerm)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Authorization failed"})
			}

			if !allow {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access Denied: You don't have permission " + requiredPerm})
			}

			// ผ่านฉลุย
			return c.Next()
		}
	}
}
