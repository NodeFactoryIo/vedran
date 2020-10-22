package repositories

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
)

type MetricsRepository interface {
	FindByID(ID string) (*models.Metrics, error)
	Save(metrics *models.Metrics) error
	GetAll() (*[]models.Metrics, error)
	GetLatestBlockMetrics() (*models.LatestBlockMetrics, error)
}

type metricsRepo struct {
	db *storm.DB
}

func NewMetricsRepo(db *storm.DB) MetricsRepository {
	return &metricsRepo{
		db: db,
	}
}

func (r *metricsRepo) FindByID(ID string) (*models.Metrics, error) {
	var metrics models.Metrics
	err := r.db.One("NodeId", ID, &metrics)
	return &metrics, err
}

func (r *metricsRepo) Save(metrics *models.Metrics) error {
	return r.db.Save(metrics)
}

func (r metricsRepo) GetAll() (*[]models.Metrics, error) {
	var metrics []models.Metrics
	err := r.db.All(&metrics)
	return &metrics, err
}

func (r *metricsRepo) GetLatestBlockMetrics() (*models.LatestBlockMetrics, error) {
	all, err := r.GetAll()
	if err != nil {
		return nil, err
	}
	latestBlockMetrics := models.LatestBlockMetrics{
		BestBlockHeight:      0,
		FinalizedBlockHeight: 0,
	}
	for _, m := range *all {
		if m.BestBlockHeight > latestBlockMetrics.BestBlockHeight {
			latestBlockMetrics.BestBlockHeight = m.BestBlockHeight
		}
		if m.FinalizedBlockHeight > latestBlockMetrics.FinalizedBlockHeight {
			latestBlockMetrics.FinalizedBlockHeight = m.FinalizedBlockHeight
		}
	}
	return &latestBlockMetrics, err
}
