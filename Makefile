.PHONY: run
run:
	go run cmd/shortener/main.go

.PHONY: env_up
env_up:
	docker-compose up -d
	docker-compose ps
	./build/wait.sh
	sql-migrate up -env=local
	sql-migrate status -env=local

.PHONY: env_down
env_down:
	docker-compose down -v --rmi local --remove-orphans