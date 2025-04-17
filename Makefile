##############################################################
# Makefile for Merhongo
##############################################################

.PHONY: test cover cover-html lint fmt check docker-up docker-down docker-logs all clean update-readme

GO = go
GOFMT = gofmt
GOLINT = golangci-lint

# Test configuration
TEST_ARGS = -v
COVER_ARGS = -coverpkg=./... -coverprofile=coverage.out

# Package list
PACKAGES = ./...

# Default target
all: test

# Run all tests
test:
	@echo "Running tests..."
	$(GO) test $(PACKAGES) $(TEST_ARGS)

# Run tests with coverage
cover:
	@echo "Running tests with coverage..."
	$(GO) test $(PACKAGES) $(TEST_ARGS) $(COVER_ARGS)
	$(GO) tool cover -func=coverage.out

# Generate HTML coverage report
cover-html: cover
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Opening coverage report in browser..."
	open coverage.html || xdg-open coverage.html || sensible-browser coverage.html || echo "Could not open browser, coverage report is at coverage.html"

# Update README with coverage information
update-readme: cover
	@echo "Updating README with coverage information..."
	@chmod +x scripts/update-readme-coverage.sh
	@./scripts/update-readme-coverage.sh

# Generate coverage report and update README
cover-update: cover update-readme

# Run linting
lint:
	@if command -v $(GOLINT) > /dev/null; then \
		echo "Running linter..."; \
		$(GOLINT) run; \
	else \
		echo "golangci-lint not found, please install it: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w -s .

# Check code formatting (without modifying files)
check-fmt:
	@echo "Checking code formatting..."
	@test -z "$$($(GOFMT) -l .)" || { echo "Some files need formatting:"; $(GOFMT) -l .; exit 1; }

# Docker commands for MongoDB
mongo-start:
	@echo "Starting MongoDB container..."
	docker run --name mongodb -d -p 27017:27017 mongo:latest || docker start mongodb

mongo-stop:
	@echo "Stopping MongoDB container..."
	docker stop mongodb

mongo-restart:
	@echo "Restarting MongoDB container..."
	docker restart mongodb

mongo-logs:
	@echo "Showing MongoDB logs..."
	docker logs mongodb

# Docker Compose commands
docker-up:
	@echo "Starting all services with Docker Compose..."
	docker-compose up -d

docker-down:
	@echo "Stopping all services with Docker Compose..."
	docker-compose down

docker-logs:
	@echo "Showing logs from all services..."
	docker-compose logs

# Clean up
clean:
	@echo "Cleaning up..."
	rm -f coverage.out coverage.html