# Order System

A complete order management system with a Go backend and frontend interface.

## Project Structure

```
order-system/
├── backend/          # Go backend application
│   ├── main.go       # Main application entry point
│   ├── go.mod        # Go module definition
│   ├── go.sum        # Go module checksums
│   ├── .env.example  # Environment variables example
│   ├── database/     # Database configuration and models
│   ├── handler/      # API request handlers
│   ├── models/       # Data models
│   └── utils/        # Utility functions
└── frontend/         # Frontend application (to be implemented)
```

## Backend Setup

### Prerequisites

- Go 1.24.5 or higher
- PostgreSQL database

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd order-system/backend
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up environment variables:
   - Copy `.env.example` to `.env`
   - Update the values in `.env` with your local configuration

4. Run the application:
   ```bash
   go run main.go
   ```

The backend server will start on `http://localhost:8080` (or the port specified in your environment variables).

### Database Configuration

The application uses PostgreSQL with the following environment variable for connection:
- `DATABASE_URL`: PostgreSQL connection string in the format `postgres://username:password@localhost:5432/database_name?sslmode=disable`

## Frontend Setup

The frontend directory is prepared but not yet implemented. Instructions will be added when the frontend is developed.

## API Documentation

API documentation is available at `/swagger/index.html` when the application is running.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request