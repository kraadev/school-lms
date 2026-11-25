package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorLog records every failed HTTP response with enough context to diagnose it.
// Handlers may still provide a more specific message and next_steps payload.
func ErrorLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		status := c.Writer.Status()
		if status < http.StatusBadRequest {
			return
		}

		reason := "Handler mengembalikan respons gagal; periksa pesan respons dan data permintaan."
		nextStep := "Baca alasan pada respons, perbaiki data atau dependensi yang disebutkan, lalu ulangi proses."
		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			reason = "Permintaan ditolak karena autentikasi atau hak akses tidak memenuhi syarat."
			nextStep = "Pastikan sesi login masih aktif dan akun memiliki hak akses untuk operasi ini."
		} else if status == http.StatusNotFound {
			reason = "Data atau endpoint yang diminta tidak ditemukan."
			nextStep = "Periksa ID dan URL, lalu muat ulang data sebelum mencoba lagi."
		} else if status >= http.StatusInternalServerError {
			reason = "Terjadi kesalahan internal pada server atau database."
			nextStep = "Periksa log server untuk detail teknis dan coba kembali setelah masalah diperbaiki."
		}

		log.Printf("ERROR HTTP: method=%s path=%s status=%d reason=%s next_step=%s", c.Request.Method, c.Request.URL.Path, status, reason, nextStep)
	}
}
