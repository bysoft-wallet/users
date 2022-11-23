.PHONY: migrate-up migrate-create

include .env

migrate-up:
	docker-compose run migrate -path /migrations -database "${POSTGRES_URL}" -verbose up

migrate-down:
	docker-compose run migrate -path /migrations -database "${POSTGRES_URL}" -verbose down

migrate-drop:
	docker-compose run migrate -path /migrations -database "${POSTGRES_URL}" -verbose drop

migrate-create:	
	docker-compose run migrate create -dir /migrations -ext sql $(name)	

deploy:
	docker compose -f docker-compose.prod.yml stop bysoft-users && docker compose -f docker-compose.prod.yml build bysoft-users --build-arg user=$(user) && docker compose -f docker-compose.prod.yml up -d