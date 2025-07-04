# Makefile for WordPress Go Next Backend

.PHONY: run test swag docker clean

run:
	go run main.go

test:
	go test ./...

swag:
	swag init -g main.go -o docs

docker:
	docker-compose up --build

clean:
	docker-compose down -v
	rm -rf docs/swagger.json docs/swagger.yaml docs/docs.go 