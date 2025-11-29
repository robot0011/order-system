# Hot Reload with Air

This project is configured with Air, a live reload tool for Go applications that automatically rebuilds and restarts your server when files change.

## Prerequisites

Air should already be installed. If you need to install it manually:

```bash
go install github.com/air-verse/air@latest
```

## Configuration

The project includes an `.air.toml` configuration file that:

- Monitors Go files (`.go`), templates (`.tpl`, `.tmpl`), and HTML files
- Excludes unnecessary directories like `node_modules`, `tmp`, `vendor`, etc.
- Builds to `./tmp/main.exe` (Windows compatible)
- Automatically restarts the server on file changes

## Usage

### Start Development with Hot Reload

```bash
air
```

This will:
- Start watching your files for changes
- Build the application initially 
- Start the Fiber server on port 3000
- Automatically rebuild and restart when you modify any `.go` files

### Regular Development (without hot reload)

```bash
go run main.go
```

### Build for Production

```bash
go build -o order-system.exe .
```

This creates a single binary executable file that can run independently without requiring Go to be installed on the target system.

### Cross-compilation for Different Platforms

You can build for different operating systems using Go's cross-compilation:

**For Linux:**
```bash
GOOS=linux GOARCH=amd64 go build -o order-system .
```

**For macOS:**
```bash
GOOS=darwin GOARCH=amd64 go build -o order-system .
```

**For Windows (64-bit):**
```bash
GOOS=windows GOARCH=amd64 go build -o order-system.exe .
```

### Optimized Build for Production

For a smaller binary size and more optimized performance:

```bash
go build -ldflags="-s -w" -o order-system.exe .
```

The `-s` flag strips the symbol table and debug information, and `-w` removes the DWARF symbol table.

### Building and Running the Binary

After building:

1. **Windows:**
   ```
   order-system.exe
   ```

2. **Linux/macOS:**
   ```
   ./order-system
   ```

Make sure your environment variables (like DATABASE_URL, PORT, etc.) are set in the target environment before running the binary.

## What Air Watches

Air monitors these file types for changes:
- `.go` files (all Go source code)
- `.tpl` and `.tmpl` files (Go templates)
- `.html` files (HTML templates)

Air ignores these directories:
- `tmp/` (build directory)
- `vendor/` (dependencies)
- `frontend-app/node_modules/` (Node.js dependencies) 
- `frontend-app/dist/` (frontend build output)
- `testdata/` (test data)

## Benefits

- **Faster Development**: No need to manually restart the server after code changes
- **Automatic Compilation**: Instantly see compilation errors
- **Live Testing**: Test API endpoints immediately after making changes
- **Better Workflow**: Focus on coding without interruption

## Troubleshooting

If Air doesn't start or has issues:

1. Make sure you're in the project root directory
2. Check that Air is installed: `air -v`
3. Verify the `.air.toml` configuration file exists
4. Ensure Go modules are properly initialized: `go mod tidy`

The server runs on `http://localhost:3000` by default (or the PORT environment variable if set).
