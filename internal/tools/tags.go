package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ryanmac8/ImmichMCP/internal/client"
	"github.com/ryanmac8/ImmichMCP/internal/models"
)

// RegisterTagTools registers all tag-related tools.
func RegisterTagTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(
		mcp.NewTool("immich_tags_list",
			mcp.WithDescription("List all tags."),
		),
		tagListHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_tags_get",
			mcp.WithDescription("Get tag by ID."),
			mcp.WithString("id", mcp.Description("Tag ID (UUID)"), mcp.Required()),
		),
		tagGetHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_tags_create",
			mcp.WithDescription("Create a new tag."),
			mcp.WithString("name", mcp.Description("Tag name"), mcp.Required()),
			mcp.WithString("color", mcp.Description("Tag color (hex, e.g., '#ff0000')")),
		),
		tagCreateHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_tags_update",
			mcp.WithDescription("Update a tag."),
			mcp.WithString("id", mcp.Description("Tag ID (UUID)"), mcp.Required()),
			mcp.WithString("name", mcp.Description("New tag name")),
			mcp.WithString("color", mcp.Description("New tag color (hex)")),
		),
		tagUpdateHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_tags_delete",
			mcp.WithDescription("Delete a tag. Requires explicit confirmation."),
			mcp.WithString("id", mcp.Description("Tag ID (UUID)"), mcp.Required()),
			mcp.WithBoolean("confirm", mcp.Description("Must be true to confirm deletion")),
		),
		tagDeleteHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_tags_assets_add",
			mcp.WithDescription("Tag assets with a specific tag."),
			mcp.WithString("tagId", mcp.Description("Tag ID (UUID)"), mcp.Required()),
			mcp.WithString("assetIds", mcp.Description("Asset IDs to tag (comma-separated UUIDs)"), mcp.Required()),
		),
		tagAssetsAddHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_tags_assets_remove",
			mcp.WithDescription("Remove tag from assets."),
			mcp.WithString("tagId", mcp.Description("Tag ID (UUID)"), mcp.Required()),
			mcp.WithString("assetIds", mcp.Description("Asset IDs to untag (comma-separated UUIDs)"), mcp.Required()),
		),
		tagAssetsRemoveHandler(c),
	)
}

func tagListHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		tags, err := c.GetTags(ctx)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}
		summaries := make([]models.TagSummary, len(tags))
		for i, t := range tags {
			summaries[i] = models.TagSummaryFromTag(t)
		}
		total := len(summaries)
		meta.Total = &total
		return mcp.NewToolResultText(mustJSON(models.Success(summaries, meta))), nil
	}
}

func tagGetHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		tag, err := c.GetTag(ctx, id)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Tag not found: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(tag, meta))), nil
	}
}

func tagCreateHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		name := strArg(req, "name")
		if name == "" {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "Tag name is required", nil, meta))), nil
		}
		createReq := models.TagCreateRequest{Name: name}
		if v := strArg(req, "color"); v != "" {
			createReq.Color = &v
		}
		tag, err := c.CreateTag(ctx, createReq)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, "Failed to create tag: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(tag, meta))), nil
	}
}

func tagUpdateHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		updateReq := models.TagUpdateRequest{}
		if v := strArg(req, "name"); v != "" {
			updateReq.Name = &v
		}
		if v := strArg(req, "color"); v != "" {
			updateReq.Color = &v
		}
		tag, err := c.UpdateTag(ctx, id, updateReq)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Update failed: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(tag, meta))), nil
	}
}

func tagDeleteHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		confirm := false
		if v, ok := req.GetArguments()["confirm"].(bool); ok {
			confirm = v
		}
		if !confirm {
			tag, err := c.GetTag(ctx, id)
			if err != nil {
				return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Tag not found: "+err.Error(), nil, meta))), nil
			}
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(
				models.ErrConfirmationRequired,
				"Deletion requires confirm=true.",
				map[string]interface{}{"tag_id": id, "name": tag.Name, "value": tag.Value},
				meta,
			))), nil
		}
		if err := c.DeleteTag(ctx, id); err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, fmt.Sprintf("Failed to delete tag %s: %v", id, err), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"deleted": true,
			"tag_id":  id,
		}, meta))), nil
	}
}

func tagAssetsAddHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		tagID := strArg(req, "tagId")
		ids := parseStringArray(strArg(req, "assetIds"))
		if len(ids) == 0 {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "No valid asset IDs provided", nil, meta))), nil
		}
		result, err := c.TagAssets(ctx, tagID, ids)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}
		tagged, failed := countBulk(result)
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"tag_id":  tagID,
			"tagged":  tagged,
			"failed":  failed,
			"results": result,
		}, meta))), nil
	}
}

func tagAssetsRemoveHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		tagID := strArg(req, "tagId")
		ids := parseStringArray(strArg(req, "assetIds"))
		if len(ids) == 0 {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "No valid asset IDs provided", nil, meta))), nil
		}
		result, err := c.UntagAssets(ctx, tagID, ids)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}
		untagged, failed := countBulk(result)
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"tag_id":   tagID,
			"untagged": untagged,
			"failed":   failed,
			"results":  result,
		}, meta))), nil
	}
}
