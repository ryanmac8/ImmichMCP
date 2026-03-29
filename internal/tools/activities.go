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

// RegisterActivityTools registers all activity-related tools.
func RegisterActivityTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(
		mcp.NewTool("immich_activities_list",
			mcp.WithDescription("List activities (comments and likes) for an album or asset."),
			mcp.WithString("albumId", mcp.Description("Album ID (UUID) - required"), mcp.Required()),
			mcp.WithString("assetId", mcp.Description("Asset ID (UUID) - optional filter")),
			mcp.WithString("type", mcp.Description("Activity type: comment or like")),
			mcp.WithString("level", mcp.Description("Level: album or asset")),
		),
		activityListHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_activities_create",
			mcp.WithDescription("Create a comment or like on an album or asset."),
			mcp.WithString("albumId", mcp.Description("Album ID (UUID)"), mcp.Required()),
			mcp.WithString("type", mcp.Description("Activity type: comment or like")),
			mcp.WithString("assetId", mcp.Description("Asset ID (UUID) - optional")),
			mcp.WithString("comment", mcp.Description("Comment text (required for comment type)")),
		),
		activityCreateHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_activities_delete",
			mcp.WithDescription("Delete an activity (comment or like)."),
			mcp.WithString("id", mcp.Description("Activity ID (UUID)"), mcp.Required()),
			mcp.WithBoolean("confirm", mcp.Description("Must be true to confirm deletion")),
		),
		activityDeleteHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_activities_statistics",
			mcp.WithDescription("Get activity statistics (comment count) for an album or asset."),
			mcp.WithString("albumId", mcp.Description("Album ID (UUID)"), mcp.Required()),
			mcp.WithString("assetId", mcp.Description("Asset ID (UUID) - optional")),
		),
		activityStatisticsHandler(c),
	)
}

func activityListHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		albumID := strArg(req, "albumId")
		if albumID == "" {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "Album ID is required", nil, meta))), nil
		}
		var assetID, actType, level *string
		if v := strArg(req, "assetId"); v != "" {
			assetID = &v
		}
		if v := strArg(req, "type"); v != "" {
			assetID = nil // reset
			actType = &v
		}
		if v := strArg(req, "level"); v != "" {
			level = &v
		}
		// Re-get assetId since we incorrectly reset it above
		if v := strArg(req, "assetId"); v != "" {
			assetID = &v
		}

		activities, err := c.GetActivities(ctx, albumID, assetID, actType, level)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}
		summaries := make([]models.ActivitySummary, len(activities))
		for i, a := range activities {
			summaries[i] = models.ActivitySummaryFromActivity(a)
		}
		total := len(summaries)
		meta.Total = &total
		return mcp.NewToolResultText(mustJSON(models.Success(summaries, meta))), nil
	}
}

func activityCreateHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		albumID := strArg(req, "albumId")
		if albumID == "" {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "Album ID is required", nil, meta))), nil
		}
		actType := strings.ToLower(strArg(req, "type"))
		if actType == "" {
			actType = "comment"
		}
		if actType != "comment" && actType != "like" {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "Type must be 'comment' or 'like'", nil, meta))), nil
		}
		comment := strArg(req, "comment")
		if actType == "comment" && comment == "" {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "Comment text is required for comment type", nil, meta))), nil
		}

		createReq := models.ActivityCreateRequest{
			AlbumID: albumID,
			Type:    actType,
		}
		if v := strArg(req, "assetId"); v != "" {
			createReq.AssetID = &v
		}
		if actType == "comment" {
			createReq.Comment = &comment
		}

		activity, err := c.CreateActivity(ctx, createReq)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, "Failed to create activity: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"id":         activity.ID,
			"type":       activity.Type,
			"comment":    activity.Comment,
			"album_id":   albumID,
			"asset_id":   activity.AssetID,
			"created_at": activity.CreatedAt,
			"user":       activity.User.Name,
		}, meta))), nil
	}
}

func activityDeleteHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		confirm := false
		if v, ok := req.GetArguments()["confirm"].(bool); ok {
			confirm = v
		}
		if !confirm {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(
				models.ErrConfirmationRequired,
				"Deletion requires confirm=true.",
				map[string]string{"activity_id": id},
				meta,
			))), nil
		}
		if err := c.DeleteActivity(ctx, id); err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, fmt.Sprintf("Failed to delete activity %s: %v", id, err), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"deleted":     true,
			"activity_id": id,
		}, meta))), nil
	}
}

func activityStatisticsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		albumID := strArg(req, "albumId")
		if albumID == "" {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "Album ID is required", nil, meta))), nil
		}
		var assetID *string
		if v := strArg(req, "assetId"); v != "" {
			assetID = &v
		}
		stats, err := c.GetActivityStatistics(ctx, albumID, assetID)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"album_id": albumID,
			"asset_id": assetID,
			"comments": stats.Comments,
		}, meta))), nil
	}
}
