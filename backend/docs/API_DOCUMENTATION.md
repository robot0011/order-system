# Order System API Documentation

## Overview

The Order System API is a comprehensive restaurant management system that provides functionality for managing restaurants, tables, menu items, and orders. It includes user authentication and authorization features to ensure secure access to the various endpoints.

## Features

- User Registration and Authentication
- Restaurant Management
- Table Management with QR Code Generation  
- Menu Item Management
- Order Management
- Real-time Order Updates via WebSocket

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. After successful login, users receive both an access token and refresh token. The access token expires in 15 minutes while the refresh token expires in 30 days.

To access protected endpoints, include the access token in the Authorization header:

```
Authorization: Bearer <access_token>
```

## Base URL

The API is served at `http://localhost:3000` by default.

## API Endpoints

### Health Check

- `GET /health` - Check if the API is running

### User Management

- `POST /api/user/register` - Register a new user
- `POST /api/user/login` - Login with username and password
- `GET /api/user/profile` - Get the profile of the authenticated user
- `POST /api/user/refresh` - Refresh access token using refresh token
- `GET /api/user/` - Get all registered users
- `DELETE /api/user/` - Delete the authenticated user

### Restaurant Management

- `POST /api/restaurant/` - Create a new restaurant
- `GET /api/restaurant/` - Get all restaurants for the authenticated user
- `GET /api/restaurant/{id}` - Get a restaurant by ID
- `PUT /api/restaurant/{id}` - Update a restaurant by ID
- `DELETE /api/restaurant/{id}` - Delete a restaurant by ID

### Table Management

- `POST /api/restaurant/{restaurant_id}/table` - Create a new table
- `GET /api/restaurant/{restaurant_id}/table` - Get all tables for a restaurant
- `PUT /api/restaurant/{restaurant_id}/table/{id}` - Update a table
- `DELETE /api/restaurant/{restaurant_id}/table/{id}` - Delete a table
- `GET /api/table` - Get all tables for all restaurants belonging to the user

### Menu Management

- `POST /api/restaurant/{restaurant_id}/menu` - Create a new menu item
- `GET /api/restaurant/{restaurant_id}/menu` - Get all menu items for a restaurant
- `PUT /api/restaurant/{restaurant_id}/menu/{id}` - Update a menu item
- `DELETE /api/restaurant/{restaurant_id}/menu/{id}` - Delete a menu item
- `GET /api/restaurants/{restaurant_id}/menu` - Get public menu items without authentication

### Order Management

- `POST /api/restaurant/{restaurant_id}/order` - Create a new order
- `GET /api/restaurant/{restaurant_id}/order` - Get all orders for a restaurant
- `GET /api/restaurant/{restaurant_id}/order/{id}` - Get a single order by ID
- `PATCH /api/restaurant/{restaurant_id}/order/{id}` - Update order status
- `DELETE /api/restaurant/{restaurant_id}/order/{id}` - Delete an order
- `POST /api/restaurants/{restaurant_id}/order` - Create a public order without authentication
- `GET /api/order` - Get all orders for all restaurants belonging to the user

### WebSocket

- `GET /ws/orders` - WebSocket connection for real-time order updates

## Public vs Protected Endpoints

Some endpoints are publicly accessible while others require authentication:

- Public endpoints: `/health`, `/api/restaurant/{id}` (public restaurant details), `/api/restaurants/{restaurant_id}/menu` (public menu items), `/api/restaurants/{restaurant_id}/order` (create public orders)
- Protected endpoints: Require valid JWT token in Authorization header

## Error Handling

API responses follow a consistent structure. Error responses include a status code and error message. The structure varies slightly based on the endpoint, but typically:

- `200 OK` - Request successful
- `201 Created` - Resource successfully created
- `400 Bad Request` - Invalid input provided
- `401 Unauthorized` - Invalid or missing authentication
- `404 Not Found` - Requested resource not found
- `409 Conflict` - Resource already exists (e.g., username/email taken)
- `500 Internal Server Error` - Unexpected server error

## Data Models

### User
- `id`: Unique identifier
- `username`: Unique username 
- `email`: User's email address
- `role`: User role (e.g., owner, staff)

### Restaurant
- `id`: Unique identifier
- `user_id`: ID of the user who owns the restaurant
- `name`: Name of the restaurant
- `address`: Address of the restaurant
- `phone_number`: Contact number
- `logo_url`: URL to the restaurant logo

### Table
- `id`: Unique identifier
- `restaurant_id`: ID of the associated restaurant
- `table_number`: Table number
- `qr_code_url`: QR code image URL for the table

### Menu Item
- `id`: Unique identifier
- `restaurant_id`: ID of the associated restaurant
- `name`: Name of the menu item
- `description`: Description of the item
- `price`: Price of the item
- `category`: Category (e.g., starter, main, dessert)
- `image_url`: URL to the item image
- `quantity`: Available quantity

### Order
- `id`: Unique identifier
- `table_id`: ID of the table the order is for
- `customer_name`: Name of the customer
- `status`: Order status (pending, preparing, served, completed, cancelled)
- `total_amount`: Total cost of the order
- `order_items`: Array of order items

### Order Item
- `id`: Unique identifier
- `order_id`: ID of the associated order
- `menu_item_id`: ID of the menu item ordered
- `quantity`: Quantity of the item ordered
- `special_instructions`: Special instructions for the item