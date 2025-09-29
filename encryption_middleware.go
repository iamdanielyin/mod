package mod

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// EncryptedRequest 加密的请求格式
type EncryptedRequest struct {
	Data      string `json:"data"`      // Base64编码的加密数据
	Signature string `json:"signature"` // Base64编码的签名
	Mode      string `json:"mode"`      // 加密模式: symmetric/asymmetric
}

// EncryptedResponse 加密的响应格式
type EncryptedResponse struct {
	Data      string `json:"data"`      // Base64编码的加密数据
	Signature string `json:"signature"` // Base64编码的签名
	Mode      string `json:"mode"`      // 加密模式: symmetric/asymmetric
}

// EncryptionMiddleware 加解密中间件
func EncryptionMiddleware(app *App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		config := app.GetModConfig()
		if config == nil || !config.Encryption.Global.Enabled {
			return c.Next()
		}

		// 获取服务和分组名称
		serviceName := c.Params("service", "")
		groupName := ""

		// 检查是否需要加密
		if !CheckEncryption(config, serviceName, groupName) {
			return c.Next()
		}

		// 解密请求
		if err := decryptRequest(c, config); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Failed to decrypt request: %v", err))
		}

		// 继续处理
		if err := c.Next(); err != nil {
			return err
		}

		// 加密响应
		if err := encryptResponse(c, config); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to encrypt response: %v", err))
		}

		return nil
	}
}

// 解密请求
func decryptRequest(c *fiber.Ctx, config *ModConfig) error {
	var encReq EncryptedRequest
	if err := c.BodyParser(&encReq); err != nil {
		return err
	}

	// 验证签名
	if config.Encryption.Signature.Enabled {
		sigVerification := NewSignatureVerification(config)
		if sigVerification != nil {
			dataBytes, err := base64.StdEncoding.DecodeString(encReq.Data)
			if err != nil {
				return fmt.Errorf("failed to decode data for signature verification: %w", err)
			}

			signatureBytes, err := base64.StdEncoding.DecodeString(encReq.Signature)
			if err != nil {
				return fmt.Errorf("failed to decode signature: %w", err)
			}

			if err := sigVerification.Verify(dataBytes, signatureBytes); err != nil {
				return fmt.Errorf("signature verification failed: %w", err)
			}
		}
	}

	// 解密数据
	encryptedData, err := base64.StdEncoding.DecodeString(encReq.Data)
	if err != nil {
		return fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	var decryptedData []byte
	mode := encReq.Mode
	if mode == "" {
		mode = config.Encryption.Global.Mode
	}

	switch mode {
	case "symmetric":
		symEncryption, err := NewSymmetricEncryption(config)
		if err != nil {
			return fmt.Errorf("failed to create symmetric encryption: %w", err)
		}
		decryptedData, err = symEncryption.Decrypt(encryptedData)
		if err != nil {
			return fmt.Errorf("symmetric decryption failed: %w", err)
		}
	case "asymmetric":
		asymEncryption, err := NewAsymmetricEncryption(config)
		if err != nil {
			return fmt.Errorf("failed to create asymmetric encryption: %w", err)
		}
		decryptedData, err = asymEncryption.Decrypt(encryptedData)
		if err != nil {
			return fmt.Errorf("asymmetric decryption failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported encryption mode: %s", mode)
	}

	// 替换请求体
	c.Request().SetBody(decryptedData)

	return nil
}

// 加密响应
func encryptResponse(c *fiber.Ctx, config *ModConfig) error {
	originalBody := c.Response().Body()
	if len(originalBody) == 0 {
		return nil
	}

	mode := config.Encryption.Global.Mode
	var encryptedData []byte
	var err error

	switch mode {
	case "symmetric":
		symEncryption, err := NewSymmetricEncryption(config)
		if err != nil {
			return fmt.Errorf("failed to create symmetric encryption: %w", err)
		}
		encryptedData, err = symEncryption.Encrypt(originalBody)
		if err != nil {
			return fmt.Errorf("symmetric encryption failed: %w", err)
		}
	case "asymmetric":
		asymEncryption, err := NewAsymmetricEncryption(config)
		if err != nil {
			return fmt.Errorf("failed to create asymmetric encryption: %w", err)
		}
		encryptedData, err = asymEncryption.Encrypt(originalBody)
		if err != nil {
			return fmt.Errorf("asymmetric encryption failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported encryption mode: %s", mode)
	}

	// 生成签名
	var signature []byte
	if config.Encryption.Signature.Enabled {
		sigVerification := NewSignatureVerification(config)
		if sigVerification != nil {
			signature, err = sigVerification.Sign(encryptedData)
			if err != nil {
				return fmt.Errorf("failed to sign response: %w", err)
			}
		}
	}

	// 构造加密响应
	encResp := EncryptedResponse{
		Data:      base64.StdEncoding.EncodeToString(encryptedData),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Mode:      mode,
	}

	responseData, err := json.Marshal(encResp)
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted response: %w", err)
	}

	c.Response().SetBody(responseData)
	c.Set("Content-Type", "application/json")

	return nil
}
