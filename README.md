# Vibanda Village Backend

A comprehensive REST API backend for Vibanda Village restaurant management system built with Go, Gin, GORM, and MongoDB.

## Features

- **Authentication & Authorization**: JWT-based authentication with role-based access control
- **User Management**: Admin, Manager, and Staff roles
- **Product Management**: Food and drink items with categories and inventory
- **Order Management**: Customer orders with payment tracking
- **Event Management**: Restaurant events and capacity management
- **Reservation System**: Table reservations with guest management
- **Swagger Documentation**: Auto-generated API documentation
- **CORS Support**: Cross-origin resource sharing configuration

## Tech Stack

- **Go**: Programming language
- **Gin**: Web framework
- **GORM**: ORM for MongoDB
- **MongoDB**: NoSQL database
- **JWT**: JSON Web Tokens for authentication
- **Swagger**: API documentation

## Project Structure

```
vibanda-village-backend/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go          # Configuration management
│   ├── database/
│   │   └── database.go        # Database connection
│   ├── handlers/
│   │   ├── auth.go            # Authentication handlers
│   │   ├── users.go           # User management handlers
│   │   ├── products.go        # Product management handlers
│   │   ├── orders.go          # Order management handlers
│   │   ├── events.go          # Event management handlers
│   │   ├── reservations.go    # Reservation handlers
│   │   └── common.go          # Common utilities
│   ├── middleware/
│   │   └── auth.go            # Authentication middleware
│   ├── models/
│   │   ├── user.go            # User model
│   │   ├── product.go         # Product model
│   │   ├── order.go           # Order model
│   │   ├── event.go           # Event model
│   │   └── reservation.go     # Reservation model
│   └── routes/
│       └── routes.go          # Route definitions
├── pkg/
│   └── utils/
│       ├── jwt.go             # JWT utilities
│       └── password.go        # Password hashing utilities
├── .env.example               # Environment variables template
├── go.mod                     # Go modules
└── README.md                  # This file
```

## Prerequisites

- Go 1.21 or higher
- MongoDB 4.0 or higher
- Git

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd vibanda-village-backend
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   ```

   Edit `.env` with your configuration:
   ```env
   PORT=8080
   GIN_MODE=release
   MONGODB_URI=mongodb://localhost:27017
   DATABASE_NAME=vibanda_village
   JWT_SECRET=your-super-secret-jwt-key-here
   JWT_EXPIRATION_HOURS=24
   ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
   MAX_FILE_SIZE=10MB
   UPLOAD_PATH=uploads/
   ```

4. **Start MongoDB**
   Make sure MongoDB is running on your system.

5. **Run the application**
   ```bash
   go run cmd/main.go
   ```

The server will start on `http://localhost:8080`.

## API Documentation

Once the server is running, visit `http://localhost:8080/swagger/index.html` for interactive API documentation.

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login user
- `GET /api/v1/auth/profile` - Get user profile

### Users (Admin only)
- `GET /api/v1/users` - Get all users
- `GET /api/v1/users/{id}` - Get user by ID
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user

### Products (Admin & Manager)
- `GET /api/v1/products` - Get all products
- `GET /api/v1/products/{id}` - Get product by ID
- `POST /api/v1/products` - Create product
- `PUT /api/v1/products/{id}` - Update product
- `DELETE /api/v1/products/{id}` - Delete product

### Orders (Admin & Manager)
- `GET /api/v1/orders` - Get all orders
- `GET /api/v1/orders/{id}` - Get order by ID
- `POST /api/v1/orders` - Create order
- `PUT /api/v1/orders/{id}` - Update order
- `DELETE /api/v1/orders/{id}` - Delete order

### Events (Admin & Manager)
- `GET /api/v1/events` - Get all events
- `GET /api/v1/events/{id}` - Get event by ID
- `POST /api/v1/events` - Create event
- `PUT /api/v1/events/{id}` - Update event
- `DELETE /api/v1/events/{id}` - Delete event

### Reservations (Admin & Manager)
- `GET /api/v1/reservations` - Get all reservations
- `GET /api/v1/reservations/{id}` - Get reservation by ID
- `POST /api/v1/reservations` - Create reservation
- `PUT /api/v1/reservations/{id}` - Update reservation
- `DELETE /api/v1/reservations/{id}` - Delete reservation

## User Roles

- **Admin**: Full access to all features
- **Manager**: Access to products, orders, events, and reservations
- **Staff**: Limited access (can be extended based on requirements)

## Development

### Generate Swagger Documentation
```bash
swag init -g cmd/main.go
```

### Run Tests
```bash
go test ./...
```

### Build for Production
```bash
go build -o bin/vibanda-backend cmd/main.go
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `GIN_MODE` | Gin mode (debug/release) | `debug` |
| `MONGODB_URI` | MongoDB connection URI | `mongodb://localhost:27017` |
| `DATABASE_NAME` | Database name | `vibanda_village` |
| `JWT_SECRET` | JWT signing secret | `your-super-secret-jwt-key-here` |
| `JWT_EXPIRATION_HOURS` | JWT token expiration | `24` |
| `ALLOWED_ORIGINS` | CORS allowed origins | `http://localhost:3000,http://localhost:5173` |
| `MAX_FILE_SIZE` | Maximum file upload size | `10MB` |
| `UPLOAD_PATH` | File upload directory | `uploads/` |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.
