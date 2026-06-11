# Makefile for AgentMesh

.PHONY: install dev docker-up docker-down test build clean help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOVENDOR=$(GOCMD) vendor

# Docker parameters
COMPOSE=docker-compose

# Service paths
SERVICES:=api-gateway user-svc plugin-svc

# Default target
help:
	@echo "AgentMesh Development Commands"
	@echo "================================"
	@echo "install      - Install Go dependencies for all services"
	@echo "dev          - Start all services in development mode"
	@echo "docker-up    - Start all services with docker-compose"
	@echo "docker-down  - Stop all docker-compose services"
	@echo "test         - Run tests for all services"
	@echo "test-coverage - Run Go and frontend tests with coverage"
	@echo "build        - Build all service binaries"
	@echo "clean        - Clean build artifacts"

install:
	@for svc in $(SERVICES); do \
		echo "Installing dependencies for $$svc..."; \
		cd services/$$svc && $(GOMOD) download && cd ../..; \
	done

dev:
	@echo "Starting development environment..."
	@for svc in $(SERVICES); do \
		echo "Starting $$svc in dev mode..."; \
		cd services/$$svc && $(GOCMD) run . & \
	done

docker-up:
	@echo "Starting Docker services..."
	$(COMPOSE) up -d
	@echo "Services started. API Gateway: http://localhost:8080"

docker-down:
	@echo "Stopping Docker services..."
	$(COMPOSE) down

test:
	@for svc in $(SERVICES); do \
		echo "Testing $$svc..."; \
		(cd services/$$svc && $(GOTEST) ./...); \
	done

test-coverage:
	@for svc in $(SERVICES); do \
		echo "Coverage for $$svc..."; \
		(cd services/$$svc && $(GOTEST) -cover ./...); \
	done
	@echo "Frontend coverage:"; \
	cd web && pnpm test:coverage

build:
	@echo "Building all services..."
	@mkdir -p bin
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		cd services/$$svc && $(GOBUILD) -o ../../bin/$$svc . && cd ../..; \
	done

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@for svc in $(SERVICES); do \
		rm -f services/$$svc/*.exe; \
	done
