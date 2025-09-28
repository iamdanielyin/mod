package mod

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
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
	"math/big"

	"golang.org/x/crypto/chacha20poly1305"
)

// EncryptionManager manages encryption and decryption operations
type EncryptionManager struct {
	app    *App
	config *ModConfig
}

// NewEncryptionManager creates a new encryption manager instance
func NewEncryptionManager(app *App) *EncryptionManager {
	return &EncryptionManager{
		app:    app,
		config: app.GetModConfig(),
	}
}

// IsEnabled checks if encryption is enabled globally
func (e *EncryptionManager) IsEnabled() bool {
	return e.config != nil && e.config.Encryption.Global.Enabled
}

// IsServiceEnabled checks if encryption is enabled for a specific service
func (e *EncryptionManager) IsServiceEnabled(serviceName, groupName string) bool {
	if !e.IsEnabled() {
		return false
	}

	// Check whitelist first
	for _, whitelistService := range e.config.Encryption.Whitelist.Services {
		if whitelistService == serviceName {
			return false
		}
	}
	for _, whitelistGroup := range e.config.Encryption.Whitelist.Groups {
		if whitelistGroup == groupName {
			return false
		}
	}

	// Check service-specific configuration
	if serviceConfig, exists := e.config.Encryption.Services[serviceName]; exists {
		return serviceConfig.Enabled
	}

	// Check group-specific configuration
	if groupConfig, exists := e.config.Encryption.Groups[groupName]; exists {
		return groupConfig.Enabled
	}

	// Fall back to global configuration
	return e.config.Encryption.Global.Enabled
}

// GetEncryptionMode returns the encryption mode for a service
func (e *EncryptionManager) GetEncryptionMode(serviceName, groupName string) string {
	// Check service-specific configuration
	if serviceConfig, exists := e.config.Encryption.Services[serviceName]; exists && serviceConfig.Mode != "" {
		return serviceConfig.Mode
	}

	// Check group-specific configuration
	if groupConfig, exists := e.config.Encryption.Groups[groupName]; exists && groupConfig.Mode != "" {
		return groupConfig.Mode
	}

	// Fall back to global configuration
	return e.config.Encryption.Global.Mode
}

// GetEncryptionAlgorithm returns the encryption algorithm for a service
func (e *EncryptionManager) GetEncryptionAlgorithm(serviceName, groupName string) string {
	// Check service-specific configuration
	if serviceConfig, exists := e.config.Encryption.Services[serviceName]; exists && serviceConfig.Algorithm != "" {
		return serviceConfig.Algorithm
	}

	// Check group-specific configuration
	if groupConfig, exists := e.config.Encryption.Groups[groupName]; exists && groupConfig.Algorithm != "" {
		return groupConfig.Algorithm
	}

	// Fall back to global configuration
	return e.config.Encryption.Global.Algorithm
}

// SymmetricEncryption handles symmetric encryption operations
type SymmetricEncryption struct {
	config *ModConfig
}

// NewSymmetricEncryption creates a new symmetric encryption instance
func NewSymmetricEncryption(config *ModConfig) *SymmetricEncryption {
	return &SymmetricEncryption{config: config}
}

// getKey returns the encryption key from configuration
func (s *SymmetricEncryption) getKey() ([]byte, error) {
	if s.config.Encryption.Symmetric.Key != "" {
		// Decode base64 key
		return base64.StdEncoding.DecodeString(s.config.Encryption.Symmetric.Key)
	}

	if s.config.Encryption.Symmetric.KeyFile != "" {
		// Read key from file
		keyData, err := ioutil.ReadFile(s.config.Encryption.Symmetric.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file: %w", err)
		}

		// Try to decode as base64, if fails use raw data
		if decoded, err := base64.StdEncoding.DecodeString(string(keyData)); err == nil {
			return decoded, nil
		}
		return keyData, nil
	}

	return nil, errors.New("no encryption key configured")
}

// deriveKey derives a 32-byte key from the configured key material
func (s *SymmetricEncryption) deriveKey() ([]byte, error) {
	keyMaterial, err := s.getKey()
	if err != nil {
		return nil, err
	}

	// Use SHA256 to derive a consistent 32-byte key
	hash := sha256.Sum256(keyMaterial)
	return hash[:], nil
}

