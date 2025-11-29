# Project Architecture

## Overview

The Order System Backend is built using Go with the Fiber web framework and follows a modular architecture with clear separation of concerns.

## Tech Stack

- **Language**: Go (Golang)
- **Web Framework**: Fiber
- **Database**: GORM with support for MySQL/PostgreSQL/SQLite
- **Authentication**: JWT (JSON Web Tokens)
- **Documentation**: Swagger
- **WebSocket**: Real-time order updates
- **QR Code Generation**: For table-specific links

## Project Structure

```
backend/
├── docs/              # API documentation
├── handler/           # Request handlers
├── models/            # Database models
├── database/          # Database configuration
├── utils/             # Utility functions
├── constants/         # Constant values
├── main.go           # Main entrypoint
├── router.go         # Route definitions
├── docs/             # Auto-generated Swagger docs
├── go.mod            # Go module definition
└── go.sum            # Dependency checksums
```

## Layered Architecture

### 1. Router Layer (`router.go`)

Handles route registration and middleware setup. Defines all API endpoints and links them to appropriate handlers.

### 2. Handler Layer (`handler/`)

Contains business logic for handling HTTP requests. Each handler function:
- Parses incoming requests
- Validates input data
- Interacts with the database through models
- Returns appropriate responses

### 3. Model Layer (`models/`)

Defines database schemas using GORM. Contains:
- Database entity definitions
- Relationships between entities
- Model-specific methods

### 4. Database Layer (`database/`)

Manages database connections and configuration. Handles:
- Database initialization
- Connection pooling
- Migration setup

### 5. Utils Layer (`utils/`)

Contains reusable utility functions such as:
- QR code generation
- Status mapping utilities
- Helper functions

## Authentication Flow

1. **Registration**: New users register with username, password, email, and role
2. **Login**: Users authenticate with credentials and receive JWT tokens
3. **Access Token**: Short-lived token (15 minutes) for API requests
4. **Refresh Token**: Long-lived token (30 days) to obtain new access tokens
5. **Authorization**: Protected endpoints verify JWT tokens in the Authorization header

## Database Schema

### Users
- User accounts with role-based access
- Related to restaurants they own

### Restaurants
- Owned by users
- Associated with tables and menu items

### Tables
- Belong to restaurants
- Include QR code URLs for direct access

### Menu Items
- Belong to restaurants
- Include pricing, description, and availability

### Orders
- Associated with tables
- Include order items and status tracking

### Order Items
- Links orders to menu items
- Contains quantities and special instructions

## WebSocket Integration

The system includes WebSocket support for real-time order updates:
- Clients can connect to `/ws/orders`
- Order status changes are broadcast to connected clients
- Provides live updates to kitchen displays and admin panels

## Error Handling

- Consistent error response format
- Appropriate HTTP status codes
- Detailed error messages where appropriate
- Graceful degradation for system errors

## Security Considerations

- JWT tokens with proper expiration
- Database input validation
- Parameterized queries to prevent SQL injection
- Role-based access control
- Secure password hashing with bcrypt

## Performance Optimizations

- Database indexing where appropriate
- Efficient query patterns
- Connection pooling
- Caching of frequently accessed data
- Proper resource cleanup