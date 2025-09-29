package mod

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"golang.org/x/crypto/chacha20poly1305"
)

// EncryptionManager 加解密管理器
type EncryptionManager struct {
	config *ModConfig
}

// NewEncryptionManager 创建加解密管理器
func NewEncryptionManager(config *ModConfig) *EncryptionManager {
	return &EncryptionManager{
		config: config,
	}
}

// SymmetricEncryption 对称加密
type SymmetricEncryption struct {
	Algorithm string
	Key       []byte
}

// NewSymmetricEncryption 创建对称加密实例
func NewSymmetricEncryption(config *ModConfig) (*SymmetricEncryption, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	symConfig := config.Encryption.Symmetric

	var key []byte
	var err error

	if symConfig.KeyFile != "" {
		key, err = ioutil.ReadFile(symConfig.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file: %w", err)
		}
	} else if symConfig.Key != "" {
		key, err = base64.StdEncoding.DecodeString(symConfig.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to decode key: %w", err)
		}
	} else {
		return nil, errors.New("no key specified")
	}

	return &SymmetricEncryption{
		Algorithm: symConfig.Algorithm,
		Key:       key,
	}, nil
}

// Encrypt 对称加密
func (s *SymmetricEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	switch s.Algorithm {
	case "AES256-GCM":
		return s.encryptAESGCM(plaintext)
	case "ChaCha20-Poly1305":
		return s.encryptChaCha20Poly1305(plaintext)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", s.Algorithm)
	}
}

// Decrypt 对称解密
func (s *SymmetricEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	switch s.Algorithm {
	case "AES256-GCM":
		return s.decryptAESGCM(ciphertext)
	case "ChaCha20-Poly1305":
		return s.decryptChaCha20Poly1305(ciphertext)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", s.Algorithm)
	}
}

