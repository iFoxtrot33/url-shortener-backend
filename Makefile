start:
	go run cmd/main.go

build:
	go build -o app cmd/main.go

migration:
	go run migrations/auto.go

swagger:
	go run github.com/swaggo/swag/cmd/swag init -g cmd/main.go
