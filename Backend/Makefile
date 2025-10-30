.PHONY: build run test clean docker-build docker-run python-services run-all run-test

# Go build settings
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=salmar-api
BINARY_UNIX=$(BINARY_NAME)_unix

# Python settings
PYTHON=python3

# Build Go API
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/api

# Run Go API
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/api
	./$(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/api

# Run Python services
python-services:
	cd python && $(PYTHON) run_services.py

# Run a specific Python service
python-service-%:
	cd python && $(PYTHON) run_services.py --service $*

# Install Go dependencies
deps:
	$(GOCMD) mod tidy

# Install Python dependencies
python-deps:
	pip install -r python/requirements.txt
	cd python && pip install -r requirements.txt

# Run the entire Salmar AI system using the runner script
run-all:
	python run_salmar.py

# Run the system with tests
run-test:
	python run_salmar.py --test

# Build Docker images
docker-build:
	docker-compose build

# Run with Docker Compose
docker-run:
	docker-compose up

# Stop Docker Compose services
docker-stop:
	docker-compose down

# Setup development environment
setup: deps python-deps
	cp -n .env.example .env || true

# Default target
all: setup build