package rest

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *handler) mwDecompress() gin.HandlerFunc {
	return func(c *gin.Context) {
		const gzipScheme = "gzip"
		if !strings.Contains(c.GetHeader("Content-Encoding"), gzipScheme) {
			c.Next()
			return
		}

		gzipReader, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			h.logger.Errorf("failed to create gzip reader: %v", err)
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer func() {
			err := gzipReader.Close()
			if err != nil {
				h.logger.Errorf("failed to close gzip reader: %v", err)
			}
		}()

		c.Request.Body = io.NopCloser(gzipReader)

		c.Writer.Header().Set("Content-Encoding", gzipScheme)
		c.Writer.Header().Set("Accept-Encoding", gzipScheme)

		c.Next()
	}
}

type gzipResponseWriter struct {
	io.Writer
	gin.ResponseWriter
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	// Проверяем тип контента и выполняем сжатие только для JSON и HTML
	contentType := w.Header().Get("Content-Type")
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
		n, err := w.Writer.Write(data)
		if err != nil {
			return n, fmt.Errorf("failed to write data: %w", err)
		}

		return n, nil
	}

	write, err := w.ResponseWriter.Write(data)
	if err != nil {
		return write, fmt.Errorf("failed to write data: %w", err)
	}

	return write, nil
}

func (h *handler) responseGzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, поддерживает ли клиент gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Выполняем обработку запроса и сохраняем ответ
		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Header().Set("Accept-Encoding", "gzip")

		// Перенаправляем вывод в gzip.Writer
		gz := gzip.NewWriter(c.Writer)
		defer func() {
			err := gz.Close()
			if err != nil {
				h.logger.Errorf("failed to close gzip writer: %v", err)
			}
		}()

		// Заменяем Writer на обертку для gzip
		c.Writer = &gzipResponseWriter{Writer: gz, ResponseWriter: c.Writer}

		c.Next()
	}
}

func (h *handler) encryptionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.hashKey == "" {
			c.Next()
			return
		}

		// Проверяем наличие ключа в заголовке
		hashKey := c.GetHeader("HashSHA256")

		if !checkHashKey(hashKey, h.hashKey) {
			c.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		c.Next()
	}
}

func checkHashKey(hash string, key string) bool {
	bytes, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}

	h := hmac.New(sha256.New, []byte(key))
	_, err = h.Write(bytes)
	if err != nil {
		return false
	}

	return hmac.Equal(h.Sum(nil), bytes)
}
