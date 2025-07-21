# Marketplace Backend

This is a REST API for a minimal marketplace, built with Go, following clean architecture principles.

## How to Run (Local Developer)

1. **Prerequisites**:
   - Docker and Docker Compose installed
   - Go 1.24.0


2. **Setup**:
   Copy the example environment file and fill in the required secrets:
   ```bash
   cp .env.example .env
   # Edit .env and set JWT_SECRET and other required values
   ```

3. **Start Services**:
   ```bash
   make compose-up
   ```
   This will start:
   - API server on port 8080
   - PostgreSQL database
   - Migration service
4. **API Documentation**:
   - Swagger UI: http://localhost:8080/swagger/index.html
   - API Base URL: http://localhost:8080/v1

5. **Authentication**:
   - GET /ads optional authentication
   - POST /ads require authentication
   - Include the JWT token in the `Authorization` header:
     ```
     Bearer YOUR_JWT_TOKEN
     ```
   - Get a token by authenticating at `/v1/auth/login`

6. **Development**:
   - Run tests: `make test`
   - Generate mocks: `make generate`
   - Lint code: `make lint`
   - Generate Swagger docs: `make swagger`

7. **Stop Services**:
   ```bash
   make compose-down
   ```

## Conscious Design Decisions

### 1. Service Layer Validation
- Business logic and validation rules are implemented in the service layer
- This keeps handlers clean and focused on HTTP concerns

### 2. Clean Architecture with Separate Interfaces
- Each service defines its own repository interfaces
- Database implementation is decoupled from business logic
- Makes the code more testable and maintainable
- Allows for easier database migrations or changes

### 3. Denormalized User Data in Posts
- `user_login` is denormalized and stored with each post
- Improves read performance for common queries
- Reduces the need for joins when displaying post information
- Maintains data consistency through application logic

### 4. Error Handling
- Clear separation between different error types (not found, validation, auth, etc.)
- Consistent error responses across the API
- Proper error conversion between layers (repository → service → handler)

### 5. Containerization
- Self-contained development environment
- Consistent database setup across all developers
- Easy to onboard new team members
