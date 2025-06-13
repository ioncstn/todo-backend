# Default service name(s) (can be left empty to act on all)
srv =

# Start containers
start:
	docker compose up -d $(srv)

# Stop containers
stop:
	docker compose down $(srv)

# Restart containers
restart: stop start

# Follow container logs
logs:
	docker compose logs -f $(srv)

# Test database management
test-db-start:
	docker compose -f docker-compose.test.yaml up -d

test-db-stop:
	docker compose -f docker-compose.test.yaml down

# Run tests
test: test-db-start
	go test -v ./... -count=1
	$(MAKE) test-db-stop

# Run tests with coverage
test-coverage: test-db-start
	go test -v ./... -count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out
	$(MAKE) test-db-stop