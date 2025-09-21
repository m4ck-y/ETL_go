# ETL Go Service

Servicio ETL que extrae datos de Ads y CRM, calcula métricas de negocio.

## Docker

```bash
docker-compose up --build
```

## Endpoints

### Ingestar datos
```bash
curl -X POST http://localhost:8080/ingest/run
# o con filtro de fecha
curl -X POST "http://localhost:8080/ingest/run?since=2025-08-01"
```

### Resetear datos
```bash
curl -X POST http://localhost:8080/admin/reset
```

### Health checks
```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

### Ver metricas
```bash
curl http://localhost:8080/metrics
```

## Documentacion API

Se opto por documentacion con Swagger en lugar de Postman para las pruebas interactivas de la API.

La documentacion se genera automaticamente con docker-compose.

### Comandos para desarrollo
```bash
go get github.com/swaggo/swag/cmd/swag@latest
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files

go run github.com/swaggo/swag/cmd/swag@latest init -g ./cmd/main.go
```

Despues acceder a: http://localhost:8080/swagger/index.html

### Despliegue
Desplegado en Google Cloud Run vinculado directamente con el repositorio de GitHub.

![Configuración GCP](https://storage.googleapis.com/etl_go/config_gcp_run.png)

URL en GCP: https://etl-go-967885369144.europe-west1.run.app/swagger


![Swagger UI](https://storage.googleapis.com/etl_go/swagger.png)
