.PHONY: migrate-up migrate-create deploy

include .env

prod?=
user?=bysoft

docker_compose_args=$(if $(prod), -f docker-compose.prod.yml, -f docker-compose.yml)

migrate-up:
	docker compose $(docker_compose_args) run migrate -path /migrations -database "${POSTGRES_URL}" -verbose up

migrate-down:
	docker compose $(docker_compose_args) run migrate -path /migrations -database "${POSTGRES_URL}" -verbose down

migrate-drop:
	docker compose $(docker_compose_args) run migrate -path /migrations -database "${POSTGRES_URL}" -verbose drop

migrate-create:	
	docker compose $(docker_compose_args) run migrate create -dir /migrations -ext sql $(name)	

docker-stop:
	docker compose $(docker_compose_args) stop

docker-build:
	docker compose $(docker_compose_args) build bysoft-users --build-arg user=$(user) --delete-orphans

docker-up:
	docker compose $(docker_compose_args) up -d

git-pull:
	git pull origin main 

deploy: docker-stop git-pull docker-build docker-up
