package middleware

import (
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"

	"html2pdf-auth-service/internal/domain"
)

type TokenStore interface {
	Ready() bool
	Validate(token string) bool
}

func OptionalAPIKeyAuth(tokens TokenStore) fiber.Handler {
	return keyauth.New(keyauth.Config{
		KeyLookup:  "header:X-API-Key",
		ContextKey: "api_key",
		Validator: func(c *fiber.Ctx, key string) (bool, error) {
			trimmed := strings.TrimSpace(key)
			if trimmed != key {
				key = trimmed
			}
			if !tokens.Ready() {
				log.Printf("auth reject reason=token_store_not_ready method=%s path=%s", c.Method(), c.Path())
				return false, domain.ErrTokenStoreNotReady
			}
			if !tokens.Validate(key) {
				log.Printf("auth reject reason=invalid_key key=%s method=%s path=%s", redactToken(key), c.Method(), c.Path())
				return false, domain.ErrInvalidAPIKey
			}
			log.Printf("auth allow key=%s method=%s path=%s", redactToken(key), c.Method(), c.Path())
			return true, nil
		},
		// Missing key = public access. Also allow OPTIONS preflight.
		Next: func(c *fiber.Ctx) bool {
			if c.Method() == fiber.MethodOptions || c.Get("X-API-Key") == "" {
				log.Printf("auth allow reason=public method=%s path=%s", c.Method(), c.Path())
				return true
			}
			return false
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status := fiber.StatusUnauthorized
			if err == nil {
				err = fiber.ErrUnauthorized
			}
			if errors.Is(err, domain.ErrTokenStoreNotReady) {
				status = fiber.StatusServiceUnavailable
			}
			log.Printf("auth error status=%d message=%s method=%s path=%s", status, err.Error(), c.Method(), c.Path())
			return c.Status(status).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    status,
					"message": err.Error(),
				},
			})
		},
	})
}

func redactToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
