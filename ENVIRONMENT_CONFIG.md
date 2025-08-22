# Environment Configuration for Go Backend

## Overview
This document explains how to configure environment variables for your Go backend, including CORS settings.

## Environment Variables

### Server Configuration
```bash
# Server port (default: 3000)
PORT=3000
```

### CORS Configuration
```bash
# Comma-separated list of allowed origins
# For development:
CORS_ORIGINS=http://localhost:5173,http://localhost:3000,http://localhost:8080

# For production:
CORS_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

### Database Configuration (if needed)
```bash
DB_HOST=localhost
DB_PORT=5432
DB_NAME=order_system
DB_USER=postgres
DB_PASSWORD=password
```

### JWT Configuration (if needed)
```bash
JWT_SECRET=your-secret-key-here
JWT_REFRESH_SECRET=your-refresh-secret-key-here
```

## How to Use

### Option 1: Environment File
Create a `.env` file in your project root:
```bash
# .env
PORT=3000
CORS_ORIGINS=http://localhost:5173,http://localhost:3000
```

### Option 2: System Environment Variables
Set environment variables in your system:
```bash
# Windows (PowerShell)
$env:PORT="3000"
$env:CORS_ORIGINS="http://localhost:5173,http://localhost:3000"

# Windows (Command Prompt)
set PORT=3000
set CORS_ORIGINS=http://localhost:5173,http://localhost:3000

# Linux/Mac
export PORT=3000
export CORS_ORIGINS="http://localhost:5173,http://localhost:3000"
```

### Option 3: Docker (if using containers)
```yaml
# docker-compose.yml
environment:
  - PORT=3000
  - CORS_ORIGINS=http://localhost:5173,http://localhost:3000
```

## CORS Configuration Details

The CORS middleware is configured with the following settings:

- **AllowOrigins**: List of allowed origins (frontend URLs)
- **AllowHeaders**: Allowed HTTP headers
- **AllowMethods**: Allowed HTTP methods
- **AllowCredentials**: Allows cookies and authentication headers
- **ExposeHeaders**: Headers exposed to the frontend
- **MaxAge**: How long preflight requests can be cached (24 hours)

## Frontend Integration

Your frontend is configured to make requests to:
- Login: `POST /api/user/login`
- Register: `POST /api/user/register`
- Profile: `GET /api/user/profile`

Make sure your frontend is running on one of the allowed origins specified in `CORS_ORIGINS`.

## Testing CORS

To test if CORS is working:

1. Start your Go backend
2. Start your frontend (should be on one of the allowed origins)
3. Try to make a request from the frontend to the backend
4. Check the browser's Network tab for CORS errors

If you see CORS errors, verify that:
- Your frontend URL is in the `CORS_ORIGINS` list
- The CORS middleware is loaded before your routes
- Your environment variables are set correctly