// EncryptAES256GCM encrypts data using AES-256-GCM
func (s *SymmetricEncryption) EncryptAES256GCM(plaintext []byte) ([]byte, error) {
	key, err := s.deriveKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptAES256GCM decrypts data using AES-256-GCM
func (s *SymmetricEncryption) DecryptAES256GCM(ciphertext []byte) ([]byte, error) {
	key, err := s.deriveKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// EncryptChaCha20Poly1305 encrypts data using ChaCha20-Poly1305
func (s *SymmetricEncryption) EncryptChaCha20Poly1305(plaintext []byte) ([]byte, error) {
	key, err := s.deriveKey()
	if err != nil {
		return nil, err
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %w", err)
	}

	// Generate a random nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptChaCha20Poly1305 decrypts data using ChaCha20-Poly1305
func (s *SymmetricEncryption) DecryptChaCha20Poly1305(ciphertext []byte) ([]byte, error) {
	key, err := s.deriveKey()
	if err != nil {
		return nil, err
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %w", err)
	}

	if len(ciphertext) < aead.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce := ciphertext[:aead.NonceSize()]
	ciphertext = ciphertext[aead.NonceSize():]

	// Decrypt the data
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// Encrypt encrypts data using the configured symmetric algorithm
func (s *SymmetricEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	algorithm := s.config.Encryption.Symmetric.Algorithm
	if algorithm == "" {
		algorithm = "AES256-GCM" // Default
	}

	switch algorithm {
	case "AES256-GCM":
		return s.EncryptAES256GCM(plaintext)
	case "ChaCha20-Poly1305":
		return s.EncryptChaCha20Poly1305(plaintext)
	default:
		return nil, fmt.Errorf("unsupported symmetric algorithm: %s", algorithm)
	}
}

// Decrypt decrypts data using the configured symmetric algorithm
func (s *SymmetricEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	algorithm := s.config.Encryption.Symmetric.Algorithm
	if algorithm == "" {
		algorithm = "AES256-GCM" // Default
	}

	switch algorithm {
	case "AES256-GCM":
		return s.DecryptAES256GCM(ciphertext)
	case "ChaCha20-Poly1305":
		return s.DecryptChaCha20Poly1305(ciphertext)
	default:
		return nil, fmt.Errorf("unsupported symmetric algorithm: %s", algorithm)
	}
}

// GetEncryptionManager returns the encryption manager for the app
func (app *App) GetEncryptionManager() *EncryptionManager {
	return NewEncryptionManager(app)
}

// AsymmetricEncryption handles asymmetric encryption operations
type AsymmetricEncryption struct {
	config *ModConfig
}

// NewAsymmetricEncryption creates a new asymmetric encryption instance
func NewAsymmetricEncryption(config *ModConfig) *AsymmetricEncryption {
	return &AsymmetricEncryption{config: config}
}

// getPublicKey returns the public key from configuration
func (a *AsymmetricEncryption) getPublicKey() (interface{}, error) {
	var keyData []byte
	var err error

	if a.config.Encryption.Asymmetric.PublicKey != "" {
		// Use key from config
		keyData = []byte(a.config.Encryption.Asymmetric.PublicKey)
	} else if a.config.Encryption.Asymmetric.PublicKeyFile != "" {
		// Read key from file
		keyData, err = ioutil.ReadFile(a.config.Encryption.Asymmetric.PublicKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read public key file: %w", err)
		}
	} else {
		return nil, errors.New("no public key configured")
	}

	// Parse PEM block
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing public key")
	}

	// Parse public key based on algorithm
	algorithm := a.config.Encryption.Asymmetric.Algorithm
	switch algorithm {
	case "RSA-OAEP", "":
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
		}
		rsaPubKey, ok := pubKey.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("not an RSA public key")
		}
		return rsaPubKey, nil
	case "ECDH":
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ECDH public key: %w", err)
		}
		return pubKey, nil
	default:
		return nil, fmt.Errorf("unsupported asymmetric algorithm: %s", algorithm)
	}
}

// getPrivateKey returns the private key from configuration
func (a *AsymmetricEncryption) getPrivateKey() (interface{}, error) {
	var keyData []byte
	var err error

	if a.config.Encryption.Asymmetric.PrivateKey != "" {
		// Use key from config
		keyData = []byte(a.config.Encryption.Asymmetric.PrivateKey)
	} else if a.config.Encryption.Asymmetric.PrivateKeyFile != "" {
		// Read key from file
		keyData, err = ioutil.ReadFile(a.config.Encryption.Asymmetric.PrivateKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}
	} else {
		return nil, errors.New("no private key configured")
	}

	// Parse PEM block
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing private key")
	}

	// Parse private key based on algorithm
	algorithm := a.config.Encryption.Asymmetric.Algorithm
	switch algorithm {
	case "RSA-OAEP", "":
		privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			// Try PKCS1 format
			privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
			}
		}
		rsaPrivKey, ok := privKey.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
		return rsaPrivKey, nil
	case "ECDH":
		privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ECDH private key: %w", err)
		}
		return privKey, nil
	default:
		return nil, fmt.Errorf("unsupported asymmetric algorithm: %s", algorithm)
	}
}

