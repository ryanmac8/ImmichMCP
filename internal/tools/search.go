package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ryanmac8/ImmichMCP/internal/client"
	"github.com/ryanmac8/ImmichMCP/internal/models"
)

// RegisterSearchTools registers all search-related tools.
func RegisterSearchTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(
		mcp.NewTool("immich_search_metadata",
			mcp.WithDescription("Search assets by metadata filters (dates, type, location, camera, person, etc.)."),
			mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			mcp.WithNumber("size", mcp.Description("Page size (default: 25, max: 100)")),
			mcp.WithString("type", mcp.Description("Asset type: IMAGE, VIDEO, or ALL")),
			mcp.WithBoolean("isFavorite", mcp.Description("Filter by favorite status")),
			mcp.WithBoolean("isArchived", mcp.Description("Filter by archived status")),
			mcp.WithBoolean("isTrashed", mcp.Description("Filter by trashed status")),
			mcp.WithString("takenAfter", mcp.Description("Filter assets taken after (YYYY-MM-DD)")),
			mcp.WithString("takenBefore", mcp.Description("Filter assets taken before (YYYY-MM-DD)")),
			mcp.WithString("city", mcp.Description("Filter by city")),
			mcp.WithString("state", mcp.Description("Filter by state/province")),
			mcp.WithString("country", mcp.Description("Filter by country")),
			mcp.WithString("make", mcp.Description("Filter by camera make")),
			mcp.WithString("model", mcp.Description("Filter by camera model")),
			mcp.WithString("lensModel", mcp.Description("Filter by lens model")),
			mcp.WithString("personIds", mcp.Description("Filter by person IDs (comma-separated UUIDs)")),
			mcp.WithString("originalFileName", mcp.Description("Filter by original file name (partial match)")),
			mcp.WithString("order", mcp.Description("Sort order: asc or desc")),
		),
		searchMetadataHandler(c),
	)

	s.AddTool(
		mcp.NewTool("immich_search_smart",
			mcp.WithDescription("ML-based semantic search using CLIP. Search using natural language queries like 'sunset at the beach'."),
			mcp.WithString("query", mcp.Description("Natural language search query"), mcp.Required()),
			mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			mcp.WithNumber("size", mcp.Description("Page size (default: 25, max: 100)")),
			mcp.WithString("type", mcp.Description("Asset type: IMAGE, VIDEO, or ALL")),
			mcp.WithBoolean("isFavorite", mcp.Description("Filter by favorite status")),
			mcp.WithBoolean("isArchived", mcp.Description("Filter by archived status")),
			mcp.WithString("takenAfter", mcp.Description("Filter assets taken after (YYYY-MM-DD)")),
			mcp.WithString("takenBefore", mcp.Description("Filter assets taken before (YYYY-MM-DD)")),
			mcp.WithString("city", mcp.Description("Filter by city")),
			mcp.WithString("state", mcp.Description("Filter by state/province")),
			mcp.WithString("country", mcp.Description("Filter by country")),
			mcp.WithString("make", mcp.Description("Filter by camera make")),
			mcp.WithString("model", mcp.Description("Filter by camera model")),
			mcp.WithString("personIds", mcp.Description("Filter by person IDs (comma-separated UUIDs)")),
		),
		searchSmartHandler(c),
	)

	s.AddTool(
		mcp.NewTool("immich_search_explore",
			mcp.WithDescription("Get explore/discovery data showing popular places, things, and people from your library."),
		),
		searchExploreHandler(c),
	)
}

func searchMetadataHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())

		page := 1
		size := 25
		if v, ok := req.GetArguments()["page"].(float64); ok && v >= 1 {
			page = int(v)
		}
		if v, ok := req.GetArguments()["size"].(float64); ok {
			size = clamp(int(v), 1, 100)
		}

		searchReq := models.MetadataSearchRequest{
			Page: &page,
			Size: &size,
		}
		if v := strArg(req, "type"); v != "" {
			up := strings.ToUpper(v)
			searchReq.Type = &up
		}
		if v, ok := req.GetArguments()["isFavorite"].(bool); ok {
			searchReq.IsFavorite = &v
		}
		if v, ok := req.GetArguments()["isArchived"].(bool); ok {
			searchReq.IsArchived = &v
		}
		if v, ok := req.GetArguments()["isTrashed"].(bool); ok {
			searchReq.IsTrashed = &v
		}
		searchReq.TakenAfter = parseDate(strArg(req, "takenAfter"))
		searchReq.TakenBefore = parseDate(strArg(req, "takenBefore"))
		optStr(req, "city", &searchReq.City)
		optStr(req, "state", &searchReq.State)
		optStr(req, "country", &searchReq.Country)
		optStr(req, "make", &searchReq.Make)
		optStr(req, "model", &searchReq.Model)
		optStr(req, "lensModel", &searchReq.LensModel)
		optStr(req, "originalFileName", &searchReq.OriginalFileName)
		optStr(req, "order", &searchReq.Order)
		searchReq.PersonIDs = parseStringArray(strArg(req, "personIds"))

		result, err := c.SearchMetadata(ctx, searchReq)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}

		summaries := make([]models.AssetSummary, len(result.Items))
		for i, a := range result.Items {
			summaries[i] = models.AssetSummaryFromAsset(a)
		}
		total := result.Total
		meta.Total = &total
		meta.Page = &page
		meta.PageSize = &size
		meta.Next = result.NextPage
		return mcp.NewToolResultText(mustJSON(models.Success(summaries, meta))), nil
	}
}

func searchSmartHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		query := strArg(req, "query")
		if query == "" {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "Search query is required", nil, meta))), nil
		}

		page := 1
		size := 25
		if v, ok := req.GetArguments()["page"].(float64); ok && v >= 1 {
			page = int(v)
		}
		if v, ok := req.GetArguments()["size"].(float64); ok {
			size = clamp(int(v), 1, 100)
		}

		searchReq := models.SmartSearchRequest{
			Query: query,
			Page:  &page,
			Size:  &size,
		}
		if v := strArg(req, "type"); v != "" {
			up := strings.ToUpper(v)
			searchReq.Type = &up
		}
		if v, ok := req.GetArguments()["isFavorite"].(bool); ok {
			searchReq.IsFavorite = &v
		}
		if v, ok := req.GetArguments()["isArchived"].(bool); ok {
			searchReq.IsArchived = &v
		}
		searchReq.TakenAfter = parseDate(strArg(req, "takenAfter"))
		searchReq.TakenBefore = parseDate(strArg(req, "takenBefore"))
		optStr(req, "city", &searchReq.City)
		optStr(req, "state", &searchReq.State)
		optStr(req, "country", &searchReq.Country)
		optStr(req, "make", &searchReq.Make)
		optStr(req, "model", &searchReq.Model)
		searchReq.PersonIDs = parseStringArray(strArg(req, "personIds"))

		result, err := c.SearchSmart(ctx, searchReq)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}

		summaries := make([]models.AssetSummary, len(result.Items))
		for i, a := range result.Items {
			summaries[i] = models.AssetSummaryFromAsset(a)
		}
		total := result.Total
		meta.Total = &total
		meta.Page = &page
		meta.PageSize = &size
		meta.Next = result.NextPage
		return mcp.NewToolResultText(mustJSON(models.Success(summaries, meta))), nil
	}
}

func searchExploreHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		data, err := c.SearchExplore(ctx)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(data, meta))), nil
	}
}

// optStr sets *dst to a pointer to the string arg if non-empty.
func optStr(req mcp.CallToolRequest, key string, dst **string) {
	if v := strArg(req, key); v != "" {
		*dst = &v
	}
}
