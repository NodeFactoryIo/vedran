package repositories

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
)

type DowntimeRepo struct {
	db *storm.DB
}

func NewDowntimeRepo(db *storm.DB) *DowntimeRepo {
	return &DowntimeRepo{
		db: db,
	}
}

func (r *DowntimeRepo) Save(downtime *models.Downtime) error {
	return r.db.Save(downtime)
}
