build-server:
	go build -o ./bin/server ./cmd/server

run-server-db:
	./bin/server -d "postgres://user:password@localhost:5432/db?sslmode=disable" 

run-server-file:
	./bin/server -f "./metrics.json"

run-server-memory:
	./bin/server 

build-agent:
	go build -o ./bin/agent ./cmd/agent

run-agent:
	./bin/agent -k "key"

mockgen:	
	mockgen -source=$(file) \
		-destination=$(dir $(file))$(notdir $(basename $(file)))_mock.go \
		-package=$(shell basename $(dir $(file)))

test:
	go test ./internal/... -cover	

migrate:
	goose -dir ./migrations postgres "postgres://user:password@localhost:5432/db?sslmode=disable" up

docker-run:
	docker run --name metrics-postgres \
		-e POSTGRES_USER=user \
		-e POSTGRES_PASSWORD=password \
		-e POSTGRES_DB=db \
		-p 5432:5432 \
		-d postgres:15