// AES-GCM 加密
func (s *SymmetricEncryption) encryptAESGCM(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.Key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// AES-GCM 解密
func (s *SymmetricEncryption) decryptAESGCM(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.Key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

// ChaCha20-Poly1305 加密
func (s *SymmetricEncryption) encryptChaCha20Poly1305(plaintext []byte) ([]byte, error) {
	if len(s.Key) != 32 {
		return nil, errors.New("ChaCha20-Poly1305 requires 32-byte key")
	}

	aead, err := chacha20poly1305.New(s.Key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// ChaCha20-Poly1305 解密
func (s *SymmetricEncryption) decryptChaCha20Poly1305(ciphertext []byte) ([]byte, error) {
	if len(s.Key) != 32 {
		return nil, errors.New("ChaCha20-Poly1305 requires 32-byte key")
	}

	aead, err := chacha20poly1305.New(s.Key)
	if err != nil {
		return nil, err
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aead.Open(nil, nonce, ciphertext, nil)
}

// AsymmetricEncryption 非对称加密
type AsymmetricEncryption struct {
	Algorithm  string
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// NewAsymmetricEncryption 创建非对称加密实例
func NewAsymmetricEncryption(config *ModConfig) (*AsymmetricEncryption, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	asymConfig := config.Encryption.Asymmetric

	var publicKey *rsa.PublicKey
	var privateKey *rsa.PrivateKey
	var err error

	// 读取公钥
	if asymConfig.PublicKeyFile != "" {
		publicKey, err = loadPublicKeyFromFile(asymConfig.PublicKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load public key from file: %w", err)
		}
	} else if asymConfig.PublicKey != "" {
		publicKey, err = parsePublicKeyFromPEM(asymConfig.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	// 读取私钥
	if asymConfig.PrivateKeyFile != "" {
		privateKey, err = loadPrivateKeyFromFile(asymConfig.PrivateKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key from file: %w", err)
		}
	} else if asymConfig.PrivateKey != "" {
		privateKey, err = parsePrivateKeyFromPEM(asymConfig.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	return &AsymmetricEncryption{
		Algorithm:  asymConfig.Algorithm,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

// Encrypt 非对称加密（使用公钥）
func (a *AsymmetricEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	if a.PublicKey == nil {
		return nil, errors.New("public key not available")
	}

	switch a.Algorithm {
	case "RSA-OAEP":
		return rsa.EncryptOAEP(sha256.New(), rand.Reader, a.PublicKey, plaintext, nil)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", a.Algorithm)
	}
}

// Decrypt 非对称解密（使用私钥）
func (a *AsymmetricEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	if a.PrivateKey == nil {
		return nil, errors.New("private key not available")
	}

	switch a.Algorithm {
	case "RSA-OAEP":
		return rsa.DecryptOAEP(sha256.New(), rand.Reader, a.PrivateKey, ciphertext, nil)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", a.Algorithm)
	}
}

// 辅助函数：从文件加载公钥
func loadPublicKeyFromFile(filename string) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parsePublicKeyFromPEM(string(data))
}

// 辅助函数：从PEM格式解析公钥
func parsePublicKeyFromPEM(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPub, nil
}

// 辅助函数：从文件加载私钥
func loadPrivateKeyFromFile(filename string) (*rsa.PrivateKey, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parsePrivateKeyFromPEM(string(data))
}

// 辅助函数：从PEM格式解析私钥
func parsePrivateKeyFromPEM(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试PKCS8格式
		privKey, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		rsaPriv, ok := privKey.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
		return rsaPriv, nil
	}

	return priv, nil
}

// SignatureVerification 签名验证
type SignatureVerification struct {
	Algorithm string
	Key       []byte
}

// NewSignatureVerification 创建签名验证实例
func NewSignatureVerification(config *ModConfig) *SignatureVerification {
	if config == nil {
		return nil
	}

	sigConfig := config.Encryption.Signature

	var key []byte

	if sigConfig.KeyFile != "" {
		keyData, err := ioutil.ReadFile(sigConfig.KeyFile)
		if err == nil {
			key = keyData
		}
	} else if sigConfig.Key != "" {
		keyData, err := base64.StdEncoding.DecodeString(sigConfig.Key)
		if err == nil {
			key = keyData
		} else {
			// 如果不是base64，直接使用原始字符串
			key = []byte(sigConfig.Key)
		}
	}

	return &SignatureVerification{
		Algorithm: sigConfig.Algorithm,
		Key:       key,
	}
}

// Sign 生成签名
func (s *SignatureVerification) Sign(data []byte) ([]byte, error) {
	switch s.Algorithm {
	case "HMAC-SHA256":
		return s.signHMAC(data), nil
	default:
		return nil, fmt.Errorf("unsupported signature algorithm: %s", s.Algorithm)
	}
}

// Verify 验证签名
func (s *SignatureVerification) Verify(data []byte, signature []byte) error {
	switch s.Algorithm {
	case "HMAC-SHA256":
		expectedSig := s.signHMAC(data)
		if !hmac.Equal(signature, expectedSig) {
			return errors.New("signature verification failed")
		}
		return nil
	default:
		return fmt.Errorf("unsupported signature algorithm: %s", s.Algorithm)
	}
}

// HMAC-SHA256 签名
func (s *SignatureVerification) signHMAC(data []byte) []byte {
	h := hmac.New(sha256.New, s.Key)
	h.Write(data)
	return h.Sum(nil)
}

// CheckEncryption 检查是否需要加密
func CheckEncryption(config *ModConfig, serviceName, groupName string) bool {
	if config == nil || !config.Encryption.Global.Enabled {
		return false
	}

	// 检查白名单
	for _, whiteService := range config.Encryption.Whitelist.Services {
		if whiteService == serviceName {
			return false
		}
	}

	for _, whiteGroup := range config.Encryption.Whitelist.Groups {
		if whiteGroup == groupName {
			return false
		}
	}

	// 检查服务级别配置
	if serviceConfig, exists := config.Encryption.Services[serviceName]; exists {
		return serviceConfig.Enabled
	}

	// 检查分组级别配置
	if groupConfig, exists := config.Encryption.Groups[groupName]; exists {
		return groupConfig.Enabled
	}

	// 返回全局配置
	return config.Encryption.Global.Enabled
}
