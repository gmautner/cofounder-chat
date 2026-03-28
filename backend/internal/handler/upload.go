package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"cofounder-chat/internal/database/sqlc"
)

const maxUploadSize = 15 << 20 // 15 MB

func (h *Handler) HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "file too large (max 15 MB)")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	messageIDStr := r.FormValue("message_id")
	if messageIDStr == "" {
		writeError(w, http.StatusBadRequest, "message_id is required")
		return
	}

	messageID, err := parseUUID(messageIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid message_id")
		return
	}

	// Generate a unique filename
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	ext := filepath.Ext(header.Filename)
	storageName := hex.EncodeToString(randBytes) + ext
	storagePath := filepath.Join(h.Config.BlobStoragePath, storageName)

	dst, err := os.Create(storagePath)
	if err != nil {
		slog.Error("failed to create upload file", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to save file")
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		slog.Error("failed to write upload file", "err", err)
		os.Remove(storagePath)
		writeError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	attachment, err := h.Queries.CreateAttachment(r.Context(), sqlc.CreateAttachmentParams{
		MessageID:   messageID,
		FileName:    header.Filename,
		FileSize:    written,
		ContentType: contentType,
		StoragePath: fmt.Sprintf("/uploads/%s", storageName),
	})
	if err != nil {
		slog.Error("failed to create attachment record", "err", err)
		os.Remove(storagePath)
		writeError(w, http.StatusInternalServerError, "failed to save attachment")
		return
	}

	writeJSON(w, http.StatusCreated, attachment)
}