// EncryptRSAOAEP encrypts data using RSA-OAEP
func (a *AsymmetricEncryption) EncryptRSAOAEP(plaintext []byte) ([]byte, error) {
	pubKeyInterface, err := a.getPublicKey()
	if err != nil {
		return nil, err
	}

	pubKey, ok := pubKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid RSA public key")
	}

	// RSA-OAEP has size limitations, so we encrypt the data in chunks
	keySize := pubKey.Size()
	maxPlaintextSize := keySize - 2*sha256.Size - 2 // OAEP padding overhead

	if len(plaintext) <= maxPlaintextSize {
		// Single block encryption
		return rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, plaintext, nil)
	}

	// Multi-block encryption for larger data
	var result []byte
	for i := 0; i < len(plaintext); i += maxPlaintextSize {
		end := i + maxPlaintextSize
		if end > len(plaintext) {
			end = len(plaintext)
		}

		chunk := plaintext[i:end]
		encryptedChunk, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, chunk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt chunk: %w", err)
		}

		result = append(result, encryptedChunk...)
	}

	return result, nil
}

// DecryptRSAOAEP decrypts data using RSA-OAEP
func (a *AsymmetricEncryption) DecryptRSAOAEP(ciphertext []byte) ([]byte, error) {
	privKeyInterface, err := a.getPrivateKey()
	if err != nil {
		return nil, err
	}

	privKey, ok := privKeyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("invalid RSA private key")
	}

	keySize := privKey.Size()

	if len(ciphertext) == keySize {
		// Single block decryption
		return rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, ciphertext, nil)
	}

	// Multi-block decryption
	if len(ciphertext)%keySize != 0 {
		return nil, errors.New("invalid ciphertext length for RSA decryption")
	}

	var result []byte
	for i := 0; i < len(ciphertext); i += keySize {
		chunk := ciphertext[i : i+keySize]
		decryptedChunk, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, chunk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt chunk: %w", err)
		}

		result = append(result, decryptedChunk...)
	}

	return result, nil
}

// EncryptECDH encrypts data using ECDH (generates shared secret + AES)
func (a *AsymmetricEncryption) EncryptECDH(plaintext []byte) ([]byte, error) {
	// For ECDH, we generate an ephemeral key pair, perform ECDH key exchange,
	// and use the shared secret to encrypt with AES-GCM
	pubKeyInterface, err := a.getPublicKey()
	if err != nil {
		return nil, err
	}

	// Generate ephemeral key pair
	curve := ecdh.P256() // Use P-256 curve
	ephemeralPrivKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	ephemeralPubKey := ephemeralPrivKey.PublicKey()

	// Convert recipient's public key to ECDH format
	recipientPubKey, ok := pubKeyInterface.(*ecdh.PublicKey)
	if !ok {
		return nil, errors.New("invalid ECDH public key")
	}

	// Perform ECDH key exchange
	sharedSecret, err := ephemeralPrivKey.ECDH(recipientPubKey)
	if err != nil {
		return nil, fmt.Errorf("ECDH key exchange failed: %w", err)
	}

	// Derive encryption key from shared secret
	hash := sha256.Sum256(sharedSecret)
	encryptionKey := hash[:]

	// Encrypt using AES-GCM
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Prepend ephemeral public key to ciphertext
	ephemeralPubKeyBytes := ephemeralPubKey.Bytes()
	result := make([]byte, len(ephemeralPubKeyBytes)+len(ciphertext))
	copy(result, ephemeralPubKeyBytes)
	copy(result[len(ephemeralPubKeyBytes):], ciphertext)

	return result, nil
}

