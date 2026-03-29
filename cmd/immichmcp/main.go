package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/ryanmac8/ImmichMCP/internal/client"
	"github.com/ryanmac8/ImmichMCP/internal/config"
	"github.com/ryanmac8/ImmichMCP/internal/tools"
	"github.com/ryanmac8/ImmichMCP/internal/upload"
)

func main() {
	stdio := false
	for _, arg := range os.Args[1:] {
		if arg == "--stdio" {
			stdio = true
			break
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	c := client.New(cfg)
	sessions := upload.NewSessionService(30 * time.Minute)
	defer sessions.Close()

	mcpServer := server.NewMCPServer(
		"ImmichMCP",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	tools.RegisterHealthTools(mcpServer, c)
	tools.RegisterAssetTools(mcpServer, c)
	tools.RegisterAlbumTools(mcpServer, c)
	tools.RegisterSearchTools(mcpServer, c)
	tools.RegisterPeopleTools(mcpServer, c)
	tools.RegisterTagTools(mcpServer, c)
	tools.RegisterSharedLinkTools(mcpServer, c)
	tools.RegisterActivityTools(mcpServer, c)
	tools.RegisterUploadTools(mcpServer, c, sessions)

	if stdio {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
		slog.Info("ImmichMCP starting in stdio mode")
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("stdio server error: %v", err)
		}
		return
	}

	// HTTP + SSE mode
	sseServer := server.NewSSEServer(mcpServer,
		server.WithBaseURL(fmt.Sprintf("http://0.0.0.0:%d", cfg.Port)),
	)

	mux := http.NewServeMux()

	// MCP SSE endpoint
	mux.Handle("/", sseServer)

	// Out-of-band upload endpoint
	mux.HandleFunc("/upload/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		sessionID := strings.TrimPrefix(r.URL.Path, "/upload/")
		if sessionID == "" {
			http.Error(w, "missing session ID", http.StatusBadRequest)
			return
		}

		sess := sessions.GetSession(sessionID)
		if sess == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found", "session_id": sessionID})
			return
		}
		if sess.Status == upload.StatusExpired {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session expired", "session_id": sessionID})
			return
		}
		if sess.Status == upload.StatusCompleted {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "session already completed", "asset_id": sess.AssetID})
			return
		}

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "expected multipart/form-data"})
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			sessions.SetFailed(sessionID, "no file provided")
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no file provided in form data. Use field name 'file'."})
			return
		}
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			sessions.SetFailed(sessionID, "failed to read file: "+err.Error())
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read file"})
			return
		}

		fileName := header.Filename
		if sess.FileName != nil {
			fileName = *sess.FileName
		}

		sessions.SetStatus(sessionID, upload.StatusUploading)
		slog.Info("receiving upload", "file", fileName, "size", len(fileBytes), "session", sessionID)

		deviceAssetID := fmt.Sprintf("%s-%d-%d", fileName, len(fileBytes), time.Now().UnixNano())
		asset, err := c.UploadAsset(r.Context(), fileBytes, fileName, deviceAssetID, time.Now().UTC(), sess.IsFavorite, sess.IsArchived)
		if err != nil {
			sessions.SetFailed(sessionID, err.Error())
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to upload asset to Immich: " + err.Error()})
			return
		}

		sessions.SetCompleted(sessionID, asset.ID)
		slog.Info("upload complete", "file", fileName, "asset_id", asset.ID)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success":            true,
			"asset_id":           asset.ID,
			"original_file_name": asset.OriginalFileName,
			"type":               asset.Type,
			"session_id":         sessionID,
		})
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		})
	})

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	slog.Info("ImmichMCP server starting",
		"port", cfg.Port,
		"sse_endpoint", fmt.Sprintf("http://localhost:%d/sse", cfg.Port),
		"upload_endpoint", fmt.Sprintf("http://localhost:%d/upload/{sessionId}", cfg.Port),
	)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
