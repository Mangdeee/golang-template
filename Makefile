app_container := template_server

help:
	@echo ''
	@echo 'Usage: make [TARGET] [EXTRA_ARGUMENTS]'
	@echo 'Targets:'
	@echo 'make dev: make dev for development work'
	@echo 'clean: clean for all clear docker images'
	@echo 'generate-mockery: generate mockery for testing'

dev:
	if [ ! -f .env ]; then cp .env.example .env; fi;
	docker-compose -f docker-compose.yml up --build

clean:
	docker-compose -f docker-compose.yml down -v

generate-mockery:
	docker exec -it $(app_container) /bin/sh -c "mockery --all"

test:
	go test -short ./...