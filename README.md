# LOC Counter API

A comprehensive Lines of Code (LOC) counter API written in Go that analyzes GitHub repositories with detailed statistics and language breakdowns.

## Features

- 📊 **Detailed Statistics**: Get comprehensive code metrics including total lines, code lines, comment lines, and blank lines
- 🌐 **Language Detection**: Automatic language detection using GitHub's Linguist (go-enry)
- 🔍 **File-Level Analysis**: Detailed breakdown of every file with sorting by size
- 👤 **Author Attribution**: Optional git blame analysis to see code contributions by author
- 🎯 **Advanced Filtering**: Exclude/include file extensions, directories, and file sizes
- ⚡ **High Performance**: Concurrent processing with goroutines for fast analysis
- 💾 **Result Caching**: In-memory cache with configurable TTL to speed up repeated requests
- 🔒 **Private Repository Support**: Analyze private repositories with GitHub token authentication
- 🌿 **Branch Selection**: Analyze any branch and get a list of all available branches
- 🛡️ **Size Limits**: Enforces 500MB repository size limit
- 🚀 **GitHub Integration**: Fetches repository metadata from GitHub API
- ⚙️ **.env Support**: Easy configuration via environment variables or .env file

> **Note**: Enabling author attribution (`includeAuthors: true`) will clone the full repository history, which increases processing time and download size.

## Installation

```bash
# Clone the repository
git clone https://github.com/stripsiorsior/loc.api.git
cd loc.api

# Install dependencies
go mod download

# Build the application
go build -o loc-api ./cmd/server
```

## Configuration

The API uses environment variables for configuration. You can set them directly or use a `.env` file.

### Environment Variables

- `PORT` - Server port (default: 8080)
- `GITHUB_TOKEN` - GitHub personal access token (optional, used as fallback when token is not provided in request body)
- `CACHE_TTL_MINUTES` - Cache TTL in minutes (default: 60)
- `CORS_ALLOW_ORIGIN` - Allowed CORS origins, comma-separated (default: `*` for all origins)

### Using .env File

Create a `.env` file in the project root:

```bash
cp .env.example .env
```

Edit the `.env` file with your configuration:

```env
PORT=8080
GITHUB_TOKEN=your_github_token_here
CACHE_TTL_MINUTES=60
CORS_ALLOW_ORIGIN=http://localhost:3000,https://yourdomain.com
```

Or set environment variables directly:

```bash
export PORT=8080
export GITHUB_TOKEN=your_github_token_here
export CACHE_TTL_MINUTES=60
export CORS_ALLOW_ORIGIN="http://localhost:3000,https://yourdomain.com"
```

### CORS Configuration Examples

**Allow all origins (default):**

```env
CORS_ALLOW_ORIGIN=*
```

**Allow specific origins:**

```env
CORS_ALLOW_ORIGIN=http://localhost:3000,https://yourdomain.com
```

**Allow single origin:**

```env
CORS_ALLOW_ORIGIN=https://yourdomain.com
```

## Usage

### Start the Server

```bash
./loc-api
```

Or run directly:

```bash
go run ./cmd/server
```

### API Endpoints

#### Health Check

```bash
GET /health
```

#### Analyze Repository

```bash
POST /api/analyze
```

**Request Body:**

```json
{
  "repository": "golang/go",
  "token": "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "branch": "dev",
  "filters": {
    "excludeExtensions": [".md", ".txt"],
    "excludeDirectories": ["vendor", "node_modules", "test"],
    "minFileSize": 100,
    "maxFileSize": 1000000,
    "excludeGenerated": true,
    "excludeVendored": true
  }
}
```

**Request Fields:**

