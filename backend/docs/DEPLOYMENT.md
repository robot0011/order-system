# Backend Deployment Guide

## Prerequisites

Before deploying the backend, ensure your environment has:

- Go 1.20 or higher
- Git
- Access to a database (MySQL, PostgreSQL, or SQLite)
- Environment variables configured for production

## Environment Variables

Create a `.env` file or configure environment variables:

```env
PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=order_system
DB_TYPE=postgresql  # or mysql, sqlite

# JWT Configuration
JWT_SECRET=your_secure_jwt_secret
JWT_REFRESH_SECRET=your_secure_refresh_secret

# CORS Configuration
CORS_ORIGINS=your-frontend-domain.com,another-domain.com  # For production, be specific
# For development only:
# CORS_ORIGINS=*

# Database specific settings
# For MySQL
DB_TYPE=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=your_mysql_user
DB_PASSWORD=your_mysql_password
DB_NAME=order_system

# For SQLite (no additional configuration needed)
DB_TYPE=sqlite
DB_NAME=order_system.db
```

## Database Setup

### PostgreSQL (Recommended for Production)
```bash
# Install PostgreSQL and create a database
CREATE DATABASE order_system;
CREATE USER order_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE order_system TO order_user;
```

### MySQL
```bash
# Install MySQL and create a database
CREATE DATABASE order_system;
CREATE USER 'order_user'@'localhost' IDENTIFIED BY 'secure_password';
GRANT ALL PRIVILEGES ON order_system.* TO 'order_user'@'localhost';
FLUSH PRIVILEGES;
```

### SQLite (Development Only)
No additional setup required. The database file will be created automatically.

## Building the Application

### Local Build
```bash
cd backend
go build -o order-system .
```

### Cross-compilation
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o order-system-linux .

# Windows
GOOS=windows GOARCH=amd64 go build -o order-system.exe .

# macOS
GOOS=darwin GOARCH=amd64 go build -o order-system-mac .
```

## Running the Application

### Direct Execution
```bash
# Set environment variables
export JWT_SECRET=your_secure_jwt_secret
export DB_TYPE=postgresql
export DB_HOST=localhost
# ... other environment variables

# Run the application
./order-system
```

### Using a Process Manager (Recommended)
#### PM2 (Node.js Process Manager)
```bash
npm install -g pm2
pm2 start main.go --name "order-system" --interpreter="go run"
```

#### Systemd (Linux)
Create `/etc/systemd/system/order-system.service`:
```ini
[Unit]
Description=Order System Backend
After=network.target

[Service]
Type=simple
User=appuser
WorkingDirectory=/path/to/backend
ExecStart=/path/to/order-system
Restart=always
EnvironmentFile=/path/to/backend/.env

[Install]
WantedBy=multi-user.target
```

Then:
```bash
sudo systemctl daemon-reload
sudo systemctl enable order-system
sudo systemctl start order-system
```

## Docker Deployment

### Building the Docker Image
```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/.env .  # If you have an .env file

EXPOSE 3000
CMD ["./main"]
```

```bash
docker build -t order-system:latest .
docker run -d --name order-system -p 3000:3000 order-system:latest
```

### Docker Compose
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "3000:3000"
    environment:
      - DB_TYPE=postgresql
      - DB_HOST=db
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=order_system
      - JWT_SECRET=your_secure_jwt_secret
    depends_on:
      - db

  db:
    image: postgres:13-alpine
    environment:
      - POSTGRES_DB=order_system
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## Production Considerations

### Security
- Use HTTPS in production
- Set secure JWT secrets
- Configure CORS properly (not using `*` in production)
- Implement rate limiting
- Use a reverse proxy (nginx)

### Performance
- Use a production-ready database (not SQLite)
- Enable database connection pooling
- Use Redis for session management and caching
- Monitor application performance

### Monitoring
- Implement proper logging
- Use monitoring tools (Prometheus, Grafana)
- Set up health checks
- Alerting for critical failures

### Infrastructure
- Use a load balancer for high availability
- Implement backup strategies for database
- Use environment-specific configurations
- Implement CI/CD pipeline

## API Documentation in Production

The Swagger documentation will be available at:
- `http://your-domain:3000/swagger/index.html`

## Troubleshooting

### Common Issues
- **Database Connection**: Verify environment variables and database accessibility
- **JWT Errors**: Check JWT secrets and token expiration settings
- **CORS Issues**: Ensure CORS origins are properly configured
- **Port Already in Use**: Check if the application is already running

### Logging
Check application logs for errors:
```bash
# If using systemd
journalctl -u order-system -f

# If running directly, logs are printed to stdout/stderr
```

## Health Checks

The application provides a health check endpoint:
- `GET /health` - Returns application status

Use this for load balancer health checks and monitoring.