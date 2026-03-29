package tools

import (
	"context"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ryanmac8/ImmichMCP/internal/client"
	"github.com/ryanmac8/ImmichMCP/internal/models"
	"github.com/ryanmac8/ImmichMCP/internal/upload"
)

// RegisterUploadTools registers out-of-band upload tools (HTTP mode only).
func RegisterUploadTools(s *server.MCPServer, c *client.Client, sessions *upload.SessionService) {
	s.AddTool(
		mcp.NewTool("immich_assets_upload_init",
			mcp.WithDescription("Initialize an out-of-band file upload. Returns an upload URL where you can POST the file directly."),
			mcp.WithString("fileName", mcp.Description("Suggested filename (optional)")),
			mcp.WithBoolean("isFavorite", mcp.Description("Mark as favorite")),
			mcp.WithBoolean("isArchived", mcp.Description("Mark as archived")),
		),
		uploadInitHandler(c, sessions),
	)
	s.AddTool(
		mcp.NewTool("immich_assets_upload_status",
			mcp.WithDescription("Check the status of an upload session."),
			mcp.WithString("sessionId", mcp.Description("Upload session ID"), mcp.Required()),
		),
		uploadStatusHandler(c, sessions),
	)
}

func uploadInitHandler(c *client.Client, sessions *upload.SessionService) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())

		var fileName *string
		if v := strArg(req, "fileName"); v != "" {
			fileName = &v
		}
		var isFav, isArch *bool
		if v, ok := req.GetArguments()["isFavorite"].(bool); ok {
			isFav = &v
		}
		if v, ok := req.GetArguments()["isArchived"].(bool); ok {
			isArch = &v
		}

		sess := sessions.CreateSession(fileName, isFav, isArch)

		mcpBase := firstNonEmpty(
			os.Getenv("MCP_PUBLIC_URL"),
			os.Getenv("MCP_BASE_URL"),
			"http://localhost:5000",
		)
		uploadURL := mcpBase + "/upload/" + sess.SessionID

		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"session_id": sess.SessionID,
			"upload_url": uploadURL,
			"expires_at": sess.ExpiresAt,
			"instructions": map[string]string{
				"method":       "POST",
				"content_type": "multipart/form-data",
				"form_field":   "file",
				"example_curl": `curl -X POST -F "file=@/path/to/image.jpg" "` + uploadURL + `"`,
			},
		}, meta))), nil
	}
}

func uploadStatusHandler(c *client.Client, sessions *upload.SessionService) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		sessionID := strArg(req, "sessionId")
		sess := sessions.GetSession(sessionID)
		if sess == nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Upload session not found: "+sessionID, nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"session_id": sess.SessionID,
			"status":     sess.Status.String(),
			"asset_id":   sess.AssetID,
			"error":      sess.Error,
			"created_at": sess.CreatedAt,
			"expires_at": sess.ExpiresAt,
		}, meta))), nil
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
