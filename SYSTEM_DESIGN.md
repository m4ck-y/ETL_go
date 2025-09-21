# System Design - ETL Go Service

## Idempotencia & Reprocesamiento
Se usa batch IDs únicos basados en URLs, fecha y timestamp diario. Los lotes procesados se almacenan en memoria para evitar re-ejecuciones duplicadas.

## Particionamiento & Retención
Datos particionados por UTM keys. Retención configurable por variable de entorno, con endpoint de limpieza manual.

## Concurrencia & Throughput
Procesamiento síncrono con timeouts de 30s. Sin goroutines paralelas por simplicidad. Throughput limitado por memoria.

## Calidad de Datos
UTMs normalizados a lowercase con fallbacks ("unknown_campaign", etc.). Fechas validadas con múltiples formatos.

## Observabilidad
Logs estructurados en JSON con request IDs. Health checks básicos. Sin métricas Prometheus implementadas.

## Evolución en el Ecosistema Admira
Interfaz de repositorio permite migrar a BD. Diseño modular facilita agregar nuevas fuentes. APIs documentadas con Swagger.