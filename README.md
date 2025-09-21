# ETL Go Service

Servicio que extrae datos de APIs de Ads y CRM, los procesa y calcula metricas de negocio.

## Setup

```bash
go mod tidy
go run ./cmd
```

## Endpoints

### Ingestar datos
```bash
curl -X POST http://localhost:8080/ingest/run
# o con filtro de fecha
curl -X POST "http://localhost:8080/ingest/run?since=2025-08-01"
```

Respuesta:
```json
{"status":"ETL completed"}
```

### Ver metricas
```bash
curl http://localhost:8080/metrics
```

## Lo que hace

- ETL completo con descarga de Ads/CRM
- Filtrado por fecha con parametro `since`
- Reintentos automaticos con backoff
- Idempotencia (no duplica datos)
- Calculo de metricas: CPC, CPA, CVR, ROAS
- Normalizacion de datos y UTMs

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

