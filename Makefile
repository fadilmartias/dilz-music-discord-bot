APP_NAME = dilz-music-discord-bot
DOCKER_COMPOSE = docker compose

# Default env file
ENV_FILE = .env

.PHONY: dev prod logs stop clean rebuild

## Run in development mode (with air hot reload)
dev:
	$(DOCKER_COMPOSE) -f docker-compose.override.yml up --build

## Run in production mode (detached)
prod:
	$(DOCKER_COMPOSE) -f docker-compose.yml up --build -d

## Show logs
logs:
	$(DOCKER_COMPOSE) logs -f

## Stop all containers
stop:
	$(DOCKER_COMPOSE) down

## Remove containers, networks, images, and volumes
clean:
	$(DOCKER_COMPOSE) down -v --rmi all --remove-orphans

## Rebuild containers
rebuild:
	$(DOCKER_COMPOSE) build --no-cache
