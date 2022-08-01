.PHONY: run
run:
	go run cmd/shortener/main.go

.PHONY: buildlint
buildlint:
	go build -o ./cmd/staticlint/main ./cmd/staticlint/main.go

.PHONY: startlint
startlint:
	./cmd/staticlint/main ./...

.PHONY: fmt
fmt:
	goimports -local "github.com/paramonies" -w .

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


#curl -sK -v  http://localhost:8080/debug/pprof/heap > ./profiles/result.pprof
#go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof




