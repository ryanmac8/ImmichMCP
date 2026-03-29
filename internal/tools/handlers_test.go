package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ryanmac8/ImmichMCP/internal/client"
	"github.com/ryanmac8/ImmichMCP/internal/config"
	"github.com/ryanmac8/ImmichMCP/internal/models"
	"github.com/ryanmac8/ImmichMCP/internal/upload"
	"time"
)

// testClient creates a Client pointing at the given test server.
func testClient(srv *httptest.Server) *client.Client {
	return client.New(&config.Config{
		BaseURL: srv.URL,
		APIKey:  "test-key",
	})
}

func respondJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func callResult(t *testing.T, handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]interface{}) map[string]interface{} {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	var out map[string]interface{}
	text := result.Content[0].(mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatalf("failed to parse result JSON: %v\nraw: %s", err, text)
	}
	return out
}

// ─── Health ─────────────────────────────────────────────────────────────────

func TestPingHandler_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, map[string]string{"res": "pong"})
	}))
	defer srv.Close()

	out := callResult(t, pingHandler(testClient(srv)), nil)
	if out["ok"] != true {
		t.Errorf("expected ok=true, got %v", out["ok"])
	}
}

func TestPingHandler_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer srv.Close()

	out := callResult(t, pingHandler(testClient(srv)), nil)
	if out["ok"] != false {
		t.Errorf("expected ok=false")
	}
}

// ─── Assets ──────────────────────────────────────────────────────────────────

func TestAssetGetHandler_Success(t *testing.T) {
	asset := models.Asset{ID: "asset-1", OriginalFileName: "photo.jpg", Type: "IMAGE"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, asset)
	}))
	defer srv.Close()

	out := callResult(t, assetGetHandler(testClient(srv)), map[string]interface{}{"id": "asset-1"})
	if out["ok"] != true {
		t.Errorf("expected ok=true")
	}
}

func TestAssetGetHandler_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Not found"}`))
	}))
	defer srv.Close()

	out := callResult(t, assetGetHandler(testClient(srv)), map[string]interface{}{"id": "missing-id"})
	if out["ok"] != false {
		t.Errorf("expected ok=false for not found asset")
	}
}

func TestAssetStatisticsHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, models.AssetStatistics{Images: 10, Videos: 2, Total: 12})
	}))
	defer srv.Close()

	out := callResult(t, assetStatisticsHandler(testClient(srv)), nil)
	if out["ok"] != true {
		t.Errorf("expected ok=true")
	}
}

func TestAssetDeleteHandler_RequiresConfirm(t *testing.T) {
	asset := models.Asset{ID: "asset-1", OriginalFileName: "photo.jpg", Type: "IMAGE"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, asset)
	}))
	defer srv.Close()

	// dryRun defaults to true, so without confirm=true we expect a preview/confirmation response
	out := callResult(t, assetDeleteHandler(testClient(srv)), map[string]interface{}{
		"assetIds": "asset-1",
		"dryRun":   false,
	})
	if out["ok"] != false {
		t.Errorf("expected ok=false without confirm")
	}
	errObj := out["error"].(map[string]interface{})
	if errObj["code"] != models.ErrConfirmationRequired {
		t.Errorf("expected confirmation required error, got %s", errObj["code"])
	}
}

// ─── Albums ──────────────────────────────────────────────────────────────────

func TestAlbumListHandler(t *testing.T) {
	albums := []models.Album{
		{ID: "album-1", AlbumName: "Vacation", AssetCount: 5},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, albums)
	}))
	defer srv.Close()

	out := callResult(t, albumListHandler(testClient(srv)), nil)
	if out["ok"] != true {
		t.Errorf("expected ok=true")
	}
	meta := out["meta"].(map[string]interface{})
	if meta["total"].(float64) != 1 {
		t.Errorf("expected total=1")
	}
}

func TestAlbumGetHandler_MissingID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	out := callResult(t, albumGetHandler(testClient(srv)), map[string]interface{}{})
	if out["ok"] != false {
		t.Errorf("expected validation error")
	}
}

