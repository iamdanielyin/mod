package mod

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID   string                 `json:"user_id"`
	Username string                 `json:"username"`
	Email    string                 `json:"email"`
	Role     string                 `json:"role"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
	jwt.RegisteredClaims
}

// TokenResponse represents the token response structure
type TokenResponse struct {
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	AccessTokenExpiresIn  int64  `json:"access_token_expires_in"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
	TokenType             string `json:"token_type"`
}

// JWTManager manages JWT token operations
type JWTManager struct {
	app    *App
	config *ModConfig
	logger *logrus.Logger
}

// NewJWTManager creates a new JWT manager instance
func NewJWTManager(app *App) *JWTManager {
	return &JWTManager{
		app:    app,
		config: app.GetModConfig(),
		logger: app.logger,
	}
}

// IsEnabled checks if JWT is enabled in configuration
func (j *JWTManager) IsEnabled() bool {
	return j.config != nil && j.config.Token.JWT.Enabled
}

// GenerateTokens generates both access and refresh tokens
func (j *JWTManager) GenerateTokens(userID, username, email, role string, extra map[string]interface{}) (*TokenResponse, error) {
	if !j.IsEnabled() {
		return nil, errors.New("JWT is not enabled")
	}

	jwtConfig := j.config.Token.JWT
	if jwtConfig.SecretKey == "" {
		return nil, errors.New("JWT secret key is not configured")
	}

	now := time.Now()

	// Parse expiration durations
	accessExpire, err := time.ParseDuration(jwtConfig.ExpireDuration)
	if err != nil {
		j.logger.WithError(err).Warn("Invalid JWT expire_duration, using default 24h")
		accessExpire = 24 * time.Hour
	}

	refreshExpire, err := time.ParseDuration(jwtConfig.RefreshExpireDuration)
	if err != nil {
		j.logger.WithError(err).Warn("Invalid JWT refresh_expire_duration, using default 168h")
		refreshExpire = 168 * time.Hour
	}

	// Generate access token
	accessClaims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		Extra:    extra,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtConfig.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessExpire)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken, err := j.generateToken(accessClaims, jwtConfig.SecretKey, jwtConfig.Algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtConfig.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshExpire)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	refreshToken, err := j.generateToken(refreshClaims, jwtConfig.SecretKey, jwtConfig.Algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	response := &TokenResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresIn:  int64(accessExpire.Seconds()),
		RefreshTokenExpiresIn: int64(refreshExpire.Seconds()),
		TokenType:             "Bearer",
	}

	j.logger.WithFields(logrus.Fields{
		"user_id":                    userID,
		"username":                   username,
		"access_token_expires_in":    accessExpire,
		"refresh_token_expires_in":   refreshExpire,
	}).Info("JWT tokens generated successfully")

	return response, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	if !j.IsEnabled() {
		return nil, errors.New("JWT is not enabled")
	}

	jwtConfig := j.config.Token.JWT
	if jwtConfig.SecretKey == "" {
		return nil, errors.New("JWT secret key is not configured")
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		expectedMethod := j.getSigningMethod(jwtConfig.Algorithm)
		if token.Method != expectedMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtConfig.SecretKey), nil
	})

	if err != nil {
		j.logger.WithError(err).Debug("Token validation failed")
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Additional validation
	if claims.Issuer != jwtConfig.Issuer {
		return nil, errors.New("invalid token issuer")
	}

	j.logger.WithFields(logrus.Fields{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"subject":  claims.Subject,
	}).Debug("Token validated successfully")

	return claims, nil
}

// RefreshToken refreshes an access token using a refresh token
func (j *JWTManager) RefreshToken(refreshTokenString string) (*TokenResponse, error) {
	// Validate refresh token
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new tokens
	return j.GenerateTokens(claims.UserID, claims.Username, claims.Email, claims.Role, claims.Extra)
}

// RevokeToken revokes a token by adding it to the cache blacklist
func (j *JWTManager) RevokeToken(tokenString string) error {
	if !j.IsEnabled() {
		return errors.New("JWT is not enabled")
	}

	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("cannot revoke invalid token: %w", err)
	}

	// Add token to blacklist cache
	validationConfig := j.config.Token.Validation
	if validationConfig.Enabled {
		blacklistKey := validationConfig.CacheKeyPrefix + "blacklist:" + tokenString

		// Store in cache until token expires
		err := j.app.SetToken(blacklistKey, map[string]interface{}{
			"revoked_at": time.Now(),
			"user_id":    claims.UserID,
		})
		if err != nil {
			j.logger.WithError(err).Warn("Failed to add token to blacklist cache")
		}
	}

	j.logger.WithFields(logrus.Fields{
		"user_id":    claims.UserID,
		"expires_at": claims.ExpiresAt,
	}).Info("Token revoked successfully")

	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (j *JWTManager) IsTokenBlacklisted(tokenString string) bool {
	if !j.IsEnabled() {
		return false
	}

	validationConfig := j.config.Token.Validation
	if !validationConfig.Enabled {
		return false
	}

	blacklistKey := validationConfig.CacheKeyPrefix + "blacklist:" + tokenString
	_, err := j.app.GetTokenData(blacklistKey)
	return err == nil // Token exists in blacklist
}

// generateToken generates a JWT token with the specified claims
func (j *JWTManager) generateToken(claims *JWTClaims, secretKey, algorithm string) (string, error) {
	signingMethod := j.getSigningMethod(algorithm)
	token := jwt.NewWithClaims(signingMethod, claims)
	return token.SignedString([]byte(secretKey))
}

// getSigningMethod returns the appropriate signing method for the algorithm
func (j *JWTManager) getSigningMethod(algorithm string) jwt.SigningMethod {
	switch algorithm {
	case "HS256":
		return jwt.SigningMethodHS256
	case "HS384":
		return jwt.SigningMethodHS384
	case "HS512":
		return jwt.SigningMethodHS512
	default:
		j.logger.WithField("algorithm", algorithm).Warn("Unsupported JWT algorithm, using HS256")
		return jwt.SigningMethodHS256
	}
}

// ExtractTokenFromRequest extracts JWT token from HTTP request
func (j *JWTManager) ExtractTokenFromRequest(ctx *Context) string {
	// Try to get token from Authorization header
	authHeader := ctx.Get("Authorization")
	if authHeader != "" {
		// Check for Bearer token format
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			return authHeader[7:]
		}
		// Return as-is if not Bearer format
		return authHeader
	}

	// Try to get token from configured token keys
	if j.app.tokenKeys != nil {
		for _, key := range j.app.tokenKeys {
			if token := ctx.Get(key); token != "" {
				return token
			}
		}

		// Also check query parameters
		for _, key := range j.app.tokenKeys {
			if token := ctx.Query(key); token != "" {
				return token
			}
		}
	}

	return ""
}

// GetJWTManager returns the JWT manager for the app
func (app *App) GetJWTManager() *JWTManager {
	return NewJWTManager(app)
}