package handlers

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"log/slog"
	"net/http"

	"github.com/mihailtudos/metrickit/pkg/helpers"
)

// WithRequestDecryptor is a middleware that handles RSA-AES request decryption
func WithRequestDecryptor(privateKey *rsa.PrivateKey, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			isEncrypted := r.Header.Get("X-Encryption") == "RSA-AES"

			if isEncrypted {
				if privateKey == nil {
					http.Error(w, "Server not configured for encryption", http.StatusInternalServerError)
					return
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					logger.ErrorContext(r.Context(), "failed to read request body", helpers.ErrAttr(err))
					http.Error(w, "Failed to read request body", http.StatusBadRequest)
					return
				}

				if err := r.Body.Close(); err != nil {
					logger.ErrorContext(r.Context(), "failed to close request body", helpers.ErrAttr(err))
					http.Error(w, "Failed to close request body", http.StatusInternalServerError)
					return
				}

				// Decrypt data
				keySize := privateKey.Size()
				encryptedKey := body[:keySize]
				encryptedData := body[keySize:]
				aesKey, errDec := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedKey)
				if errDec != nil {
					logger.DebugContext(r.Context(), "failed to decrypt AES key", helpers.ErrAttr(errDec))
					http.Error(w, "Failed to decrypt AES key", http.StatusBadRequest)
					return
				}

				block, errCiph := aes.NewCipher(aesKey)
				if errCiph != nil {
					http.Error(w, "Failed to create AES cipher", http.StatusInternalServerError)
					return
				}

				gcm, errGcm := cipher.NewGCM(block)
				if errGcm != nil {
					http.Error(w, "Failed to create GCM", http.StatusInternalServerError)
					return
				}

				// Extract nonce and ciphertext
				nonceSize := gcm.NonceSize()
				if len(encryptedData) < nonceSize {
					http.Error(w, "Malformed encrypted data", http.StatusBadRequest)
					return
				}
				nonce := encryptedData[:nonceSize]
				ciphertext := encryptedData[nonceSize:]

				// Decrypt the data
				body, err = gcm.Open(nil, nonce, ciphertext, nil)
				if err != nil {
					http.Error(w, "Failed to decrypt data", http.StatusBadRequest)
					return
				}

				r.Body = io.NopCloser(bytes.NewBuffer(body))
			}

			next.ServeHTTP(w, r)
		})
	}
}
