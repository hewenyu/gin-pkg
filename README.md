# Gin-Pkg

A comprehensive template for building Go API services using the Gin framework with built-in JWT authentication, security validation, and user management.

## Features

- **JWT Authentication**: Complete implementation with access and refresh tokens
- **Request Security Validation**:
  - Timestamp validation to prevent replay attacks
  - Server-generated nonce system
  - Request signing
- **User Management**: Login, registration, token refresh, and user information
- **CLI Tool**: Quickly scaffold new projects based on this template
- **Modern Stack**:
  - Gin for API routing
  - Ent for database operations
  - Redis for cache and nonce management
  - Viper for configuration
- **Comprehensive Logging**: Structured logging with Zap
- **Role-Based Access Control**: Simple but effective RBAC implementation
- **Interface-Driven Development**: Clean, testable code through interfaces

## Project Structure

```
├── cmd/                   # Command-line applications
│   ├── server/            # Main API server
│   └── gin-pkg/           # CLI tool for creating new projects
├── pkg/                   # Reusable packages
│   ├── auth/              # Authentication components
│   │   ├── jwt/           # JWT token handling
│   │   └── security/      # Security validation
│   ├── middleware/        # Gin middleware implementations
│   ├── logger/            # Logging utilities
│   └── util/              # Helper functions and utilities
├── internal/              # Application-specific code
│   ├── app/               # Application initialization
│   ├── router/            # API routes definition
│   ├── service/           # Business logic services
│   ├── model/             # Data transfer objects
│   └── ent/               # Database entity models
└── config/                # Configuration files
```

## API Specifications

- Base URL: `/api/v1`
- Content Type: `application/json`
- Character Encoding: UTF-8
- Time Format: ISO 8601 (e.g., `2023-06-15T08:00:00Z`)

### Security Parameters

All API requests (except for specific endpoints) require:

1. **Timestamp** (`X-Timestamp` header or `timestamp` parameter)
2. **Nonce** (`X-Nonce` header or `nonce` parameter) - obtained from `/api/v1/auth/nonce`
3. **Signature** (`X-Sign` header or `sign` parameter) - HMAC-SHA256 of sorted request parameters

### API Endpoints

#### Authentication

- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Authenticate and get access tokens
- `POST /api/v1/auth/refresh` - Refresh access token
- `GET /api/v1/auth/nonce` - Get a new nonce for request signing

#### User Management

- `GET /api/v1/users` - List users (admin only)
- `GET /api/v1/users/:id` - Get user details
- `PUT /api/v1/users/:id` - Update user information
- `DELETE /api/v1/users/:id` - Delete a user

## Usage

### Security Flow

1. Get a nonce from the server using `/api/v1/auth/nonce`
2. For subsequent requests:
   - Include the nonce in your request
   - Add a timestamp (current time in ISO 8601)
   - Generate a signature by creating an HMAC-SHA256 of the sorted parameters
   - Include all three values in headers or query/body parameters
3. Send the request with appropriate Authorization header for protected endpoints

### Authentication Flow

1. Register a user via `/api/v1/auth/register`
2. Login via `/api/v1/auth/login` to receive access and refresh tokens
3. Use the access token in the `Authorization` header (format: `Bearer {token}`)
4. When the access token expires, use the refresh token to get new tokens

## Getting Started

```bash
# Clone the repository
git clone https://github.com/hewenyu/gin-pkg.git

# Navigate to the project directory
cd gin-pkg

# Install the CLI tool
go install ./cmd/gin-pkg

# Create a new project
gin-pkg new my-api-project

# Navigate to your new project
cd my-api-project

# Run the server (development mode)
go run cmd/server/main.go --debug
```

### Configuration

The default configuration is in `config/default.yaml`. You can customize:

- Server settings (port, timeouts)
- Database connection (PostgreSQL)
- Redis connection
- Authentication parameters (token secrets, expiration times)
- Security settings (timestamp validity window, nonce validity duration)

## Development

### Prerequisites

- Go 1.23+
- PostgreSQL
- Redis

### Testing

```bash
# Run all tests
go test ./...
```

## License

[MIT License](LICENSE) 