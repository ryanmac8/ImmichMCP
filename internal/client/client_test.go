package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ryanmac8/ImmichMCP/internal/config"
	"github.com/ryanmac8/ImmichMCP/internal/models"
)

func newTestClient(t *testing.T, mux *http.ServeMux) *Client {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return New(&config.Config{
		BaseURL: srv.URL,
		APIKey:  "test-key",
	})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func TestPing(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/server/about", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		writeJSON(w, map[string]string{"res": "pong"})
	})
	c := newTestClient(t, mux)
	info, err := c.Ping(context.Background())
	if err != nil {
		t.Fatalf("Ping error: %v", err)
	}
	_ = info
}

func TestGetAsset(t *testing.T) {
	asset := models.Asset{ID: "abc-123", OriginalFileName: "photo.jpg", Type: "IMAGE"}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/assets/abc-123", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, asset)
	})
	c := newTestClient(t, mux)
	got, err := c.GetAsset(context.Background(), "abc-123")
	if err != nil {
		t.Fatalf("GetAsset error: %v", err)
	}
	if got.ID != "abc-123" {
		t.Errorf("expected ID abc-123, got %s", got.ID)
	}
	if got.OriginalFileName != "photo.jpg" {
		t.Errorf("unexpected filename: %s", got.OriginalFileName)
	}
}

func TestGetAsset_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/assets/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Not found"}`))
	})
	c := newTestClient(t, mux)
	_, err := c.GetAsset(context.Background(), "missing")
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestGetAlbums(t *testing.T) {
	albums := []models.Album{
		{ID: "album-1", AlbumName: "Vacation"},
		{ID: "album-2", AlbumName: "Family"},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/albums", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, albums)
	})
	c := newTestClient(t, mux)
	got, err := c.GetAlbums(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("GetAlbums error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 albums, got %d", len(got))
	}
	if got[0].AlbumName != "Vacation" {
		t.Errorf("unexpected album name: %s", got[0].AlbumName)
	}
}

func TestGetAlbum(t *testing.T) {
	album := models.Album{ID: "album-1", AlbumName: "Vacation", AssetCount: 10}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/albums/album-1", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, album)
	})
	c := newTestClient(t, mux)
	got, err := c.GetAlbum(context.Background(), "album-1", nil)
	if err != nil {
		t.Fatalf("GetAlbum error: %v", err)
	}
	if got.AssetCount != 10 {
		t.Errorf("expected 10 assets, got %d", got.AssetCount)
	}
}

func TestDeleteAlbum(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/albums/album-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	c := newTestClient(t, mux)
	if err := c.DeleteAlbum(context.Background(), "album-1"); err != nil {
		t.Fatalf("DeleteAlbum error: %v", err)
	}
}

func TestGetTags(t *testing.T) {
	tags := []models.Tag{
		{ID: "tag-1", Name: "nature", Value: "nature"},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, tags)
	})
	c := newTestClient(t, mux)
	got, err := c.GetTags(context.Background())
	if err != nil {
		t.Fatalf("GetTags error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "nature" {
		t.Errorf("unexpected tags: %v", got)
	}
}

func TestDeleteAssets(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/assets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	c := newTestClient(t, mux)
	if err := c.DeleteAssets(context.Background(), []string{"id-1", "id-2"}, false); err != nil {
		t.Fatalf("DeleteAssets error: %v", err)
	}
}

func TestGetActivityStatistics(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/activities/statistics", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("albumId") != "album-1" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		writeJSON(w, models.ActivityStatistics{Comments: 5})
	})
	c := newTestClient(t, mux)
	stats, err := c.GetActivityStatistics(context.Background(), "album-1", nil)
	if err != nil {
		t.Fatalf("GetActivityStatistics error: %v", err)
	}
	if stats.Comments != 5 {
		t.Errorf("expected 5 comments, got %d", stats.Comments)
	}
}

func TestBaseURL(t *testing.T) {
	c := New(&config.Config{BaseURL: "http://immich.local:2283/", APIKey: "key"})
	if c.BaseURL() != "http://immich.local:2283" {
		t.Errorf("expected trailing slash stripped, got %s", c.BaseURL())
	}
}
