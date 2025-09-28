package mod

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// JWTMiddleware creates a JWT authentication middleware
func JWTMiddleware(app *App) fiber.Handler {
	jwtManager := app.GetJWTManager()

	return func(c *fiber.Ctx) error {
		ctx := &Context{Ctx: c, logger: app.logger}

		// Skip if JWT is not enabled
		if !jwtManager.IsEnabled() {
			return c.Next()
		}

		// Extract token from request
		tokenString := jwtManager.ExtractTokenFromRequest(ctx)
		if tokenString == "" {
			app.logger.Debug("No JWT token found in request")
			return c.Status(401).JSON(NewErrorResponse(ctx, 401, "Missing authentication token"))
		}

		// Check if token is blacklisted
		if jwtManager.IsTokenBlacklisted(tokenString) {
			app.logger.WithField("token", tokenString[:10]+"...").Warn("Blacklisted token attempted access")
			return c.Status(401).JSON(NewErrorResponse(ctx, 401, "Token has been revoked"))
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			app.logger.WithError(err).Debug("JWT token validation failed")
			return c.Status(401).JSON(NewErrorResponse(ctx, 401, "Invalid authentication token"))
		}

		// Store claims in context for later use
		c.Locals("jwt_claims", claims)
		c.Locals("jwt_token", tokenString)
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)

		app.logger.WithFields(logrus.Fields{
			"user_id":  claims.UserID,
			"username": claims.Username,
			"role":     claims.Role,
		}).Debug("JWT authentication successful")

		return c.Next()
	}
}

// OptionalJWTMiddleware creates an optional JWT authentication middleware
// It validates JWT if present but doesn't fail if missing
func OptionalJWTMiddleware(app *App) fiber.Handler {
	jwtManager := app.GetJWTManager()

	return func(c *fiber.Ctx) error {
		ctx := &Context{Ctx: c, logger: app.logger}

		// Skip if JWT is not enabled
		if !jwtManager.IsEnabled() {
			return c.Next()
		}

		// Extract token from request
		tokenString := jwtManager.ExtractTokenFromRequest(ctx)
		if tokenString == "" {
			// No token is OK for optional middleware
			return c.Next()
		}

		// Check if token is blacklisted
		if jwtManager.IsTokenBlacklisted(tokenString) {
			app.logger.WithField("token", tokenString[:10]+"...").Warn("Blacklisted token attempted access")
			return c.Status(401).JSON(NewErrorResponse(ctx, 401, "Token has been revoked"))
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			app.logger.WithError(err).Debug("JWT token validation failed in optional middleware")
			// For optional middleware, we continue even if token is invalid
			return c.Next()
		}

		// Store claims in context for later use
		c.Locals("jwt_claims", claims)
		c.Locals("jwt_token", tokenString)
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)

		app.logger.WithFields(logrus.Fields{
			"user_id":  claims.UserID,
			"username": claims.Username,
			"role":     claims.Role,
		}).Debug("Optional JWT authentication successful")

		return c.Next()
	}
}

// RoleMiddleware creates a role-based authorization middleware
func RoleMiddleware(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := &Context{Ctx: c, logger: logrus.StandardLogger()}

		// Get JWT claims from context
		claims := ctx.GetJWTClaims()
		if claims == nil {
			return c.Status(401).JSON(NewErrorResponse(ctx, 401, "Authentication required"))
		}

		// Check if user has required role
		userRole := claims.Role
		hasRequiredRole := false

		for _, role := range requiredRoles {
			if userRole == role {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			ctx.GetLogger().WithFields(logrus.Fields{
				"user_id":        claims.UserID,
				"user_role":      userRole,
				"required_roles": requiredRoles,
			}).Warn("Access denied: insufficient permissions")

			return c.Status(403).JSON(NewErrorResponse(ctx, 403, "Insufficient permissions"))
		}

		return c.Next()
	}
}