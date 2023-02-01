export DOCKER_COMPOSE = docker-compose -f docker-compose.yaml -f docker-compose.override.yaml
DOCKER_GOLANG = docker run --rm -v ${PWD}:/app -w /app golang:1.18

start:
	$(DOCKER_COMPOSE) up --build -d
stop:
	$(DOCKER_COMPOSE) down
test:
	$(DOCKER_GOLANG) go test -v ./...
dev-env:
	sed -i 's/\(.*_HOST=\).*/\1localhost/' .env
	sed -i 's/\(DEBUG=\).*/\1true/' .env

# Global
.DEFAULT_GOAL := start
