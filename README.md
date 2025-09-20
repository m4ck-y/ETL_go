go mod tidy

go run ./cmd

# endpoints

- curl -X POST http://localhost:8080/ingest/run

    respuesta:

    - {"status":"ETL completed"}

- curl http://localhost:8080/metrics