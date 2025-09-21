package domain

import (
	"github.com/m4ck-y/ETL_go/internal/domain/models"
)

type MetricsRepository interface {
	Save(metrics map[models.UTMKey]models.AggregatedMetrics) error
	GetAll() (map[models.UTMKey]models.AggregatedMetrics, error)
	GetByKey(key models.UTMKey) (models.AggregatedMetrics, bool, error)
	Clear() error
	// Idempotence methods
	IsBatchProcessed(batchID string) (bool, error)
	MarkBatchProcessed(batchID string) error
}
