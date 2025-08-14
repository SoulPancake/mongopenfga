.PHONY: help up down run logs status clean health start-openfga stop-openfga

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

up: ## Start MongoDB and OpenFGA services
	docker compose up -d
	@echo "Waiting for MongoDB to be healthy..."
	@timeout 60s bash -c 'until docker compose ps | grep mongo | grep -q "healthy"; do sleep 2; done'
	@echo "MongoDB is running!"
	@echo "Starting OpenFGA server locally..."
	@make start-openfga

start-openfga: ## Start OpenFGA server locally in background
	@pkill -f "openfga run" || true
	@cd ../.. && nohup ./dist/openfga run \
		--datastore-engine mongo \
		--datastore-uri mongodb://localhost:27017/openfga \
		--playground-enabled \
		--http-addr 0.0.0.0:8080 \
		--grpc-addr 0.0.0.0:8081 \
		--playground-port 3000 > /tmp/openfga.log 2>&1 &
	@echo "Waiting for OpenFGA to start..."
	@timeout 30s bash -c 'until curl -s http://localhost:8080/healthz > /dev/null 2>&1; do sleep 1; done'
	@echo "OpenFGA is running!"
	@echo "OpenFGA API: http://localhost:8080"
	@echo "Playground: http://localhost:3000/playground"
	@echo "MongoDB: mongodb://localhost:27017/openfga"

stop-openfga: ## Stop OpenFGA server
	@pkill -f "openfga run" || true
	@echo "OpenFGA stopped"

down: ## Stop and remove services
	@make stop-openfga
	docker compose down -v

run: ## Run the example client
	cd client && go run main.go

build-client: ## Build the client binary
	cd client && go build -o ../client-example main.go

logs: ## Show service logs
	docker compose logs -f

status: ## Show service status
	docker compose ps
	@echo "\nOpenFGA Process:"
	@pgrep -f "openfga run" && echo "OpenFGA is running" || echo "OpenFGA is not running"

health: ## Check service health
	@echo "=== OpenFGA Health ==="
	@curl -s http://localhost:8080/healthz | jq . || echo "OpenFGA not healthy"
	@echo "\n=== MongoDB Health ==="
	@docker exec mongo mongosh mongodb://localhost:27017/openfga --eval "db.adminCommand('ping')" --quiet || echo "MongoDB not healthy"

clean: down ## Clean up everything including volumes
	docker compose down -v --remove-orphans
	docker volume prune -f