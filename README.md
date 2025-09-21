# ETL Go Service

Servicio ETL que extrae datos de Ads y CRM, calcula métricas de negocio.

## Setup

```bash
go mod tidy
go run ./cmd
```

## Docker

```bash
docker-compose up --build
```

## Endpoints


## Endpoints

### Ingestar datos
```bash
curl -X POST http://localhost:8080/ingest/run
# o con filtro de fecha
curl -X POST "http://localhost:8080/ingest/run?since=2025-08-01"
```

### Ver metricas
```bash
curl http://localhost:8080/metrics
```

## Documentacion API

### Swagger
```bash
go get github.com/swaggo/swag/cmd/swag@latest
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
```

### Generar docs
```bash
go run github.com/swaggo/swag/cmd/swag@latest init -g ./cmd/main.go
```

Despues acceder a: http://localhost:8080/swagger/index.html