// DecryptECDH decrypts data using ECDH
func (a *AsymmetricEncryption) DecryptECDH(data []byte) ([]byte, error) {
	privKeyInterface, err := a.getPrivateKey()
	if err != nil {
		return nil, err
	}

	privKey, ok := privKeyInterface.(*ecdh.PrivateKey)
	if !ok {
		return nil, errors.New("invalid ECDH private key")
	}

	// Extract ephemeral public key
	curve := ecdh.P256()
	pubKeySize := curve.PublicKeySize()

	if len(data) < pubKeySize {
		return nil, errors.New("data too short to contain ephemeral public key")
	}

	ephemeralPubKeyBytes := data[:pubKeySize]
	ciphertext := data[pubKeySize:]

	ephemeralPubKey, err := curve.NewPublicKey(ephemeralPubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ephemeral public key: %w", err)
	}

	// Perform ECDH key exchange
	sharedSecret, err := privKey.ECDH(ephemeralPubKey)
	if err != nil {
		return nil, fmt.Errorf("ECDH key exchange failed: %w", err)
	}

	// Derive decryption key from shared secret
	hash := sha256.Sum256(sharedSecret)
	decryptionKey := hash[:]

	// Decrypt using AES-GCM
	block, err := aes.NewCipher(decryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// Encrypt encrypts data using the configured asymmetric algorithm
func (a *AsymmetricEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	algorithm := a.config.Encryption.Asymmetric.Algorithm
	if algorithm == "" {
		algorithm = "RSA-OAEP" // Default
	}

	switch algorithm {
	case "RSA-OAEP":
		return a.EncryptRSAOAEP(plaintext)
	case "ECDH":
		return a.EncryptECDH(plaintext)
	default:
		return nil, fmt.Errorf("unsupported asymmetric algorithm: %s", algorithm)
	}
}

// Decrypt decrypts data using the configured asymmetric algorithm
func (a *AsymmetricEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	algorithm := a.config.Encryption.Asymmetric.Algorithm
	if algorithm == "" {
		algorithm = "RSA-OAEP" // Default
	}

	switch algorithm {
	case "RSA-OAEP":
		return a.DecryptRSAOAEP(ciphertext)
	case "ECDH":
		return a.DecryptECDH(ciphertext)
	default:
		return nil, fmt.Errorf("unsupported asymmetric algorithm: %s", algorithm)
	}
}

// SignatureVerification handles digital signature operations
type SignatureVerification struct {
	config *ModConfig
}

// NewSignatureVerification creates a new signature verification instance
func NewSignatureVerification(config *ModConfig) *SignatureVerification {
	return &SignatureVerification{config: config}
}

// getSigningKey returns the signing key from configuration
func (s *SignatureVerification) getSigningKey() ([]byte, error) {
	if s.config.Encryption.Signature.Key != "" {
		// Decode base64 key
		return base64.StdEncoding.DecodeString(s.config.Encryption.Signature.Key)
	}

	if s.config.Encryption.Signature.KeyFile != "" {
		// Read key from file
		keyData, err := ioutil.ReadFile(s.config.Encryption.Signature.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read signature key file: %w", err)
		}

		// Try to decode as base64, if fails use raw data
		if decoded, err := base64.StdEncoding.DecodeString(string(keyData)); err == nil {
			return decoded, nil
		}
		return keyData, nil
	}

	return nil, errors.New("no signature key configured")
}

// SignHMACSHA256 creates an HMAC-SHA256 signature
func (s *SignatureVerification) SignHMACSHA256(data []byte) ([]byte, error) {
	key, err := s.getSigningKey()
	if err != nil {
		return nil, err
	}

	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil), nil
}

// VerifyHMACSHA256 verifies an HMAC-SHA256 signature
func (s *SignatureVerification) VerifyHMACSHA256(data, signature []byte) error {
	expectedSignature, err := s.SignHMACSHA256(data)
	if err != nil {
		return err
	}

	if !hmac.Equal(signature, expectedSignature) {
		return errors.New("HMAC signature verification failed")
	}

	return nil
}

// Sign creates a digital signature for the given data
func (s *SignatureVerification) Sign(data []byte) ([]byte, error) {
	algorithm := s.config.Encryption.Signature.Algorithm
	if algorithm == "" {
		algorithm = "HMAC-SHA256" // Default
	}

	switch algorithm {
	case "HMAC-SHA256":
		return s.SignHMACSHA256(data)
	default:
		return nil, fmt.Errorf("unsupported signature algorithm: %s", algorithm)
	}
}

// Verify verifies a digital signature for the given data
func (s *SignatureVerification) Verify(data, signature []byte) error {
	algorithm := s.config.Encryption.Signature.Algorithm
	if algorithm == "" {
		algorithm = "HMAC-SHA256" // Default
	}

	switch algorithm {
	case "HMAC-SHA256":
		return s.VerifyHMACSHA256(data, signature)
	default:
		return fmt.Errorf("unsupported signature algorithm: %s", algorithm)
	}
}