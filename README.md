go mod tidy

go run ./cmd

# endpoints

- curl -X POST http://localhost:8080/ingest/run

    respuesta:

    - {"status":"ETL completed"}

- curl http://localhost:8080/metrics



documentacion de la api
swagger:
go get github.com/swaggo/swag/cmd/swag@latest
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files


generar docs
go run github.com/swaggo/swag/cmd/swag@latest init -g ./cmd/main.go


