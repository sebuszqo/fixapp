POSTGRES_URL=postgres://fixapp:fixapp@localhost:5432/fixapp?sslmode=disable

run:
	APP_NAME=fixapp LOG_LEVEL=debug go run ./cmd/api/main.go

swagger:
	swag init -g cmd/api/main.go