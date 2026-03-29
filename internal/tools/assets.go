package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ryanmac8/ImmichMCP/internal/client"
	"github.com/ryanmac8/ImmichMCP/internal/models"
)

// RegisterAssetTools registers all asset-related tools.
func RegisterAssetTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(
		mcp.NewTool("immich_assets_list",
			mcp.WithDescription("List recent assets with optional filters and pagination."),
			mcp.WithNumber("size", mcp.Description("Number of assets to return (default: 25, max: 1000)")),
			mcp.WithBoolean("isFavorite", mcp.Description("Filter by favorite status")),
			mcp.WithBoolean("isArchived", mcp.Description("Filter by archived status")),
			mcp.WithBoolean("isTrashed", mcp.Description("Filter by trashed status")),
			mcp.WithString("updatedAfter", mcp.Description("Filter assets updated after this date (ISO format)")),
			mcp.WithString("updatedBefore", mcp.Description("Filter assets updated before this date (ISO format)")),
		),
		assetListHandler(c),
	)

	s.AddTool(
		mcp.NewTool("immich_assets_get",
			mcp.WithDescription("Get full asset metadata by ID."),
			mcp.WithString("id", mcp.Description("Asset ID (UUID)"), mcp.Required()),
		),
		assetGetHandler(c),
	)

	s.AddTool(
		mcp.NewTool("immich_assets_exif",
			mcp.WithDescription("Get EXIF metadata for an asset."),
			mcp.WithString("id", mcp.Description("Asset ID (UUID)"), mcp.Required()),
		),
		assetExifHandler(c),
	)

	s.AddTool(
		mcp.NewTool("immich_assets_download_original",
			mcp.WithDescription("Get download URL for the original asset file."),
			mcp.WithString("id", mcp.Description("Asset ID (UUID)"), mcp.Required()),
		),
		assetDownloadOriginalHandler(c),
	)

	s.AddTool(
		mcp.NewTool("immich_assets_download_thumbnail",
			mcp.WithDescription("Get thumbnail and preview URLs for an asset."),
			mcp.WithString("id", mcp.Description("Asset ID (UUID)"), mcp.Required()),
		),
		assetDownloadThumbnailHandler(c),
	)

	s.AddTool(
		mcp.NewTool("immich_assets_upload",
			mcp.WithDescription("Upload a new asset from base64-encoded content."),
			mcp.WithString("fileContent", mcp.Description("Base64-encoded file content"), mcp.Required()),
			mcp.WithString("fileName", mcp.Description("Original filename with extension"), mcp.Required()),
			mcp.WithBoolean("isFavorite", mcp.Description("Mark as favorite")),
			mcp.WithBoolean("isArchived", mcp.Description("Mark as archived")),
		),
		assetUploadHandler(c),
	)

	s.AddTool(
		mcp.NewTool("immich_assets_update",
			mcp.WithDescription("Update asset metadata (favorite status, description, date, location, etc.)."),
			mcp.WithString("id", mcp.Description("Asset ID (UUID)"), mcp.Required()),
			mcp.WithBoolean("isFavorite", mcp.Description("Set favorite status")),
			mcp.WithBoolean("isArchived", mcp.Description("Set archived status")),
			mcp.WithString("description", mcp.Description("Set description")),
			mcp.WithString("dateTimeOriginal", mcp.Description("Set date/time original (ISO format)")),
			mcp.WithNumber("latitude", mcp.Description("Set latitude")),
			mcp.WithNumber("longitude", mcp.Description("Set longitude")),
			mcp.WithNumber("rating", mcp.Description("Set rating (0-5)")),
		),
		assetUpdateHandler(c),
	)

	s.AddTool(
		mcp.NewTool("immich_assets_bulk_update",
			mcp.WithDescription("Perform bulk operations on multiple assets. Supports dry run mode."),
			mcp.WithString("assetIds", mcp.Description("Asset IDs (comma-separated UUIDs)"), mcp.Required()),
			mcp.WithBoolean("isFavorite", mcp.Description("Set favorite status for all")),
			mcp.WithBoolean("isArchived", mcp.Description("Set archived status for all")),
			mcp.WithNumber("rating", mcp.Description("Set rating for all (0-5)")),
			mcp.WithBoolean("dryRun", mcp.Description("Dry run mode (default: true)")),
			mcp.WithBoolean("confirm", mcp.Description("Must be true to execute")),
		),
		assetBulkUpdateHandler(c),
	)

	s.AddTool(
		mcp.NewTool("immich_assets_delete",
			mcp.WithDescription("Delete asset(s). Requires explicit confirmation."),
			mcp.WithString("assetIds", mcp.Description("Asset IDs (comma-separated UUIDs)"), mcp.Required()),
			mcp.WithBoolean("force", mcp.Description("Force delete (bypass trash)")),
			mcp.WithBoolean("dryRun", mcp.Description("Dry run mode (default: true)")),
			mcp.WithBoolean("confirm", mcp.Description("Must be true to confirm deletion")),
		),
		assetDeleteHandler(c),
	)

	s.AddTool(
		mcp.NewTool("immich_assets_statistics",
			mcp.WithDescription("Get asset statistics (count of images, videos, total)."),
		),
		assetStatisticsHandler(c),
	)
}

func assetListHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())

		size := 25
		if v, ok := req.GetArguments()["size"].(float64); ok {
			size = clamp(int(v), 1, 1000)
		}
		var isFav, isArch, isTrashed *bool
		if v, ok := req.GetArguments()["isFavorite"].(bool); ok {
			isFav = &v
		}
		if v, ok := req.GetArguments()["isArchived"].(bool); ok {
			isArch = &v
		}
		if v, ok := req.GetArguments()["isTrashed"].(bool); ok {
			isTrashed = &v
		}
		updAfter := parseDate(strArg(req, "updatedAfter"))
		updBefore := parseDate(strArg(req, "updatedBefore"))

		assets, err := c.GetAssets(ctx, &size, isFav, isArch, isTrashed, updAfter, updBefore, nil)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}
		summaries := make([]models.AssetSummary, len(assets))
		for i, a := range assets {
			summaries[i] = models.AssetSummaryFromAsset(a)
		}
		total := len(summaries)
		meta.Total = &total
		return mcp.NewToolResultText(mustJSON(models.Success(summaries, meta))), nil
	}
}

func assetGetHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		asset, err := c.GetAsset(ctx, id)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Asset not found: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(asset, meta))), nil
	}
}

func assetExifHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		asset, err := c.GetAsset(ctx, id)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Asset not found: "+err.Error(), nil, meta))), nil
		}
		if asset.ExifInfo == nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "No EXIF data for asset "+id, nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(asset.ExifInfo, meta))), nil
	}
}

func assetDownloadOriginalHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		asset, err := c.GetAsset(ctx, id)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Asset not found: "+err.Error(), nil, meta))), nil
		}
		info := c.GetAssetDownloadInfo(id, &asset.OriginalFileName)
		var fileSize *int64
		if asset.ExifInfo != nil {
			fileSize = asset.ExifInfo.FileSizeInByte
		}
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"id":                id,
			"original_file_name": asset.OriginalFileName,
			"original_url":      info.OriginalURL,
			"mime_type":         asset.OriginalMimeType,
			"file_size":         fileSize,
		}, meta))), nil
	}
}

func assetDownloadThumbnailHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		asset, err := c.GetAsset(ctx, id)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Asset not found: "+err.Error(), nil, meta))), nil
		}
		info := c.GetAssetDownloadInfo(id, &asset.OriginalFileName)
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"id":                id,
			"original_file_name": asset.OriginalFileName,
			"thumbnail_url":     info.ThumbnailURL,
			"preview_url":       info.PreviewURL,
			"thumbhash":         asset.Thumbhash,
		}, meta))), nil
	}
}

func assetUploadHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())

		encoded := strArg(req, "fileContent")
		fileName := strArg(req, "fileName")

		fileBytes, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "Invalid base64 file content", nil, meta))), nil
		}

		var isFav, isArch *bool
		if v, ok := req.GetArguments()["isFavorite"].(bool); ok {
			isFav = &v
		}
		if v, ok := req.GetArguments()["isArchived"].(bool); ok {
			isArch = &v
		}

		deviceAssetID := fmt.Sprintf("%s-%d-%d", fileName, len(fileBytes), time.Now().UnixNano())
		asset, err := c.UploadAsset(ctx, fileBytes, fileName, deviceAssetID, time.Now().UTC(), isFav, isArch)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, "Upload failed: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"asset_id":           asset.ID,
			"type":               asset.Type,
			"original_file_name": asset.OriginalFileName,
			"status":             "uploaded",
		}, meta))), nil
	}
}

func assetUpdateHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")

		updateReq := models.AssetUpdateRequest{}
		if v, ok := req.GetArguments()["isFavorite"].(bool); ok {
			updateReq.IsFavorite = &v
		}
		if v, ok := req.GetArguments()["isArchived"].(bool); ok {
			updateReq.IsArchived = &v
		}
		if v := strArg(req, "description"); v != "" {
			updateReq.Description = &v
		}
		if v := strArg(req, "dateTimeOriginal"); v != "" {
			updateReq.DateTimeOriginal = parseDate(v)
		}
		if v, ok := req.GetArguments()["latitude"].(float64); ok {
			updateReq.Latitude = &v
		}
		if v, ok := req.GetArguments()["longitude"].(float64); ok {
			updateReq.Longitude = &v
		}
		if v, ok := req.GetArguments()["rating"].(float64); ok {
			r := int(v)
			updateReq.Rating = &r
		}

		asset, err := c.UpdateAsset(ctx, id, updateReq)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Update failed: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(asset, meta))), nil
	}
}

func assetBulkUpdateHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())

		ids := parseStringArray(strArg(req, "assetIds"))
		if len(ids) == 0 {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "No valid asset IDs provided", nil, meta))), nil
		}

		dryRun := true
		if v, ok := req.GetArguments()["dryRun"].(bool); ok {
			dryRun = v
		}
		confirm := false
		if v, ok := req.GetArguments()["confirm"].(bool); ok {
			confirm = v
		}

		if dryRun || !confirm {
			msg := "This is a dry run. Set dryRun=false and confirm=true to execute."
			if !dryRun {
				msg = "Set confirm=true to execute the operation."
			}
			return mcp.NewToolResultText(mustJSON(models.Success(models.BulkOperationResult{
				AffectedIDs: ids,
				Warnings:    []string{msg},
				Executed:    false,
			}, meta))), nil
		}

		bulkReq := models.AssetBulkUpdateRequest{IDs: ids}
		if v, ok := req.GetArguments()["isFavorite"].(bool); ok {
			bulkReq.IsFavorite = &v
		}
		if v, ok := req.GetArguments()["isArchived"].(bool); ok {
			bulkReq.IsArchived = &v
		}
		if v, ok := req.GetArguments()["rating"].(float64); ok {
			r := int(v)
			bulkReq.Rating = &r
		}

		if err := c.BulkUpdateAssets(ctx, bulkReq); err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, "Bulk update failed: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(models.BulkOperationResult{
			AffectedIDs: ids,
			Executed:    true,
		}, meta))), nil
	}
}

func assetDeleteHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())

		ids := parseStringArray(strArg(req, "assetIds"))
		if len(ids) == 0 {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "No valid asset IDs provided", nil, meta))), nil
		}
		force := false
		if v, ok := req.GetArguments()["force"].(bool); ok {
			force = v
		}
		dryRun := true
		if v, ok := req.GetArguments()["dryRun"].(bool); ok {
			dryRun = v
		}
		confirm := false
		if v, ok := req.GetArguments()["confirm"].(bool); ok {
			confirm = v
		}

		if dryRun || !confirm {
			// Show preview of first 10
			preview := make([]map[string]interface{}, 0, 10)
			for _, id := range ids {
				if len(preview) >= 10 {
					break
				}
				if a, err := c.GetAsset(ctx, id); err == nil {
					preview = append(preview, map[string]interface{}{
						"id":                 a.ID,
						"original_file_name": a.OriginalFileName,
						"type":               a.Type,
					})
				}
			}
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(
				models.ErrConfirmationRequired,
				fmt.Sprintf("Deletion requires confirm=true. %d asset(s) would be deleted.", len(ids)),
				map[string]interface{}{"asset_count": len(ids), "force": force, "preview": preview},
				meta,
			))), nil
		}

		if err := c.DeleteAssets(ctx, ids, force); err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, "Delete failed: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"deleted":     true,
			"asset_count": len(ids),
			"asset_ids":   ids,
			"force":       force,
		}, meta))), nil
	}
}

func assetStatisticsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		stats, err := c.GetAssetStatistics(ctx)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(stats, meta))), nil
	}
}

// strArg is a helper to extract a string argument from the request.
func strArg(req mcp.CallToolRequest, key string) string {
	if v, ok := req.GetArguments()[key].(string); ok {
		return v
	}
	return ""
}
