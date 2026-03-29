package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ryanmac8/ImmichMCP/internal/client"
	"github.com/ryanmac8/ImmichMCP/internal/models"
)

// RegisterAlbumTools registers all album-related tools.
func RegisterAlbumTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(
		mcp.NewTool("immich_albums_list",
			mcp.WithDescription("List all albums with optional filters."),
			mcp.WithBoolean("shared", mcp.Description("Filter to shared albums only")),
			mcp.WithString("assetId", mcp.Description("Filter to albums containing this asset ID")),
		),
		albumListHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_albums_get",
			mcp.WithDescription("Get album details by ID."),
			mcp.WithString("id", mcp.Description("Album ID (UUID)"), mcp.Required()),
			mcp.WithBoolean("withoutAssets", mcp.Description("Exclude assets from response")),
		),
		albumGetHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_albums_create",
			mcp.WithDescription("Create a new album."),
			mcp.WithString("albumName", mcp.Description("Album name"), mcp.Required()),
			mcp.WithString("description", mcp.Description("Album description")),
			mcp.WithString("assetIds", mcp.Description("Initial asset IDs (comma-separated UUIDs)")),
			mcp.WithString("sharedWithUserIds", mcp.Description("User IDs to share with (comma-separated UUIDs)")),
		),
		albumCreateHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_albums_update",
			mcp.WithDescription("Update album metadata (name, description)."),
			mcp.WithString("id", mcp.Description("Album ID (UUID)"), mcp.Required()),
			mcp.WithString("albumName", mcp.Description("New album name")),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithBoolean("isActivityEnabled", mcp.Description("Enable/disable activity")),
			mcp.WithString("order", mcp.Description("Sort order: asc or desc")),
		),
		albumUpdateHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_albums_assets_add",
			mcp.WithDescription("Add assets to an album."),
			mcp.WithString("albumId", mcp.Description("Album ID (UUID)"), mcp.Required()),
			mcp.WithString("assetIds", mcp.Description("Asset IDs to add (comma-separated UUIDs)"), mcp.Required()),
		),
		albumAddAssetsHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_albums_assets_remove",
			mcp.WithDescription("Remove assets from an album."),
			mcp.WithString("albumId", mcp.Description("Album ID (UUID)"), mcp.Required()),
			mcp.WithString("assetIds", mcp.Description("Asset IDs to remove (comma-separated UUIDs)"), mcp.Required()),
		),
		albumRemoveAssetsHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_albums_delete",
			mcp.WithDescription("Delete an album. Requires explicit confirmation."),
			mcp.WithString("id", mcp.Description("Album ID (UUID)"), mcp.Required()),
			mcp.WithBoolean("confirm", mcp.Description("Must be true to confirm deletion")),
		),
		albumDeleteHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_albums_statistics",
			mcp.WithDescription("Get album statistics (owned, shared, not shared counts)."),
		),
		albumStatisticsHandler(c),
	)
}

func albumListHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		var shared *bool
		if v, ok := req.GetArguments()["shared"].(bool); ok {
			shared = &v
		}
		var assetID *string
		if v := strArg(req, "assetId"); v != "" {
			assetID = &v
		}
		albums, err := c.GetAlbums(ctx, shared, assetID)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}
		summaries := make([]models.AlbumSummary, len(albums))
		for i, a := range albums {
			summaries[i] = models.AlbumSummaryFromAlbum(a)
		}
		total := len(summaries)
		meta.Total = &total
		return mcp.NewToolResultText(mustJSON(models.Success(summaries, meta))), nil
	}
}

func albumGetHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		var withoutAssets *bool
		if v, ok := req.GetArguments()["withoutAssets"].(bool); ok {
			withoutAssets = &v
		}
		album, err := c.GetAlbum(ctx, id, withoutAssets)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Album not found: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(album, meta))), nil
	}
}

func albumCreateHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		name := strArg(req, "albumName")
		if name == "" {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "Album name is required", nil, meta))), nil
		}
		createReq := models.AlbumCreateRequest{AlbumName: name}
		if v := strArg(req, "description"); v != "" {
			createReq.Description = &v
		}
		createReq.AssetIDs = parseStringArray(strArg(req, "assetIds"))
		createReq.SharedWithUserIDs = parseStringArray(strArg(req, "sharedWithUserIds"))

		album, err := c.CreateAlbum(ctx, createReq)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, "Failed to create album: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(album, meta))), nil
	}
}

func albumUpdateHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		updateReq := models.AlbumUpdateRequest{}
		if v := strArg(req, "albumName"); v != "" {
			updateReq.AlbumName = &v
		}
		if v := strArg(req, "description"); v != "" {
			updateReq.Description = &v
		}
		if v, ok := req.GetArguments()["isActivityEnabled"].(bool); ok {
			updateReq.IsActivityEnabled = &v
		}
		if v := strArg(req, "order"); v != "" {
			updateReq.Order = &v
		}
		album, err := c.UpdateAlbum(ctx, id, updateReq)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Update failed: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(album, meta))), nil
	}
}

func albumAddAssetsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		albumID := strArg(req, "albumId")
		ids := parseStringArray(strArg(req, "assetIds"))
		if len(ids) == 0 {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "No valid asset IDs provided", nil, meta))), nil
		}
		result, err := c.AddAssetsToAlbum(ctx, albumID, ids)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}
		added, failed := countBulk(result)
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"album_id": albumID,
			"added":    added,
			"failed":   failed,
			"results":  result,
		}, meta))), nil
	}
}

func albumRemoveAssetsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		albumID := strArg(req, "albumId")
		ids := parseStringArray(strArg(req, "assetIds"))
		if len(ids) == 0 {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "No valid asset IDs provided", nil, meta))), nil
		}
		result, err := c.RemoveAssetsFromAlbum(ctx, albumID, ids)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}
		removed, failed := countBulk(result)
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"album_id": albumID,
			"removed":  removed,
			"failed":   failed,
			"results":  result,
		}, meta))), nil
	}
}

func albumDeleteHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		confirm := false
		if v, ok := req.GetArguments()["confirm"].(bool); ok {
			confirm = v
		}
		if !confirm {
			album, err := c.GetAlbum(ctx, id, ptr(true))
			if err != nil {
				return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Album not found: "+err.Error(), nil, meta))), nil
			}
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(
				models.ErrConfirmationRequired,
				"Deletion requires confirm=true.",
				map[string]interface{}{
					"album_id":    id,
					"album_name":  album.AlbumName,
					"asset_count": album.AssetCount,
					"shared":      album.Shared,
				}, meta,
			))), nil
		}
		if err := c.DeleteAlbum(ctx, id); err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, fmt.Sprintf("Failed to delete album %s: %v", id, err), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"deleted":  true,
			"album_id": id,
		}, meta))), nil
	}
}

func albumStatisticsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		stats, err := c.GetAlbumStatistics(ctx)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(stats, meta))), nil
	}
}

func countBulk(results []models.BulkIDResponse) (success, failed int) {
	for _, r := range results {
		if r.Success {
			success++
		} else {
			failed++
		}
	}
	return
}
