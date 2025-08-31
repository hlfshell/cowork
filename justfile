# Cowork CLI Development Commands
# 
# This justfile provides common development tasks for the cowork CLI project.
# Run `just --list` to see all available commands.
#
# Examples:
#   just build          # Build the main cowork binary
#   just test           # Run all tests
#   just clean          # Clean build artifacts
#   just fmt            # Format all Go code
#   just lint           # Run linter
#   just coverage       # Run tests with coverage

# Default target when no arguments provided
default:
    @just --list

# Build the main cowork CLI binary
# Usage: just build
build:
    @echo "🔨 Building cowork CLI..."
    go build -o cw ./cmd/cowork
    @echo "✅ Built cowork cw app"

# Run all tests
# Usage: just test
test:
    @echo "🧪 Running all tests..."
    go test ./... -v

# Run tests with coverage
# Usage: just coverage
coverage:
    @echo "📊 Running tests with coverage..."
    go test ./... -v -coverprofile=coverage.out
    @echo "📈 Coverage report generated: coverage.out"
    @echo "🌐 Opening coverage report in browser..."
    go tool cover -html=coverage.out


# Clean build artifacts
# Usage: just clean
clean:
    @echo "🧹 Cleaning build artifacts..."
    rm -f cw cowork coverage.out
    go clean -cache
    @echo "✅ Cleaned"

# Clean and rebuild
# Usage: just rebuild
rebuild: clean build
    @echo "✅ Rebuilt"

# Install dependencies
# Usage: just deps
deps:
    @echo "📦 Installing dependencies..."
    go mod download
    go mod tidy
    @echo "✅ Dependencies installed"

# Update dependencies
# Usage: just update-deps
update-deps:
    @echo "🔄 Updating dependencies..."
    go get -u ./...
    go mod tidy
    @echo "✅ Dependencies updated"


# Check for security vulnerabilities
# Usage: just security
security:
    @echo "🔒 Checking for security vulnerabilities..."
    go list -json -deps ./... | nancy sleuth

# Generate documentation
# Usage: just docs
docs:
    @echo "📚 Generating documentation..."
    godoc -http=:6060 &
    @echo "📖 Documentation available at http://localhost:6060"
    @echo "Press Ctrl+C to stop"