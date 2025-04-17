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

## Usage

1. Get a nonce from the server
2. Construct your API request with the nonce, timestamp, and signature
3. Send the request with appropriate authorization header for protected endpoints

## Getting Started

```bash
# Clone the repository
git clone https://github.com/yourusername/gin-pkg.git

# Navigate to the project directory
cd gin-pkg

# Install the CLI tool
go install ./cmd/gin-pkg

# Create a new project
gin-pkg new my-api-project
```

## License

[MIT License](LICENSE) 