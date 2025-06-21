# Library Management System

A microservices-based library management system with gRPC and REST API interfaces, with OpenAPI/Swagger documentation.

## Microservices Architecture

The system is composed of three loosely coupled microservices:

1. **User Service**: Handles user registration, authentication, and profile management
2. **Book Service**: Manages book inventory, including adding, updating, and retrieving books
3. **Transaction Service**: Manages book borrowing and returning operations

Each service has its own database and can be deployed and scaled independently.

## Protocol Buffers and Code Generation

This project uses [buf.build](https://buf.build) for Protocol Buffer management and code generation.

### Buf Configuration Files

- `buf.yaml`: Main configuration file for the buf module
- `buf.gen.yaml`: Configuration for code generation
- `buf.work.yaml`: Workspace configuration defining the proto file locations

### Available Commands

- `make generate`: Generate code from proto files using buf
- `make swagger`: Generate OpenAPI/Swagger documentation
- `make lint`: Lint proto files using buf
- `make breaking`: Check for breaking changes in proto files
- `make clean`: Remove generated files

This will create a `swagger.yaml` file in the `swagger` directory that describes the API endpoints and data models.

## Getting Started

1. Install buf: https://buf.build/docs/installation
2. Run `make generate` to generate code from proto files
3. Set up PostgreSQL databases for each service and update the `.env` file with your database credentials
4. Start each service:
   - User Service: `go run cmd/user-service/main.go`
   - Book Service: `go run cmd/book-service/main.go`
   - Transaction Service: `go run cmd/transaction-service/main.go`

## User Service

The user service handles user management with role-based access control:

### Features

- User registration and authentication
- JWT-based authentication
- Role-based access control (admin and operation users)
- User profile management

### API Endpoints

#### gRPC Endpoints

- `Register`: Register a new user
- `Login`: Authenticate a user and get a JWT token
- `Get`: Get user details
- `Update`: Update user information

#### REST Endpoints (via gRPC Gateway)

- `POST /api/users/register`: Register a new user
- `POST /api/users/login`: Authenticate a user and get a JWT token
- `GET /api/users/{id}`: Get user details
- `PATCH /api/users/{id}`: Update user information

## Book Service

The book service manages the book inventory:

### Features

- Book creation and management
- Book search and retrieval
- Book recommendations

### API Endpoints

#### gRPC Endpoints

- `Create`: Create a new book
- `ListBooks`: List all books
- `GetBook`: Get book details
- `Recommend`: Get book recommendations

#### REST Endpoints (via gRPC Gateway)

- `POST /api/books`: Create a new book
- `GET /api/books`: List all books
- `GET /api/books/{id}`: Get book details
- `GET /api/books/recommend`: Get book recommendations

## Transaction Service

The transaction service manages book borrowing and returning:

### Features

- Book borrowing
- Book returning
- Transaction history

### API Endpoints

#### gRPC Endpoints

- `Borrow`: Borrow a book
- `Return`: Return a book
- `History`: Get transaction history for a user

#### REST Endpoints (via gRPC Gateway)

- `POST /api/transactions/borrow`: Borrow a book
- `POST /api/transactions/return`: Return a book
- `GET /api/transactions/user/{user_id}`: Get transaction history for a user

## Authentication

The API uses JWT tokens for authentication. To access protected endpoints:

1. Login to get a token
2. Include the token in the `Authorization` header of subsequent requests:
   - For gRPC: Include metadata with key `authorization` and value `Bearer <token>`
   - For REST: Include header `Authorization: Bearer <token>`

## Role-Based Access Control

- Operation users can only access and modify their own data
- Admin users can access and modify any user's data
