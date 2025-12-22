[![Build Status](https://drone.ykonkov.com/api/badges/ykonkov/survey-bot/status.svg)](https://drone.ykonkov.com/ykonkov/survey-bot)

# Survey Bot 

A sophisticated Telegram bot designed for conducting psychological surveys and assessments with automated result processing and analysis. 

**Telegram link:** [survey-bot](https://t.me/survey_1_bot)

## Description

Survey Bot is a comprehensive Telegram bot platform that enables users to participate in various psychological surveys and assessments. The bot provides an interactive experience for completing surveys, automatic result calculation, and detailed analytics for administrators.

## Features

- **Interactive Survey Experience**: Users can complete surveys through intuitive Telegram interface
- **Multiple Survey Types**: Support for various psychological assessments including burnout tests, personality assessments, and more
- **Automated Result Processing**: Real-time calculation and analysis of survey responses
- **Administrative Dashboard**: Web-based admin interface for managing surveys and viewing results
- **Metrics and Monitoring**: Built-in Prometheus metrics for monitoring bot performance
- **Database Management**: PostgreSQL backend with automated backup and restore capabilities
- **CLI Tools**: Command-line interface for survey management and data export
- **Development Environment**: Docker-based development setup with VSCode devcontainer support 

## Installation

### Prerequisites

- Go 1.22 or higher
- PostgreSQL 12+
- Docker and Docker Compose (for development)
- Telegram Bot Token (obtain from [@BotFather](https://t.me/botfather))

### Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/Goalt/survey-bot.git
   cd survey-bot
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Configure environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Set up the database:**
   ```bash
   # Create PostgreSQL database
   createdb survey_bot
   
   # Run migrations (automatically handled when DB_MIGRATIONS_UP=true)
   ```

5. **Build the application:**
   ```bash
   make build
   ```

### Configuration

The application uses environment variables for configuration:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `TOKEN` | Telegram Bot Token | - | ✅ |
| `DB_HOST` | PostgreSQL host | - | ✅ |
| `DB_PORT` | PostgreSQL port | 5432 | ❌ |
| `DB_NAME` | Database name | - | ✅ |
| `DB_USER` | Database user | - | ✅ |
| `DB_PWD` | Database password | - | ✅ |
| `DB_SSL_MODE` | SSL mode for database | disable | ❌ |
| `DB_MIGRATIONS_UP` | Run migrations on startup | true | ❌ |
| `ADMIN_USER_ID` | Comma-separated admin user IDs | -1 | ❌ |
| `LEVEL` | Log level (debug, info, warn, error) | info | ❌ |
| `ENV` | Environment (dev, prod) | dev | ❌ |
| `RELEASE_VERSION` | Application version | - | ✅ |
| `POLL_DURATION` | Bot polling interval | 1m | ❌ |
| `SENTRY_DSN` | Sentry DSN for error tracking | - | ❌ |
| `SENTRY_TIMEOUT` | Sentry timeout | 5s | ❌ |
| `METRICS_PORT` | Prometheus metrics port | 7777 | ❌ |
| `API_PORT` | HTTP API port | 8080 | ❌ |
| `ALLOWED_ORIGINS` | CORS allowed origins | - | ❌ |


## Development

### Development Environment Setup

The project uses VSCode devcontainer for a consistent development environment. All necessary configurations are included in the repository.

#### Using VSCode Devcontainer (Recommended)

1. Install [Docker](https://www.docker.com/get-started) and [VSCode](https://code.visualstudio.com/)
2. Install the [Remote - Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension
3. Open the project in VSCode
4. When prompted, click "Reopen in Container" or use `Ctrl+Shift+P` → "Remote-Containers: Reopen in Container"

#### Manual Development Setup

1. **Install Go 1.22+**
2. **Install PostgreSQL** and create a database
3. **Set environment variables** (see Configuration section)
4. **Run the application:**
   ```bash
   # Start the bot
   go run cmd/bot/main.go
   
   # Or use the CLI
   go run cmd/cli/main.go --help
   ```

### Available Make Commands

| Command | Description |
|---------|-------------|
| `make test` | Run all tests with coverage |
| `make lint` | Run golangci-lint |
| `make build` | Build CLI binary for arm64 and darwin |
| `make mocks` | Generate mocks for testing |

### Running Tests

```bash
# Run all tests
make test

# Run tests for specific package
go test ./internal/service -v

# Run tests with coverage
go test ./... -cover
```

### Code Quality

The project uses `golangci-lint` for code quality checks:

```bash
make lint
```

## CLI Usage

The survey-bot includes a powerful CLI for administrative tasks:

### Survey Management

#### Create a Survey
```bash
./bin/cli survey-create /path/to/survey.json
```

Creates a new survey from a JSON file and prints the survey GUID and ID.

#### Update a Survey
```bash
./bin/cli survey-update <survey_guid> /path/to/survey.json
```

Updates an existing survey by GUID. Updates "name", "description", "questions" and "calculations_type" fields.

#### Export Results
```bash
./bin/cli survey-get-results > results.csv
```

Exports all survey results in CSV format to stdout.

### Survey JSON Format

Survey files should follow this structure:

```json
{
  "name": "Survey Name",
  "description": "Survey description",
  "calculations_type": "test_1",
  "questions": [
    {
      "text": "Question text",
      "answer_type": "select",
      "possible_answers": [1, 2, 3, 4, 5],
      "answers_text": ["Never", "Rarely", "Sometimes", "Often", "Always"]
    }
  ]
}
```

## API Endpoints

The bot includes a REST API for administrative access:

### Base URL
```
http://localhost:8080
```

### Endpoints

- **GET /metrics** - Prometheus metrics (port 7777)
- **API endpoints** - Protected by Telegram authentication

### Metrics

Prometheus metrics are available at `:7777/metrics` for monitoring:
- Bot performance metrics
- Database connection metrics
- Survey completion rates
- User activity metrics

## Database

### Architecture

The application uses PostgreSQL with the following main tables:
- `users` - Telegram user information
- `surveys` - Survey definitions and metadata
- `survey_states` - User progress and responses

### Backup and Restore

#### Automated Backups
Database dumps are automatically created and stored in S3 (Yandex Cloud).

#### Manual Backup
```bash
pg_dump -h localhost -U postgres -p 5432 survey_bot > backup.sql
```

#### Restore from Backup
```bash
psql -h localhost -U postgres -p 54751 -f restore.sql
```

### Migrations

Database migrations are handled automatically when `DB_MIGRATIONS_UP=true` (default). Migration files are embedded in the application binary.

## Survey Management

### Adding New Survey Types

To add a new survey type to the system:

1. **Create Survey JSON File**
   ```bash
   # Add your survey definition to the surveytests folder
   cp template.json surveytests/new_survey.json
   # Edit the file with your survey content
   ```

2. **Add Calculation Logic**
   ```bash
   # Add new calculation_type to internal/resultsprocessor.go
   # Implement the scoring algorithm for your survey type
   ```

3. **Build and Deploy**
   ```bash
   make build
   ./bin/cli survey-create /workspace/surveytests/new_survey.json
   ```

4. **Add Bot Handlers**
   ```bash
   # Add button handler in internal/listener/handler.go
   # Add button selector in internal/listener/listener.go
   ```

### Survey Types Available

The system includes several pre-built psychological assessments:
- **Burnout Assessment** (`test_1`) - Professional burnout evaluation
- **Personality Tests** - Various personality assessment tools
- **Custom Surveys** - Configurable survey templates

## Deployment

### Production Deployment Flow

1. **Create Merge Request**
   - Submit code changes via MR/PR
   - Ensure all tests pass

2. **Automated Checks**
   - Linting (`make lint`)
   - Testing (`make test`)
   - Build verification

3. **Merge to Master**
   - Review and merge approved changes

4. **Release Process**
   - Create release tag
   - Automated deployment to production
   - Monitor deployment metrics

### Docker Deployment

```bash
# Build Docker image
docker build -t survey-bot .

# Run with docker-compose
docker-compose up -d
```

### Environment-Specific Configuration

- **Development**: Use `.env` file with devcontainer
- **Production**: Set environment variables via container orchestration
- **Testing**: Use test database with migrations

## Contributing

### Getting Started

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Set up development environment (see Development section)
4. Make your changes following the coding standards

### Code Standards

- Follow Go best practices and conventions
- Write tests for new functionality
- Ensure `make lint` passes without errors
- Add documentation for new features
- Use conventional commit messages

### Pull Request Process

1. Update documentation if needed
2. Add tests for new functionality
3. Ensure all tests pass (`make test`)
4. Run linter (`make lint`)
5. Submit pull request with clear description

### Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Include relevant logs and configuration details
- Provide steps to reproduce for bugs

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contact

- **Telegram Bot**: [@survey_1_bot](https://t.me/survey_1_bot)
- **Repository**: [github.com/Goalt/survey-bot](https://github.com/Goalt/survey-bot)
- **Issues**: [GitHub Issues](https://github.com/Goalt/survey-bot/issues)

## Acknowledgments

- Built with [Go](https://golang.org/)
- Telegram integration via [telebot](https://github.com/tucnak/telebot)
- Database management with [PostgreSQL](https://www.postgresql.org/)
- Monitoring with [Prometheus](https://prometheus.io/)
- Development environment powered by [Docker](https://www.docker.com/)