#!/bin/bash
# Docker Development Environment Management Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
}

# Setup environment file
setup_env() {
    if [ ! -f "$SCRIPT_DIR/.env" ]; then
        log_info "Creating .env file from example..."
        cp "$SCRIPT_DIR/.env.example" "$SCRIPT_DIR/.env"
        log_warning "Please review and modify .env file with your settings"
    fi
}

# Build development image
build_dev() {
    log_info "Building development Docker image..."
    cd "$PROJECT_ROOT"
    docker build -f .docker/Dockerfile.dev -t repo-onboarding-copilot:dev .
    log_success "Development image built successfully"
}

# Build production image
build_prod() {
    log_info "Building production Docker image..."
    cd "$PROJECT_ROOT"
    docker build -f .docker/Dockerfile -t repo-onboarding-copilot:latest .
    log_success "Production image built successfully"
}

# Start development environment
start_dev() {
    log_info "Starting development environment..."
    cd "$SCRIPT_DIR"
    docker-compose --profile dev up -d
    log_success "Development environment started"
    log_info "Access the application at: http://localhost:8080"
    log_info "Debugger port available at: localhost:2345"
}

# Start full stack (with database and cache)
start_full() {
    log_info "Starting full development stack..."
    cd "$SCRIPT_DIR"
    docker-compose --profile dev --profile database --profile cache up -d
    log_success "Full development stack started"
    log_info "Services available:"
    log_info "  - Application: http://localhost:8080"
    log_info "  - PostgreSQL: localhost:5432"
    log_info "  - Redis: localhost:6379"
}

# Start with monitoring
start_monitoring() {
    log_info "Starting development environment with monitoring..."
    cd "$SCRIPT_DIR"
    docker-compose --profile dev --profile monitoring up -d
    log_success "Development environment with monitoring started"
    log_info "Services available:"
    log_info "  - Application: http://localhost:8080"
    log_info "  - Prometheus: http://localhost:9090"
    log_info "  - Grafana: http://localhost:3000"
}

# Stop all services
stop() {
    log_info "Stopping all Docker services..."
    cd "$SCRIPT_DIR"
    docker-compose down
    log_success "All services stopped"
}

# Clean up everything
clean() {
    log_warning "This will remove all containers, images, and volumes. Are you sure? (y/N)"
    read -r confirmation
    if [[ "$confirmation" =~ ^[Yy]$ ]]; then
        log_info "Cleaning up Docker resources..."
        cd "$SCRIPT_DIR"
        docker-compose down -v --rmi all --remove-orphans
        docker system prune -f
        log_success "Clean up completed"
    else
        log_info "Clean up cancelled"
    fi
}

# Show logs
logs() {
    local service=${1:-dev}
    log_info "Showing logs for service: $service"
    cd "$SCRIPT_DIR"
    docker-compose logs -f "$service"
}

# Run tests in container
test() {
    log_info "Running tests in Docker container..."
    cd "$PROJECT_ROOT"
    docker run --rm -v "$(pwd):/workspace" -w /workspace golang:1.23-alpine sh -c "
        apk add --no-cache make git &&
        make deps &&
        make test
    "
    log_success "Tests completed"
}

# Run linting in container
lint() {
    log_info "Running linting in Docker container..."
    cd "$PROJECT_ROOT"
    docker run --rm -v "$(pwd):/workspace" -w /workspace golang:1.23-alpine sh -c "
        apk add --no-cache make git &&
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest &&
        make lint
    "
    log_success "Linting completed"
}

# Run security scan in container
security() {
    log_info "Running security scan in Docker container..."
    cd "$PROJECT_ROOT"
    docker run --rm -v "$(pwd):/workspace" -w /workspace golang:1.23-alpine sh -c "
        apk add --no-cache make git &&
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest &&
        make security
    "
    log_success "Security scan completed"
}

# Shell into development container
shell() {
    log_info "Opening shell in development container..."
    cd "$SCRIPT_DIR"
    docker-compose exec dev bash || docker-compose run --rm dev bash
}

# Show container status
status() {
    log_info "Docker container status:"
    cd "$SCRIPT_DIR"
    docker-compose ps
}

# Show help
show_help() {
    cat << EOF
🐳 Docker Development Environment Manager

Usage: $0 [COMMAND]

Commands:
  build-dev       Build development Docker image
  build-prod      Build production Docker image
  start-dev       Start development environment (app only)
  start-full      Start full stack (app + database + cache)
  start-monitor   Start with monitoring (app + prometheus + grafana)
  stop            Stop all services
  clean           Remove all containers, images, and volumes
  logs [service]  Show logs for service (default: dev)
  test            Run tests in container
  lint            Run linting in container
  security        Run security scan in container
  shell           Open shell in development container
  status          Show container status
  help            Show this help message

Examples:
  $0 start-dev          # Start basic development environment
  $0 start-full         # Start with database and cache
  $0 logs dev           # Show development container logs
  $0 shell              # Open bash shell in dev container
  $0 clean              # Clean up all Docker resources

Environment:
  Copy .env.example to .env and customize settings before running.
EOF
}

# Main script logic
main() {
    check_docker
    setup_env

    case "${1:-help}" in
        build-dev)
            build_dev
            ;;
        build-prod)
            build_prod
            ;;
        start-dev)
            start_dev
            ;;
        start-full)
            start_full
            ;;
        start-monitor)
            start_monitoring
            ;;
        stop)
            stop
            ;;
        clean)
            clean
            ;;
        logs)
            logs "$2"
            ;;
        test)
            test
            ;;
        lint)
            lint
            ;;
        security)
            security
            ;;
        shell)
            shell
            ;;
        status)
            status
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"