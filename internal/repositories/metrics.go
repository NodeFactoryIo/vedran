package repositories

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
)

type MetricsRepo struct {
	db *storm.DB
}

func NewMetricsRepo(db *storm.DB) *MetricsRepo {
	return &MetricsRepo{
		db: db,
	}
}

func (r *MetricsRepo) FindByID(ID string) (*models.Metrics, error) {
	var metrics *models.Metrics
	err := r.db.One("ID", ID, metrics)
	return metrics, err
}

func (r *MetricsRepo) Save(metrics *models.Metrics) error {
	return r.db.Save(metrics)
}

func (r MetricsRepo) GetAll() (*[]models.Metrics, error) {
	var metrics *[]models.Metrics
	err := r.db.All(metrics)
	return metrics, err
}