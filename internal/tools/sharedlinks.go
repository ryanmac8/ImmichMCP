package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ryanmac8/ImmichMCP/internal/client"
	"github.com/ryanmac8/ImmichMCP/internal/models"
)

// RegisterSharedLinkTools registers all shared-link-related tools.
func RegisterSharedLinkTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(
		mcp.NewTool("immich_shared_links_list",
			mcp.WithDescription("List all shared links."),
		),
		sharedLinkListHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_shared_links_get",
			mcp.WithDescription("Get shared link by ID."),
			mcp.WithString("id", mcp.Description("Shared link ID (UUID)"), mcp.Required()),
		),
		sharedLinkGetHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_shared_links_create",
			mcp.WithDescription("Create a new shared link for an album or assets."),
			mcp.WithString("type", mcp.Description("Link type: ALBUM or INDIVIDUAL")),
			mcp.WithString("albumId", mcp.Description("Album ID (required if type is ALBUM)")),
			mcp.WithString("assetIds", mcp.Description("Asset IDs (comma-separated, required if type is INDIVIDUAL)")),
			mcp.WithString("expiresAt", mcp.Description("Expiration date (ISO format)")),
			mcp.WithBoolean("allowUpload", mcp.Description("Allow upload (default: false)")),
			mcp.WithBoolean("allowDownload", mcp.Description("Allow download (default: true)")),
			mcp.WithBoolean("showMetadata", mcp.Description("Show metadata (default: true)")),
			mcp.WithString("password", mcp.Description("Password protection")),
			mcp.WithString("description", mcp.Description("Description")),
		),
		sharedLinkCreateHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_shared_links_update",
			mcp.WithDescription("Update shared link settings (expiry, permissions, etc.)."),
			mcp.WithString("id", mcp.Description("Shared link ID (UUID)"), mcp.Required()),
			mcp.WithString("expiresAt", mcp.Description("New expiration date (ISO format)")),
			mcp.WithBoolean("allowUpload", mcp.Description("Allow upload")),
			mcp.WithBoolean("allowDownload", mcp.Description("Allow download")),
			mcp.WithBoolean("showMetadata", mcp.Description("Show metadata")),
			mcp.WithString("password", mcp.Description("Password")),
			mcp.WithString("description", mcp.Description("Description")),
		),
		sharedLinkUpdateHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_shared_links_delete",
			mcp.WithDescription("Delete a shared link. Requires explicit confirmation."),
			mcp.WithString("id", mcp.Description("Shared link ID (UUID)"), mcp.Required()),
			mcp.WithBoolean("confirm", mcp.Description("Must be true to confirm deletion")),
		),
		sharedLinkDeleteHandler(c),
	)
}

func sharedLinkListHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		links, err := c.GetSharedLinks(ctx)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}
		summaries := make([]models.SharedLinkSummary, len(links))
		for i, l := range links {
			summaries[i] = models.SharedLinkSummaryFromLink(l)
		}
		total := len(summaries)
		meta.Total = &total
		return mcp.NewToolResultText(mustJSON(models.Success(summaries, meta))), nil
	}
}

func sharedLinkGetHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		link, err := c.GetSharedLink(ctx, id)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Shared link not found: "+err.Error(), nil, meta))), nil
		}
		shareURL := strings.TrimRight(c.BaseURL(), "/") + "/share/" + link.Key
		assetCount := len(link.Assets)
		if link.Album != nil && assetCount == 0 {
			assetCount = link.Album.AssetCount
		}
		var albumName *string
		if link.Album != nil {
			albumName = &link.Album.AlbumName
		}
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"id":            link.ID,
			"key":           link.Key,
			"share_url":     shareURL,
			"type":          link.Type,
			"created_at":    link.CreatedAt,
			"expires_at":    link.ExpiresAt,
			"allow_upload":  link.AllowUpload,
			"allow_download": link.AllowDownload,
			"show_metadata": link.ShowMetadata,
			"description":   link.Description,
			"album_name":    albumName,
			"asset_count":   assetCount,
		}, meta))), nil
	}
}

func sharedLinkCreateHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		linkType := strings.ToUpper(strArg(req, "type"))
		if linkType == "" {
			linkType = "INDIVIDUAL"
		}
		albumID := strArg(req, "albumId")
		assetIDsStr := strArg(req, "assetIds")

		if linkType == "ALBUM" && albumID == "" {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "Album ID is required for ALBUM type shared links", nil, meta))), nil
		}
		if linkType == "INDIVIDUAL" && assetIDsStr == "" {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "Asset IDs are required for INDIVIDUAL type shared links", nil, meta))), nil
		}

		createReq := models.SharedLinkCreateRequest{
			Type:     linkType,
			AssetIDs: parseStringArray(assetIDsStr),
		}
		if albumID != "" {
			createReq.AlbumID = &albumID
		}
		createReq.ExpiresAt = parseDate(strArg(req, "expiresAt"))
		if v, ok := req.GetArguments()["allowUpload"].(bool); ok {
			createReq.AllowUpload = &v
		}
		if v, ok := req.GetArguments()["allowDownload"].(bool); ok {
			createReq.AllowDownload = &v
		}
		if v, ok := req.GetArguments()["showMetadata"].(bool); ok {
			createReq.ShowMetadata = &v
		}
		if v := strArg(req, "password"); v != "" {
			createReq.Password = &v
		}
		if v := strArg(req, "description"); v != "" {
			createReq.Description = &v
		}

		link, err := c.CreateSharedLink(ctx, createReq)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, "Failed to create shared link: "+err.Error(), nil, meta))), nil
		}
		shareURL := strings.TrimRight(c.BaseURL(), "/") + "/share/" + link.Key
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"id":             link.ID,
			"key":            link.Key,
			"share_url":      shareURL,
			"type":           link.Type,
			"created_at":     link.CreatedAt,
			"expires_at":     link.ExpiresAt,
			"allow_upload":   link.AllowUpload,
			"allow_download": link.AllowDownload,
			"show_metadata":  link.ShowMetadata,
		}, meta))), nil
	}
}

func sharedLinkUpdateHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		updateReq := models.SharedLinkUpdateRequest{}
		updateReq.ExpiresAt = parseDate(strArg(req, "expiresAt"))
		if v, ok := req.GetArguments()["allowUpload"].(bool); ok {
			updateReq.AllowUpload = &v
		}
		if v, ok := req.GetArguments()["allowDownload"].(bool); ok {
			updateReq.AllowDownload = &v
		}
		if v, ok := req.GetArguments()["showMetadata"].(bool); ok {
			updateReq.ShowMetadata = &v
		}
		if v := strArg(req, "password"); v != "" {
			updateReq.Password = &v
		}
		if v := strArg(req, "description"); v != "" {
			updateReq.Description = &v
		}
		link, err := c.UpdateSharedLink(ctx, id, updateReq)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Update failed: "+err.Error(), nil, meta))), nil
		}
		shareURL := strings.TrimRight(c.BaseURL(), "/") + "/share/" + link.Key
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"id":             link.ID,
			"key":            link.Key,
			"share_url":      shareURL,
			"type":           link.Type,
			"expires_at":     link.ExpiresAt,
			"allow_upload":   link.AllowUpload,
			"allow_download": link.AllowDownload,
			"show_metadata":  link.ShowMetadata,
		}, meta))), nil
	}
}

func sharedLinkDeleteHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		confirm := false
		if v, ok := req.GetArguments()["confirm"].(bool); ok {
			confirm = v
		}
		if !confirm {
			link, err := c.GetSharedLink(ctx, id)
			if err != nil {
				return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Shared link not found: "+err.Error(), nil, meta))), nil
			}
			assetCount := len(link.Assets)
			if link.Album != nil && assetCount == 0 {
				assetCount = link.Album.AssetCount
			}
			var albumName *string
			if link.Album != nil {
				albumName = &link.Album.AlbumName
			}
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(
				models.ErrConfirmationRequired,
				"Deletion requires confirm=true.",
				map[string]interface{}{
					"link_id":    id,
					"key":        link.Key,
					"type":       link.Type,
					"album_name": albumName,
					"asset_count": assetCount,
				}, meta,
			))), nil
		}
		if err := c.DeleteSharedLink(ctx, id); err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, fmt.Sprintf("Failed to delete shared link %s: %v", id, err), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"deleted": true,
			"link_id": id,
		}, meta))), nil
	}
}
