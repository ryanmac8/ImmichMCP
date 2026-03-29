# ImmichMCP

A [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server for [Immich](https://immich.app/) — the self-hosted photo and video management solution. This server provides a first-class AI interface to manage your Immich library.

[![CI](https://github.com/ryanmac8/ImmichMCP/actions/workflows/ci.yml/badge.svg)](https://github.com/ryanmac8/ImmichMCP/actions/workflows/ci.yml)
[![Docker](https://github.com/ryanmac8/ImmichMCP/actions/workflows/docker.yml/badge.svg)](https://github.com/ryanmac8/ImmichMCP/actions/workflows/docker.yml)

## Features

- **Asset Management**: Search, browse, upload, update, and delete photos/videos
- **Smart Search**: ML-powered semantic search using CLIP (e.g., "sunset at the beach")
- **Metadata Search**: Filter by date, location, camera, people, and more
- **Albums**: Create, manage, and share photo albums
- **People**: View and manage face recognition clusters
- **Tags**: Organize assets with custom tags
- **Shared Links**: Create shareable URLs for albums and assets
- **Activities**: Add comments and likes to albums/assets
- **Out-of-band Upload**: Upload files directly via HTTP without going through the MCP protocol

## Requirements

- Go 1.23+ (for building from source)
- Immich server instance
- Immich API key

## Quick Start

### Docker (recommended)

```bash
docker run -d \
  -e IMMICH_BASE_URL=http://your-immich-server:2283 \
  -e IMMICH_API_KEY=your-api-key \
  -p 5000:5000 \
  ghcr.io/ryanmac8/immichmcp:main
```

### Build from Source

```bash
git clone https://github.com/ryanmac8/ImmichMCP.git
cd ImmichMCP
go build -o immichmcp ./cmd/immichmcp
```

## Configuration

| Environment Variable | Required | Default | Description |
|---|---|---|---|
| `IMMICH_BASE_URL` | Yes | — | Base URL of your Immich server (e.g. `http://192.168.1.10:2283`) |
| `IMMICH_API_KEY` | Yes | — | API key from Immich → Account Settings → API Keys |
| `MCP_PORT` | No | `5000` | HTTP port for the MCP server |
| `MAX_PAGE_SIZE` | No | `100` | Maximum number of results per page |
| `DOWNLOAD_MODE` | No | `url` | Asset download mode: `url` or `inline` |
| `MCP_PUBLIC_URL` | No | — | Public base URL (used in upload instructions) |

## Transport Modes

### HTTP + SSE (default)

The server starts on `MCP_PORT` (default 5000) and exposes:
- `GET /sse` — MCP SSE transport endpoint
- `POST /upload/{sessionId}` — Out-of-band file upload
- `GET /health` — Health check

### stdio (Claude Desktop)

```bash
./immichmcp --stdio
```

### Claude Desktop Configuration

```json
{
  "mcpServers": {
    "immich": {
      "command": "/path/to/immichmcp",
      "args": ["--stdio"],
      "env": {
        "IMMICH_BASE_URL": "http://your-immich-server:2283",
        "IMMICH_API_KEY": "your-api-key"
      }
    }
  }
}
```

## Available Tools

| Tool | Description |
|---|---|
| `immich_ping` | Check server connectivity |
| `immich_capabilities` | List available features |
| `immich_assets_list` | List/search assets |
| `immich_assets_get` | Get asset details |
| `immich_assets_exif` | Get EXIF metadata |
| `immich_assets_upload` | Upload a new asset |
| `immich_assets_update` | Update asset metadata |
| `immich_assets_bulk_update` | Update multiple assets |
| `immich_assets_delete` | Delete assets |
| `immich_assets_statistics` | Asset count statistics |
| `immich_assets_download_original` | Download original file |
| `immich_assets_download_thumbnail` | Download thumbnail |
| `immich_assets_upload_init` | Initialize out-of-band upload |
| `immich_assets_upload_status` | Check upload status |
| `immich_albums_list` | List albums |
| `immich_albums_get` | Get album details |
| `immich_albums_create` | Create album |
| `immich_albums_update` | Update album |
| `immich_albums_assets_add` | Add assets to album |
| `immich_albums_assets_remove` | Remove assets from album |
| `immich_albums_delete` | Delete album |
| `immich_albums_statistics` | Album statistics |
| `immich_search_metadata` | Metadata search |
| `immich_search_smart` | Semantic/CLIP search |
| `immich_search_explore` | Explore suggested search terms |
| `immich_people_list` | List recognized people |
| `immich_people_get` | Get person details |
| `immich_people_update` | Update person info |
| `immich_people_merge` | Merge duplicate faces |
| `immich_people_assets` | Assets containing a person |
| `immich_tags_list` | List tags |
| `immich_tags_create` | Create tag |
| `immich_tags_update` | Update tag |
| `immich_tags_delete` | Delete tag |
| `immich_tags_assets_add` | Tag assets |
| `immich_tags_assets_remove` | Untag assets |
| `immich_shared_links_list` | List shared links |
| `immich_shared_links_get` | Get shared link |
| `immich_shared_links_create` | Create shared link |
| `immich_shared_links_update` | Update shared link |
| `immich_shared_links_delete` | Delete shared link |
| `immich_activities_list` | List activities |
| `immich_activities_create` | Create comment/like |
| `immich_activities_delete` | Delete activity |
| `immich_activities_statistics` | Activity statistics |

## License

MIT
