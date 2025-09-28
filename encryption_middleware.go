package mod

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// EncryptionMiddleware provides encryption and decryption middleware for services
func EncryptionMiddleware(app *App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if encryption is not enabled globally
		encManager := app.GetEncryptionManager()
		if !encManager.IsEnabled() {
			return c.Next()
		}

		// Extract service information from path
		serviceName := extractServiceName(c.Path(), app.cfg.ModConfig.App.ServicePathPrefix)
		if serviceName == "" {
			return c.Next() // Not a service endpoint
		}

		// Get service configuration
		service := findServiceByName(app.services, serviceName)
		if service == nil {
			return c.Next() // Service not found
		}

		// Check if encryption is enabled for this service
		if !encManager.IsServiceEnabled(serviceName, service.Group) {
			return c.Next()
		}

		// Get encryption mode and algorithm
		mode := encManager.GetEncryptionMode(serviceName, service.Group)
		algorithm := encManager.GetEncryptionAlgorithm(serviceName, service.Group)

		ctx := &Context{Ctx: c, logger: app.logger}

		// Process request (decrypt and verify signature)
		if err := processEncryptedRequest(ctx, app.cfg.ModConfig, mode, algorithm); err != nil {
			app.logger.WithFields(logrus.Fields{
				"service":   serviceName,
				"group":     service.Group,
				"mode":      mode,
				"algorithm": algorithm,
				"error":     err.Error(),
				"rid":       ctx.GetRequestID(),
			}).Error("Failed to process encrypted request")

			return c.Status(400).JSON(NewErrorResponse(ctx, 400, "Decryption failed", err.Error()))
		}

		return c.Next()
	}
}

// processEncryptedRequest handles the decryption and signature verification of the request
func processEncryptedRequest(ctx *Context, config *ModConfig, mode, algorithm string) error {
	// Parse the encrypted request body
	var encryptedReq EncryptedRequest
	if err := ctx.BodyParser(&encryptedReq); err != nil {
		return fmt.Errorf("failed to parse encrypted request: %w", err)
	}

	// Decode the encrypted data and signature
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedReq.Data)
	if err != nil {
		return fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	signature, err := base64.StdEncoding.DecodeString(encryptedReq.Signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Verify signature first (before decryption)
	if config.Encryption.Signature.Enabled {
		sigVerifier := NewSignatureVerification(config)
		if err := sigVerifier.Verify(encryptedData, signature); err != nil {
			return fmt.Errorf("signature verification failed: %w", err)
		}
	}

	// Decrypt the data
	var decryptedData []byte
	switch mode {
	case "symmetric":
		symEncryption := NewSymmetricEncryption(config)
		decryptedData, err = symEncryption.Decrypt(encryptedData)
		if err != nil {
			return fmt.Errorf("symmetric decryption failed: %w", err)
		}
	case "asymmetric":
		asymEncryption := NewAsymmetricEncryption(config)
		decryptedData, err = asymEncryption.Decrypt(encryptedData)
		if err != nil {
			return fmt.Errorf("asymmetric decryption failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported encryption mode: %s", mode)
	}

	// Replace the request body with decrypted data
	ctx.Request().SetBody(decryptedData)
	ctx.Request().Header.Set("Content-Type", "application/json")

	return nil
}

// EncryptedRequest represents the structure of an encrypted request
type EncryptedRequest struct {
	Data      string `json:"data"`      // Base64 encoded encrypted data
	Signature string `json:"signature"` // Base64 encoded signature
	Algorithm string `json:"algorithm"` // Encryption algorithm used
	Mode      string `json:"mode"`      // Encryption mode (symmetric/asymmetric)
}

// EncryptedResponse represents the structure of an encrypted response
type EncryptedResponse struct {
	Data      string `json:"data"`      // Base64 encoded encrypted data
	Signature string `json:"signature"` // Base64 encoded signature
	Algorithm string `json:"algorithm"` // Encryption algorithm used
	Mode      string `json:"mode"`      // Encryption mode (symmetric/asymmetric)
}

// EncryptResponse encrypts a response based on the service configuration
func EncryptResponse(app *App, serviceName, groupName string, data []byte) (*EncryptedResponse, error) {
	encManager := app.GetEncryptionManager()
	if !encManager.IsServiceEnabled(serviceName, groupName) {
		return nil, fmt.Errorf("encryption not enabled for service %s", serviceName)
	}

	mode := encManager.GetEncryptionMode(serviceName, groupName)
	algorithm := encManager.GetEncryptionAlgorithm(serviceName, groupName)
	config := app.cfg.ModConfig

	// Encrypt the data
	var encryptedData []byte
	var err error

	switch mode {
	case "symmetric":
		symEncryption := NewSymmetricEncryption(config)
		encryptedData, err = symEncryption.Encrypt(data)
		if err != nil {
			return nil, fmt.Errorf("symmetric encryption failed: %w", err)
		}
	case "asymmetric":
		asymEncryption := NewAsymmetricEncryption(config)
		encryptedData, err = asymEncryption.Encrypt(data)
		if err != nil {
			return nil, fmt.Errorf("asymmetric encryption failed: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported encryption mode: %s", mode)
	}

	// Create signature if enabled
	var signature []byte
	if config.Encryption.Signature.Enabled {
		sigVerifier := NewSignatureVerification(config)
		signature, err = sigVerifier.Sign(encryptedData)
		if err != nil {
			return nil, fmt.Errorf("signing failed: %w", err)
		}
	}

	return &EncryptedResponse{
		Data:      base64.StdEncoding.EncodeToString(encryptedData),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Algorithm: algorithm,
		Mode:      mode,
	}, nil
}

// extractServiceName extracts the service name from the request path
func extractServiceName(path, servicePathPrefix string) string {
	if len(path) <= len(servicePathPrefix)+1 {
		return ""
	}

	if path[:len(servicePathPrefix)] != servicePathPrefix {
		return ""
	}

	// Remove service path prefix and leading slash
	servicePath := path[len(servicePathPrefix):]
	if servicePath[0] == '/' {
		servicePath = servicePath[1:]
	}

	return servicePath
}

// findServiceByName finds a service by its name in the services list
func findServiceByName(services []Service, name string) *Service {
	for _, service := range services {
		if service.Name == name {
			return &service
		}
	}
	return nil
}

// UseEncryption enables encryption middleware for all service routes
func (app *App) UseEncryption() {
	app.Use(EncryptionMiddleware(app))
}

// EncryptServiceResponse encrypts a service response if encryption is enabled
func (app *App) EncryptServiceResponse(serviceName, groupName string, response interface{}) (interface{}, error) {
	encManager := app.GetEncryptionManager()
	if !encManager.IsServiceEnabled(serviceName, groupName) {
		return response, nil // No encryption needed
	}

	// Marshal the response to JSON
	responseData, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	// Encrypt the response
	encryptedResp, err := EncryptResponse(app, serviceName, groupName, responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt response: %w", err)
	}

	return encryptedResp, nil
}