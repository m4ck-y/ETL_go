package repository

import (
	"sync"

	"github.com/m4ck-y/ETL_go/internal/domain/models"
)

type InMemoryMetricsRepository struct {
	data             map[models.UTMKey]models.AggregatedMetrics
	processedBatches map[string]bool
	mu               sync.RWMutex
}

func NewInMemoryMetricsRepository() *InMemoryMetricsRepository {
	return &InMemoryMetricsRepository{
		data:             make(map[models.UTMKey]models.AggregatedMetrics),
		processedBatches: make(map[string]bool),
	}
}

func (r *InMemoryMetricsRepository) Save(metrics map[models.UTMKey]models.AggregatedMetrics) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, v := range metrics {
		r.data[k] = v
	}
	return nil
}

func (r *InMemoryMetricsRepository) GetAll() (map[models.UTMKey]models.AggregatedMetrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	copy := make(map[models.UTMKey]models.AggregatedMetrics)
	for k, v := range r.data {
		copy[k] = v
	}
	return copy, nil
}

func (r *InMemoryMetricsRepository) GetByKey(key models.UTMKey) (models.AggregatedMetrics, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, found := r.data[key]
	return value, found, nil
}

func (r *InMemoryMetricsRepository) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data = make(map[models.UTMKey]models.AggregatedMetrics)
	r.processedBatches = make(map[string]bool)
	return nil
}

func (r *InMemoryMetricsRepository) IsBatchProcessed(batchID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	processed, exists := r.processedBatches[batchID]
	return processed && exists, nil
}

func (r *InMemoryMetricsRepository) MarkBatchProcessed(batchID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.processedBatches[batchID] = true
	return nil
}
