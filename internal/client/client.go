package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/ryanmac8/ImmichMCP/internal/config"
	"github.com/ryanmac8/ImmichMCP/internal/models"
)

// Client wraps all Immich API operations.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New creates a new Immich API client.
func New(cfg *config.Config) *Client {
	return &Client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string { return c.baseURL }

// ─── Health ────────────────────────────────────────────────────────────────

// ServerInfo holds Immich server version/build information.
type ServerInfo struct {
	Version    string `json:"version"`
	VersionURL string `json:"versionUrl"`
	Licensed   bool   `json:"licensed"`
	Build      string `json:"build"`
	Nodejs     string `json:"nodejs"`
	Ffmpeg     string `json:"ffmpeg"`
	Exiftool   string `json:"exiftool"`
	Libvips    string `json:"libvips"`
}

// ServerFeatures holds Immich feature flags.
type ServerFeatures struct {
	Trash               bool `json:"trash"`
	Map                 bool `json:"map"`
	ReverseGeocoding    bool `json:"reverseGeocoding"`
	Import              bool `json:"import"`
	Sidecar             bool `json:"sidecar"`
	Search              bool `json:"search"`
	FacialRecognition   bool `json:"facialRecognition"`
	Oauth               bool `json:"oauth"`
	OauthAutoLaunch     bool `json:"oauthAutoLaunch"`
	PasswordLogin       bool `json:"passwordLogin"`
	ConfigFile          bool `json:"configFile"`
	DuplicateDetection  bool `json:"duplicateDetection"`
	Email               bool `json:"email"`
	SmartSearch         bool `json:"smartSearch"`
}

func (c *Client) Ping(ctx context.Context) (*ServerInfo, error) {
	var info ServerInfo
	if err := c.get(ctx, "api/server/about", &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *Client) GetFeatures(ctx context.Context) (*ServerFeatures, error) {
	var features ServerFeatures
	if err := c.get(ctx, "api/server/features", &features); err != nil {
		return nil, err
	}
	return &features, nil
}

// ─── Assets ───────────────────────────────────────────────────────────────

func (c *Client) GetAssets(ctx context.Context, size *int, isFavorite, isArchived, isTrashed *bool, updatedAfter, updatedBefore *time.Time, userID *string) ([]models.Asset, error) {
	q := url.Values{}
	if size != nil {
		q.Set("size", fmt.Sprintf("%d", *size))
	}
	if isFavorite != nil {
		q.Set("isFavorite", boolStr(*isFavorite))
	}
	if isArchived != nil {
		q.Set("isArchived", boolStr(*isArchived))
	}
	if isTrashed != nil {
		q.Set("isTrashed", boolStr(*isTrashed))
	}
	if updatedAfter != nil {
		q.Set("updatedAfter", updatedAfter.Format(time.RFC3339))
	}
	if updatedBefore != nil {
		q.Set("updatedBefore", updatedBefore.Format(time.RFC3339))
	}
	if userID != nil {
		q.Set("userId", *userID)
	}
	endpoint := "api/assets"
	if len(q) > 0 {
		endpoint += "?" + q.Encode()
	}
	var assets []models.Asset
	if err := c.get(ctx, endpoint, &assets); err != nil {
		return nil, err
	}
	return assets, nil
}

func (c *Client) GetAsset(ctx context.Context, id string) (*models.Asset, error) {
	var asset models.Asset
	if err := c.get(ctx, "api/assets/"+id, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

func (c *Client) UpdateAsset(ctx context.Context, id string, req models.AssetUpdateRequest) (*models.Asset, error) {
	var asset models.Asset
	if err := c.put(ctx, "api/assets/"+id, req, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

func (c *Client) BulkUpdateAssets(ctx context.Context, req models.AssetBulkUpdateRequest) error {
	return c.putNoResponse(ctx, "api/assets", req)
}

func (c *Client) DeleteAssets(ctx context.Context, ids []string, force bool) error {
	body := map[string]interface{}{"ids": ids, "force": force}
	return c.deleteWithBody(ctx, "api/assets", body)
}

func (c *Client) GetAssetStatistics(ctx context.Context) (*models.AssetStatistics, error) {
	var stats models.AssetStatistics
	if err := c.get(ctx, "api/assets/statistics", &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

func (c *Client) UploadAsset(ctx context.Context, fileContent []byte, fileName, deviceAssetID string, deviceModifiedAt time.Time, isFavorite, isArchived *bool) (*models.Asset, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("assetData", fileName)
	if err != nil {
		return nil, err
	}
	if _, err = fw.Write(fileContent); err != nil {
		return nil, err
	}

	fields := map[string]string{
		"deviceAssetId":    deviceAssetID,
		"deviceId":         "mcp-server",
		"deviceModifiedAt": deviceModifiedAt.Format(time.RFC3339),
		"fileCreatedAt":    time.Now().UTC().Format(time.RFC3339),
		"fileModifiedAt":   time.Now().UTC().Format(time.RFC3339),
	}
	if isFavorite != nil {
		fields["isFavorite"] = boolStr(*isFavorite)
	}
	if isArchived != nil {
		fields["isArchived"] = boolStr(*isArchived)
	}
	for k, v := range fields {
		if err = w.WriteField(k, v); err != nil {
			return nil, err
		}
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("api/assets"), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var asset models.Asset
	if err = json.NewDecoder(resp.Body).Decode(&asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

func (c *Client) GetAssetDownloadInfo(id string, originalFileName *string) models.AssetDownloadInfo {
	base := c.baseURL
	thumb := base + "/api/assets/" + id + "/thumbnail"
	preview := thumb + "?size=preview"
	return models.AssetDownloadInfo{
		ID:               id,
		OriginalFileName: originalFileName,
		OriginalURL:      base + "/api/assets/" + id + "/original",
		ThumbnailURL:     &thumb,
		PreviewURL:       &preview,
	}
}

// ─── Search ───────────────────────────────────────────────────────────────

func (c *Client) SearchMetadata(ctx context.Context, req models.MetadataSearchRequest) (*models.SearchAssetResult, error) {
	var result models.SearchResult
	if err := c.post(ctx, "api/search/metadata", req, &result); err != nil {
		return nil, err
	}
	return &result.Assets, nil
}

func (c *Client) SearchSmart(ctx context.Context, req models.SmartSearchRequest) (*models.SearchAssetResult, error) {
	var result models.SearchResult
	if err := c.post(ctx, "api/search/smart", req, &result); err != nil {
		return nil, err
	}
	return &result.Assets, nil
}

func (c *Client) SearchExplore(ctx context.Context) ([]models.ExploreData, error) {
	var data []models.ExploreData
	if err := c.get(ctx, "api/search/explore", &data); err != nil {
		return nil, err
	}
	return data, nil
}

// ─── Albums ───────────────────────────────────────────────────────────────

func (c *Client) GetAlbums(ctx context.Context, shared *bool, assetID *string) ([]models.Album, error) {
	q := url.Values{}
	if shared != nil {
		q.Set("shared", boolStr(*shared))
	}
	if assetID != nil {
		q.Set("assetId", *assetID)
	}
	endpoint := "api/albums"
	if len(q) > 0 {
		endpoint += "?" + q.Encode()
	}
	var albums []models.Album
	if err := c.get(ctx, endpoint, &albums); err != nil {
		return nil, err
	}
	return albums, nil
}

func (c *Client) GetAlbum(ctx context.Context, id string, withoutAssets *bool) (*models.Album, error) {
	endpoint := "api/albums/" + id
	if withoutAssets != nil && *withoutAssets {
		endpoint += "?withoutAssets=true"
	}
	var album models.Album
	if err := c.get(ctx, endpoint, &album); err != nil {
		return nil, err
	}
	return &album, nil
}

func (c *Client) CreateAlbum(ctx context.Context, req models.AlbumCreateRequest) (*models.Album, error) {
	var album models.Album
	if err := c.post(ctx, "api/albums", req, &album); err != nil {
		return nil, err
	}
	return &album, nil
}

func (c *Client) UpdateAlbum(ctx context.Context, id string, req models.AlbumUpdateRequest) (*models.Album, error) {
	var album models.Album
	if err := c.patch(ctx, "api/albums/"+id, req, &album); err != nil {
		return nil, err
	}
	return &album, nil
}

func (c *Client) DeleteAlbum(ctx context.Context, id string) error {
	return c.delete(ctx, "api/albums/"+id)
}

func (c *Client) AddAssetsToAlbum(ctx context.Context, albumID string, assetIDs []string) ([]models.BulkIDResponse, error) {
	var result []models.BulkIDResponse
	if err := c.put(ctx, "api/albums/"+albumID+"/assets", map[string]interface{}{"ids": assetIDs}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) RemoveAssetsFromAlbum(ctx context.Context, albumID string, assetIDs []string) ([]models.BulkIDResponse, error) {
	var result []models.BulkIDResponse
	if err := c.deleteWithBodyResult(ctx, "api/albums/"+albumID+"/assets", map[string]interface{}{"ids": assetIDs}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetAlbumStatistics(ctx context.Context) (*models.AlbumStatistics, error) {
	var stats models.AlbumStatistics
	if err := c.get(ctx, "api/albums/statistics", &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// ─── People ───────────────────────────────────────────────────────────────

func (c *Client) GetPeople(ctx context.Context, withHidden *bool) (*models.PeopleResponse, error) {
	endpoint := "api/people"
	if withHidden != nil {
		endpoint += "?withHidden=" + boolStr(*withHidden)
	}
	var result models.PeopleResponse
	if err := c.get(ctx, endpoint, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetPerson(ctx context.Context, id string) (*models.Person, error) {
	var person models.Person
	if err := c.get(ctx, "api/people/"+id, &person); err != nil {
		return nil, err
	}
	return &person, nil
}

func (c *Client) UpdatePerson(ctx context.Context, id string, req models.PersonUpdateRequest) (*models.Person, error) {
	var person models.Person
	if err := c.put(ctx, "api/people/"+id, req, &person); err != nil {
		return nil, err
	}
	return &person, nil
}

func (c *Client) MergePeople(ctx context.Context, targetID string, sourceIDs []string) ([]models.BulkIDResponse, error) {
	var result []models.BulkIDResponse
	if err := c.post(ctx, "api/people/"+targetID+"/merge", map[string]interface{}{"ids": sourceIDs}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetPersonAssets(ctx context.Context, personID string) ([]models.Asset, error) {
	var assets []models.Asset
	if err := c.get(ctx, "api/people/"+personID+"/assets", &assets); err != nil {
		return nil, err
	}
	return assets, nil
}

// ─── Tags ─────────────────────────────────────────────────────────────────

func (c *Client) GetTags(ctx context.Context) ([]models.Tag, error) {
	var tags []models.Tag
	if err := c.get(ctx, "api/tags", &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

func (c *Client) GetTag(ctx context.Context, id string) (*models.Tag, error) {
	var tag models.Tag
	if err := c.get(ctx, "api/tags/"+id, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

func (c *Client) CreateTag(ctx context.Context, req models.TagCreateRequest) (*models.Tag, error) {
	var tag models.Tag
	if err := c.post(ctx, "api/tags", req, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

func (c *Client) UpdateTag(ctx context.Context, id string, req models.TagUpdateRequest) (*models.Tag, error) {
	var tag models.Tag
	if err := c.put(ctx, "api/tags/"+id, req, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

func (c *Client) DeleteTag(ctx context.Context, id string) error {
	return c.delete(ctx, "api/tags/"+id)
}

func (c *Client) TagAssets(ctx context.Context, tagID string, assetIDs []string) ([]models.BulkIDResponse, error) {
	var result []models.BulkIDResponse
	if err := c.put(ctx, "api/tags/"+tagID+"/assets", map[string]interface{}{"ids": assetIDs}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) UntagAssets(ctx context.Context, tagID string, assetIDs []string) ([]models.BulkIDResponse, error) {
	var result []models.BulkIDResponse
	if err := c.deleteWithBodyResult(ctx, "api/tags/"+tagID+"/assets", map[string]interface{}{"ids": assetIDs}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ─── Shared Links ─────────────────────────────────────────────────────────

func (c *Client) GetSharedLinks(ctx context.Context) ([]models.SharedLink, error) {
	var links []models.SharedLink
	if err := c.get(ctx, "api/shared-links", &links); err != nil {
		return nil, err
	}
	return links, nil
}

func (c *Client) GetSharedLink(ctx context.Context, id string) (*models.SharedLink, error) {
	var link models.SharedLink
	if err := c.get(ctx, "api/shared-links/"+id, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

func (c *Client) CreateSharedLink(ctx context.Context, req models.SharedLinkCreateRequest) (*models.SharedLink, error) {
	var link models.SharedLink
	if err := c.post(ctx, "api/shared-links", req, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

func (c *Client) UpdateSharedLink(ctx context.Context, id string, req models.SharedLinkUpdateRequest) (*models.SharedLink, error) {
	var link models.SharedLink
	if err := c.patch(ctx, "api/shared-links/"+id, req, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

func (c *Client) DeleteSharedLink(ctx context.Context, id string) error {
	return c.delete(ctx, "api/shared-links/"+id)
}

// ─── Activities ───────────────────────────────────────────────────────────

func (c *Client) GetActivities(ctx context.Context, albumID string, assetID, actType, level *string) ([]models.Activity, error) {
	q := url.Values{}
	q.Set("albumId", albumID)
	if assetID != nil {
		q.Set("assetId", *assetID)
	}
	if actType != nil {
		q.Set("type", *actType)
	}
	if level != nil {
		q.Set("level", *level)
	}
	var activities []models.Activity
	if err := c.get(ctx, "api/activities?"+q.Encode(), &activities); err != nil {
		return nil, err
	}
	return activities, nil
}

func (c *Client) CreateActivity(ctx context.Context, req models.ActivityCreateRequest) (*models.Activity, error) {
	var activity models.Activity
	if err := c.post(ctx, "api/activities", req, &activity); err != nil {
		return nil, err
	}
	return &activity, nil
}

func (c *Client) DeleteActivity(ctx context.Context, id string) error {
	return c.delete(ctx, "api/activities/"+id)
}

func (c *Client) GetActivityStatistics(ctx context.Context, albumID string, assetID *string) (*models.ActivityStatistics, error) {
	q := url.Values{}
	q.Set("albumId", albumID)
	if assetID != nil {
		q.Set("assetId", *assetID)
	}
	var stats models.ActivityStatistics
	if err := c.get(ctx, "api/activities/statistics?"+q.Encode(), &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// ─── HTTP helpers ─────────────────────────────────────────────────────────

func (c *Client) url(path string) string {
	return c.baseURL + "/" + path
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.url(path), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Client) do(req *http.Request, out interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) get(ctx context.Context, path string, out interface{}) error {
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) post(ctx context.Context, path string, body, out interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := c.newRequest(ctx, http.MethodPost, path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *Client) put(ctx context.Context, path string, body, out interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := c.newRequest(ctx, http.MethodPut, path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *Client) putNoResponse(ctx context.Context, path string, body interface{}) error {
	return c.put(ctx, path, body, nil)
}

func (c *Client) patch(ctx context.Context, path string, body, out interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := c.newRequest(ctx, http.MethodPatch, path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *Client) delete(ctx context.Context, path string) error {
	req, err := c.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

func (c *Client) deleteWithBody(ctx context.Context, path string, body interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := c.newRequest(ctx, http.MethodDelete, path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, nil)
}

func (c *Client) deleteWithBodyResult(ctx context.Context, path string, body, out interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := c.newRequest(ctx, http.MethodDelete, path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// fileExt is used to validate file extensions (unused directly but kept for future use).
var _ = filepath.Ext
