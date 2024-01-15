# Tigerhall Kittens - Backend

## Project Overview

This project aims to create a small web app for tracking the population of tigers in the wild. The backend is built using Golang and follows best practices for software design, unit testing, and web technologies like HTTP and REST/GraphQL.

## Project Name

Tigerhall Kittens

## Technology

Golang

## Delivered Solution

The solution is hosted on [GitHub](https://github.com/chegde20121/Tigerhall-Kittens).

## Supported APIs

| Endpoint                  | Description                                              |
|---------------------------|----------------------------------------------------------|
| `POST /api/v1/register`   | Create a new user with attributes: username, password, email. |
| `POST /api/v1/login`      | Log in using authentication credentials.                 |
| `POST /api/v1/createTigers`  | Create a new tiger with attributes: Name, Date of birth, Last seen timestamp, Last seen coordinates (Lat/Lon). |
| `GET /api/v1/listTigers`     | List all tigers, sorted by the last time they were seen.  |
| `POST /api/v1/createSights`     | Create a new sighting of a tiger with attributes: Lat/Lon, Timestamp. Supports image upload (resized to 250x200). |
| `GET /api/v1/tigers/:id/listSightings` | List all sightings of a specific tiger, sorted by date (Latest first). |
| `GET /api/v1/logout` | List all sightings of a specific tiger, sorted by date (Latest first). |


## Project Structure
```
tigerhall-kittens/
|-- docs/                      # Swagger docs
|-- internal/app/
|   |
|   |-- database/
|   |   |-- migrations/        # Database migration files
|   |   |-- repository/        # Database repository
|   |   |-- database_handler.go        # Database initialization
|   |
|   |-- handlers/              # Request handlers
|   |
|   |-- models/                # Data Models
|   |
|   |-- service/               # Business logic handlers
|-- pkg/                       # Reusable packages
|   |-- config/                # Environment config reader
|   |
|   |-- messaging/             # MessageQueue handlers
|
|-- scripts/
|   |-- deploy.sh             # Database migration script
|
|-- tests/
|   |-- unit_tests/             # Unit tests
|   |-- e2e/                    # End-to-end tests
|
|-- main.go                     # Starting point of the application
|-- app.log                     # log file
|-- .env                        # env file
|-- go.mod
|-- go.sum
|-- README.md
```
## How to Run

### Environment configurations
Update .env files before running the application.

#### Environment Variables

| Variable                 | Description                                             |
|--------------------------|---------------------------------------------------------|
| POSTGRES_USER            | PostgreSQL username                                     |
| POSTGRES_PASSWORD        | PostgreSQL password                                     |
| POSTGRES_DB              | PostgreSQL database name                                |
| POSTGRES_HOST            | PostgreSQL server host                                  |
| POSTGRES_PORT            | PostgreSQL server port                                  |
| POSTGRES_SSL             | PostgreSQL SSL mode (e.g., 'disable')                   |
| POSTGRES_TIMEZONE        | PostgreSQL server timezone                              |
| POSTGRES_SCHEMA          | PostgreSQL schema name                                  |
| MIGRATION_FILES          | Path to general migration files                         |
| MIGRATION_FILES_DB       | Path to database-specific migration files               |
| MIGRATION_UP             | Perform database migrations on startup (true/false)     |
| PAGE_SIZE                | Default page size for paginated results                 |
| MAX_EMAIL_GOROUTINES     | Maximum number of goroutines for email sending          |
| SENDER_EMAIL             | Sender email address for outgoing emails                |
| SENDER_EMAIL_PASSWORD    | Password for the sender email account                   |
| SMTP_HOST                | SMTP server host for email sending                      |
| SMTP_PORT                | SMTP server port for email sending                      |

*Note: Replace the placeholder values with your actual configuration. Ensure sensitive information like passwords is kept secure and not disclosed in the README.*


### Using deploy.sh Script
1. Navigate to the project root directory.

2. Run the deploy script with the desired options:

   ```bash
   ./scripts/deploy.sh [options]
   ```
3. Available options are as below
| Option      | Description                                      |
|-------------|--------------------------------------------------|
| -notest     | Skip running tests.                              |
| -build      | Skip tests and only build the project.           |
| -run        | Skip tests and only run the server.              |
| -shutdown   | Gracefully shutdown the running server.          |


## Server Configuration

| Server Name | Port   | Swagger URL                                      |
|-------------|--------|--------------------------------------------------|
| localhost | 8888   | http://localhost:8888/swagger/index.html           |


## Pending Tasks
- [ ] Start server in https mode.
- [ ] Refactor code for better modularity.
- [ ] Write comprehensive unit tests.
- [ ] Update documentation for API endpoints.
- [ ] Enhance error handling and logging.
