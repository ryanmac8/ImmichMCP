package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ryanmac8/ImmichMCP/internal/client"
	"github.com/ryanmac8/ImmichMCP/internal/models"
)

// RegisterHealthTools registers all health/capability tools on the server.
func RegisterHealthTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(
		mcp.NewTool("immich_ping",
			mcp.WithDescription("Verify connectivity and authentication with the Immich instance. Returns server version if available."),
		),
		pingHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_capabilities",
			mcp.WithDescription("Return supported API features and detected Immich server capabilities."),
		),
		capabilitiesHandler(c),
	)
}

func pingHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		info, err := c.Ping(ctx)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(
				models.ErrUpstreamError,
				"Failed to connect to Immich: "+err.Error(),
				nil, meta,
			))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"connected": true,
			"version":   info.Version,
			"build":     info.Build,
			"nodejs":    info.Nodejs,
			"ffmpeg":    info.Ffmpeg,
			"exiftool":  info.Exiftool,
		}, meta))), nil
	}
}

func capabilitiesHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())

		info, _ := c.Ping(ctx)
		features, _ := c.GetFeatures(ctx)

		var version *string
		if info != nil {
			version = &info.Version
		}

		caps := map[string]interface{}{
			"connected": info != nil,
			"version":   version,
			"features":  features,
			"endpoints": map[string]interface{}{
				"assets": map[string]string{
					"list":       "/api/assets",
					"get":        "/api/assets/{id}",
					"upload":     "/api/assets",
					"update":     "/api/assets/{id}",
					"bulk_update": "/api/assets",
					"delete":     "/api/assets",
					"statistics": "/api/assets/statistics",
					"original":   "/api/assets/{id}/original",
					"thumbnail":  "/api/assets/{id}/thumbnail",
				},
				"search": map[string]string{
					"metadata": "/api/search/metadata",
					"smart":    "/api/search/smart",
					"explore":  "/api/search/explore",
				},
				"albums": map[string]string{
					"list":         "/api/albums",
					"get":          "/api/albums/{id}",
					"create":       "/api/albums",
					"update":       "/api/albums/{id}",
					"delete":       "/api/albums/{id}",
					"add_assets":   "/api/albums/{id}/assets",
					"remove_assets": "/api/albums/{id}/assets",
					"statistics":   "/api/albums/statistics",
				},
				"people": map[string]string{
					"list":   "/api/people",
					"get":    "/api/people/{id}",
					"update": "/api/people/{id}",
					"merge":  "/api/people/{id}/merge",
					"assets": "/api/people/{id}/assets",
				},
				"tags": map[string]string{
					"list":        "/api/tags",
					"get":         "/api/tags/{id}",
					"create":      "/api/tags",
					"update":      "/api/tags/{id}",
					"delete":      "/api/tags/{id}",
					"tag_assets":  "/api/tags/{id}/assets",
					"untag_assets": "/api/tags/{id}/assets",
				},
				"shared_links": map[string]string{
					"list":   "/api/shared-links",
					"get":    "/api/shared-links/{id}",
					"create": "/api/shared-links",
					"update": "/api/shared-links/{id}",
					"delete": "/api/shared-links/{id}",
				},
				"activities": map[string]string{
					"list":       "/api/activities",
					"create":     "/api/activities",
					"delete":     "/api/activities/{id}",
					"statistics": "/api/activities/statistics",
				},
			},
		}

		return mcp.NewToolResultText(mustJSON(models.Success(caps, meta))), nil
	}
}