func TestAlbumDeleteHandler_RequiresConfirm(t *testing.T) {
	album := models.Album{ID: "album-1", AlbumName: "Vacation", AssetCount: 3}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, album)
	}))
	defer srv.Close()

	out := callResult(t, albumDeleteHandler(testClient(srv)), map[string]interface{}{"id": "album-1"})
	if out["ok"] != false {
		t.Errorf("expected ok=false without confirm")
	}
	errObj := out["error"].(map[string]interface{})
	if errObj["code"] != models.ErrConfirmationRequired {
		t.Errorf("expected confirmation required")
	}
}

// ─── Tags ────────────────────────────────────────────────────────────────────

func TestTagListHandler(t *testing.T) {
	tags := []models.Tag{{ID: "tag-1", Name: "nature", Value: "nature"}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, tags)
	}))
	defer srv.Close()

	out := callResult(t, tagListHandler(testClient(srv)), nil)
	if out["ok"] != true {
		t.Errorf("expected ok=true")
	}
}

func TestTagCreateHandler_MissingName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	out := callResult(t, tagCreateHandler(testClient(srv)), map[string]interface{}{})
	if out["ok"] != false {
		t.Errorf("expected validation error for missing name")
	}
}

// ─── Activities ──────────────────────────────────────────────────────────────

func TestActivityListHandler_MissingAlbumID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	out := callResult(t, activityListHandler(testClient(srv)), map[string]interface{}{})
	if out["ok"] != false {
		t.Errorf("expected validation error for missing albumId")
	}
	errObj := out["error"].(map[string]interface{})
	if errObj["code"] != models.ErrValidation {
		t.Errorf("expected ErrValidation, got %s", errObj["code"])
	}
}

func TestActivityCreateHandler_InvalidType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	out := callResult(t, activityCreateHandler(testClient(srv)), map[string]interface{}{
		"albumId": "album-1",
		"type":    "reaction",
	})
	if out["ok"] != false {
		t.Errorf("expected validation error for invalid type")
	}
}

// ─── Upload ──────────────────────────────────────────────────────────────────

func TestUploadInitHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	sessions := upload.NewSessionService(time.Minute)
	defer sessions.Close()

	out := callResult(t, uploadInitHandler(testClient(srv), sessions), map[string]interface{}{
		"fileName": "photo.jpg",
	})
	if out["ok"] != true {
		t.Errorf("expected ok=true")
	}
	result := out["result"].(map[string]interface{})
	if result["session_id"] == "" {
		t.Error("expected non-empty session_id")
	}
	uploadURL := result["upload_url"].(string)
	if !strings.Contains(uploadURL, result["session_id"].(string)) {
		t.Error("upload_url should contain session_id")
	}
}

func TestUploadStatusHandler_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	sessions := upload.NewSessionService(time.Minute)
	defer sessions.Close()

	out := callResult(t, uploadStatusHandler(testClient(srv), sessions), map[string]interface{}{
		"sessionId": "nonexistent-id",
	})
	if out["ok"] != false {
		t.Errorf("expected ok=false for unknown session")
	}
}

// ─── Shared Links ─────────────────────────────────────────────────────────────

func TestSharedLinkCreateHandler_MissingAssetIDs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	out := callResult(t, sharedLinkCreateHandler(testClient(srv)), map[string]interface{}{
		"type": "INDIVIDUAL",
	})
	if out["ok"] != false {
		t.Errorf("expected validation error for missing assetIds")
	}
}

func TestSharedLinkCreateHandler_MissingAlbumID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	out := callResult(t, sharedLinkCreateHandler(testClient(srv)), map[string]interface{}{
		"type": "ALBUM",
	})
	if out["ok"] != false {
		t.Errorf("expected validation error for missing albumId")
	}
}

// ─── Search ───────────────────────────────────────────────────────────────────

func TestSearchSmartHandler_MissingQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	out := callResult(t, searchSmartHandler(testClient(srv)), map[string]interface{}{})
	if out["ok"] != false {
		t.Errorf("expected validation error for missing query")
	}
}

// ─── People ───────────────────────────────────────────────────────────────────

func TestPeopleMergeHandler_MissingSourceIDs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	out := callResult(t, peopleMergeHandler(testClient(srv)), map[string]interface{}{
		"targetId":  "person-1",
		"sourceIds": "",
	})
	if out["ok"] != false {
		t.Errorf("expected validation error for missing sourceIds")
	}
}