- `repository` (required): Repository identifier in various formats
- `token` (optional): User's GitHub personal access token for private repository access
- `branch` (optional): Branch name to analyze (defaults to repository's default branch)
- `includeAuthors` (optional): Enable author attribution analysis (default: false, increases processing time)
- `filters` (optional): Filtering options for the analysis

**Repository Format:** Accepts multiple formats:

- `owner/repo` (e.g., `golang/go`)
- `https://github.com/owner/repo`
- `github.com/owner/repo`

**Filter Options (all optional):**

- `excludeExtensions`: Array of file extensions to exclude (e.g., `[".md", ".txt"]`)
- `includeExtensions`: Array of file extensions to include (exclusive filter)
- `excludeDirectories`: Array of directory patterns to exclude
- `minFileSize`: Minimum file size in bytes
- `maxFileSize`: Maximum file size in bytes
- `excludeGenerated`: Exclude auto-generated files (default: true)
- `excludeVendored`: Exclude vendored/third-party code (default: true)

**Response:**

```json
{
  "repository": {
    "owner": "golang",
    "name": "go",
    "fullName": "golang/go",
    "description": "The Go programming language",
    "stars": 120000,
    "forks": 17000,
    "size": 250000,
    "defaultBranch": "master",
    "analyzedBranch": "master",
    "branches": ["master", "dev", "release-branch.go1.21", "release-branch.go1.20"],
    "private": false
  },
  "summary": {
    "totalFiles": 8543,
    "totalLines": 2500000,
    "totalCodeLines": 1800000,
    "totalCommentLines": 400000,
    "totalBlankLines": 300000,
    "totalBytes": 85000000
  },
  "languages": {
    "Go": {
      "language": "Go",
      "fileCount": 7234,
      "totalLines": 2200000,
      "codeLines": 1600000,
      "commentLines": 350000,
      "blankLines": 250000,
      "bytes": 75000000,
      "percentage": 88.5
    },
    "Assembly": {
      "language": "Assembly",
      "fileCount": 234,
      "totalLines": 150000,
      "codeLines": 120000,
      "commentLines": 20000,
      "blankLines": 10000,
      "bytes": 5000000,
      "percentage": 6.0
    }
  },
  "authors": {
    "john.doe@example.com": {
      "name": "John Doe",
      "email": "john.doe@example.com",
      "avatarUrl": "https://avatars.githubusercontent.com/u/123456?v=4",
      "totalLines": 850000,
      "codeLines": 650000,
      "commentLines": 120000,
      "blankLines": 80000,
      "fileCount": 234,
      "percentage": 34.0
    },
    "jane.smith@example.com": {
      "name": "Jane Smith",
      "email": "jane.smith@example.com",
      "avatarUrl": "https://avatars.githubusercontent.com/u/789012?v=4",
      "totalLines": 720000,
      "codeLines": 550000,
      "commentLines": 100000,
      "blankLines": 70000,
      "fileCount": 198,
      "percentage": 28.8
    }
  },
  "files": [
    {
      "path": "src/runtime/proc.go",
      "language": "Go",
      "totalLines": 6543,
      "codeLines": 5234,
      "commentLines": 987,
      "blankLines": 322,
      "size": 234567
    }
  ],
  "processingTime": "3.45s"
}
```

### Example Requests

**Basic Analysis:**

```bash
curl -X POST http://localhost:8080/api/analyze \
  -H "Content-Type: application/json" \
  -d '{"repository": "golang/example"}'
```

**With Filters:**

```bash
curl -X POST http://localhost:8080/api/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "repository": "https://github.com/gin-gonic/gin",
    "filters": {
      "excludeExtensions": [".md", ".txt", ".json"],
      "excludeDirectories": ["vendor", "testdata"],
      "minFileSize": 50
    }
  }'
```

**Include Only Specific Extensions:**

```bash
curl -X POST http://localhost:8080/api/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "repository": "facebook/react",
    "filters": {
      "includeExtensions": [".js", ".jsx", ".ts", ".tsx"]
    }
  }'
```

**Analyze Private Repository (with user token):**

```bash
curl -X POST http://localhost:8080/api/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "repository": "myusername/my-private-repo",
    "token": "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  }'
```

**Analyze Specific Branch:**

```bash
curl -X POST http://localhost:8080/api/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "repository": "golang/go",
    "branch": "dev"
  }'
```

**Analyze with Author Attribution:**

```bash
curl -X POST http://localhost:8080/api/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "repository": "golang/go",
    "includeAuthors": true
  }'
```

> **Warning**: Enabling `includeAuthors` clones the full repository history, which significantly increases processing time and download size.

## Supported Languages

The API supports 20+ programming languages with accurate comment detection:

- Go, JavaScript, TypeScript, Java, C, C++, C#, Rust, Swift, Kotlin
- Python, Ruby, PHP, Shell, Bash
- HTML, CSS, SQL, Lua, R
- YAML, TOML, Dockerfile

## Performance

- **Concurrent Processing**: Uses 10 worker goroutines for parallel file analysis
- **Shallow Cloning**: Only clones the latest commit for faster processing
- **Efficient Filtering**: Skips binary, vendored, and generated files automatically
- **Memory Efficient**: Streams file content instead of loading everything into memory

## Error Handling

The API returns appropriate HTTP status codes and error messages:

- `400 Bad Request`: Invalid repository format or filters
- `404 Not Found`: Repository not found on GitHub
- `413 Payload Too Large`: Repository exceeds 500MB limit
- `500 Internal Server Error`: Analysis or cloning failed

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
