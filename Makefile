.PHONY: build run clean install

# Build the application
build:
	go build -o sdtop ./cmd/main.go

# Run the application
run:
	go run ./cmd/main.go

# Clean build artifacts
clean:
	rm -f sdtop

# Install to system
install: build
	sudo mv sdtop /usr/local/bin/

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run all checks
check: fmt vet
	@echo "All checks passed!"
