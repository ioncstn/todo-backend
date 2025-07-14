# Todo App

A modern, containerized Todo application built with Go, featuring a RESTful API backend and MySQL database. The application uses Echo framework for the HTTP server, GORM for database operations, and Zap for logging.

## Features & Tech Stack

- RESTful API backend built with Go and Echo framework
- MySQL database with GORM ORM
- Structured logging with Zap
- Hot-reload development environment with Air
- Docker containerization
- Makefile for easy development workflow

## Prerequisites

- Docker and Docker Compose
- Go 1.24.3 or later
- Make (optional, for using Makefile commands)

## Getting Started

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd todo-app
   ```

2. Start the application using Docker Compose:
   ```bash
   make start
   ```
   This will start both the application and MySQL database containers.

3. The application will be available at:
   - API Server: http://localhost:8765
   - MySQL Database: localhost:3306

## API Documentation

Detailed API documentation is available in each internal domain package.
It includes available endpoints, request / response formats, data models, error handling.

ToDo Items - [read here](internal/todos/API.md)

## Development

The project includes several Makefile commands to help with development:

- `make start` - Start the containers
- `make stop` - Stop the containers
- `make restart` - Restart the containers
- `make logs` - View container logs

To follow logs for a specific service:
```bash
make logs srv=todo-app
```

### Debugging

The local environment container also installs delve, so we can make use of debugging and live reloading while developing.

In order to make use of delve, you just need to connect to it through your IDE.

You have to add a new Go Remote debug configuration, that uses port `2345`.

There is no need to rebuild the app container after every code change, because air will rebuild the app on file modifications and run it using delve.

## Email Verification

This application includes email verification functionality for user registration:

### Features
- When a user creates an account via `POST /user`, they receive a verification email
- The account is not active until the email is verified
- Verification tokens expire after 24 hours
- Users can verify their email by clicking the link in the email or visiting `/verify-email?token=<token>`

### Mailpit Integration
The application uses Mailpit for email testing in development:
- Mailpit SMTP server runs on port 1025
- Mailpit web interface is available at http://localhost:8025
- All emails are captured and can be viewed in the web interface
- Mailpit offers better performance and a more modern UI compared to Mailhog

### Environment Variables
```
SMTP_HOST=mailpit
SMTP_PORT=1025
SMTP_USER=""
SMTP_PASSWORD=""
APP_URL=http://localhost:8765
```

### Running the Application
```bash
# Start all services including Mailpit
docker-compose up

# The application will be available at:
# - API: http://localhost:8765
# - Mailpit UI: http://localhost:8025
# - Swagger: http://localhost:8765/swagger/
```

### Testing Email Verification
1. Create a user via POST `/user`
2. Check Mailpit UI at http://localhost:8025 for the verification email
3. Click the verification link or copy the token and visit `/verify-email?token=<token>`
4. The user account will be verified and ready to use