package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ryanmac8/ImmichMCP/internal/client"
	"github.com/ryanmac8/ImmichMCP/internal/models"
)

// RegisterPeopleTools registers all people-related tools.
func RegisterPeopleTools(s *server.MCPServer, c *client.Client) {
	s.AddTool(
		mcp.NewTool("immich_people_list",
			mcp.WithDescription("List all recognized people in the library."),
			mcp.WithBoolean("withHidden", mcp.Description("Include hidden people")),
			mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			mcp.WithNumber("size", mcp.Description("Page size (default: 25, max: 100)")),
		),
		peopleListHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_people_get",
			mcp.WithDescription("Get person details by ID."),
			mcp.WithString("id", mcp.Description("Person ID (UUID)"), mcp.Required()),
		),
		peopleGetHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_people_update",
			mcp.WithDescription("Update person information (name, birth date, hidden status)."),
			mcp.WithString("id", mcp.Description("Person ID (UUID)"), mcp.Required()),
			mcp.WithString("name", mcp.Description("Person's name")),
			mcp.WithString("birthDate", mcp.Description("Birth date (YYYY-MM-DD)")),
			mcp.WithBoolean("isHidden", mcp.Description("Hide this person from views")),
		),
		peopleUpdateHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_people_merge",
			mcp.WithDescription("Merge multiple people into one (for duplicate face clusters)."),
			mcp.WithString("targetId", mcp.Description("Target person ID to merge into (UUID)"), mcp.Required()),
			mcp.WithString("sourceIds", mcp.Description("Source person IDs to merge from (comma-separated UUIDs)"), mcp.Required()),
			mcp.WithBoolean("confirm", mcp.Description("Must be true to confirm merge")),
		),
		peopleMergeHandler(c),
	)
	s.AddTool(
		mcp.NewTool("immich_people_assets",
			mcp.WithDescription("List assets containing a specific person."),
			mcp.WithString("personId", mcp.Description("Person ID (UUID)"), mcp.Required()),
			mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			mcp.WithNumber("size", mcp.Description("Page size (default: 25, max: 100)")),
		),
		peopleAssetsHandler(c),
	)
}

func peopleListHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())

		withHidden := false
		if v, ok := req.GetArguments()["withHidden"].(bool); ok {
			withHidden = v
		}
		page := 1
		size := 25
		if v, ok := req.GetArguments()["page"].(float64); ok && v >= 1 {
			page = int(v)
		}
		if v, ok := req.GetArguments()["size"].(float64); ok {
			size = clamp(int(v), 1, 100)
		}

		result, err := c.GetPeople(ctx, &withHidden)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}

		all := result.People
		total := len(all)
		skip := (page - 1) * size
		end := skip + size
		if end > total {
			end = total
		}
		var paged []models.PersonSummary
		if skip < total {
			for _, p := range all[skip:end] {
				paged = append(paged, models.PersonSummaryFromPerson(p.Person))
			}
		}
		if paged == nil {
			paged = []models.PersonSummary{}
		}

		meta.Page = &page
		meta.PageSize = &size
		meta.Total = &total
		if end < total {
			next := fmt.Sprintf("page=%d&size=%d", page+1, size)
			meta.Next = &next
		}

		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"people":  paged,
			"visible": result.Visible,
			"hidden":  result.Hidden,
		}, meta))), nil
	}
}

func peopleGetHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		person, err := c.GetPerson(ctx, id)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Person not found: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(person, meta))), nil
	}
}

func peopleUpdateHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		id := strArg(req, "id")
		updateReq := models.PersonUpdateRequest{}
		if v := strArg(req, "name"); v != "" {
			updateReq.Name = &v
		}
		if v := strArg(req, "birthDate"); v != "" {
			updateReq.BirthDate = &v
		}
		if v, ok := req.GetArguments()["isHidden"].(bool); ok {
			updateReq.IsHidden = &v
		}
		person, err := c.UpdatePerson(ctx, id, updateReq)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrNotFound, "Update failed: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(person, meta))), nil
	}
}

func peopleMergeHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		targetID := strArg(req, "targetId")
		sourceIDs := parseStringArray(strArg(req, "sourceIds"))
		if len(sourceIDs) == 0 {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrValidation, "No valid source person IDs provided", nil, meta))), nil
		}
		confirm := false
		if v, ok := req.GetArguments()["confirm"].(bool); ok {
			confirm = v
		}
		if !confirm {
			target, _ := c.GetPerson(ctx, targetID)
			sources := make([]map[string]string, 0)
			for _, id := range sourceIDs {
				if len(sources) >= 5 {
					break
				}
				if p, err := c.GetPerson(ctx, id); err == nil {
					sources = append(sources, map[string]string{"id": p.ID, "name": p.Name})
				}
			}
			var targetInfo interface{}
			if target != nil {
				targetInfo = map[string]string{"id": target.ID, "name": target.Name}
			}
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(
				models.ErrConfirmationRequired,
				"Merge requires confirm=true.",
				map[string]interface{}{"target": targetInfo, "sources": sources, "source_count": len(sourceIDs)},
				meta,
			))), nil
		}
		result, err := c.MergePeople(ctx, targetID, sourceIDs)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, "Merge failed: "+err.Error(), nil, meta))), nil
		}
		return mcp.NewToolResultText(mustJSON(models.Success(map[string]interface{}{
			"target_id":    targetID,
			"merged_count": len(sourceIDs),
			"results":      result,
		}, meta))), nil
	}
}

func peopleAssetsHandler(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		meta := models.NewMeta(c.BaseURL())
		personID := strArg(req, "personId")
		page := 1
		size := 25
		if v, ok := req.GetArguments()["page"].(float64); ok && v >= 1 {
			page = int(v)
		}
		if v, ok := req.GetArguments()["size"].(float64); ok {
			size = clamp(int(v), 1, 100)
		}

		assets, err := c.GetPersonAssets(ctx, personID)
		if err != nil {
			return mcp.NewToolResultText(mustJSON(models.ErrorResponse(models.ErrUpstreamError, err.Error(), nil, meta))), nil
		}

		total := len(assets)
		skip := (page - 1) * size
		end := skip + size
		if end > total {
			end = total
		}
		var paged []models.AssetSummary
		if skip < total {
			for _, a := range assets[skip:end] {
				paged = append(paged, models.AssetSummaryFromAsset(a))
			}
		}
		if paged == nil {
			paged = []models.AssetSummary{}
		}

		meta.Page = &page
		meta.PageSize = &size
		meta.Total = &total
		if end < total {
			next := fmt.Sprintf("page=%d&size=%d", page+1, size)
			meta.Next = &next
		}
		return mcp.NewToolResultText(mustJSON(models.Success(paged, meta))), nil
	}
}
