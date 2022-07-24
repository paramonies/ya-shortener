.PHONY: run
run:
	go run cmd/shortener/main.go

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

.PHONY: save_mem_profile
save_mem_profile:
	curl -sK -v http://localhost:8080/debug/pprof/heap > profiles/base.pprof

.PHONY: display_mem_profile
display_mem_profile:
	go tool pprof -http=":9090" -seconds=30 profiles/base.pprof

.PHONY: show_diff_mem
show_diff_mem:
	go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof

#go tool pprof -http=":9090" -seconds=30 http://localhost:8080/debug/pprof/heap
#make request POST http://localhost:8080/
#curl -sK -v http://localhost:8080/debug/pprof/heap > profiles/base.pprof



