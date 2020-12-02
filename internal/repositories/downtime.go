package repositories

import (
	"time"

	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
)

type DowntimeRepository interface {
	Save(downtime *models.Downtime) error
	// FindDowntimesInsideInterval returns all models.Downtime that started or ended inside interval
	// defined with arguments from and to
	FindDowntimesInsideInterval(nodeID string, from time.Time, to time.Time) ([]models.Downtime, error)
}

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

func (r *DowntimeRepo) FindDowntimesInsideInterval(nodeID string, from time.Time, to time.Time) ([]models.Downtime, error) {
	var downtimes []models.Downtime
	err := r.db.Select(q.And(
		q.Eq("NodeId", nodeID),
		q.Or(
			q.And( // start inside interval
				q.Gte("Start", from),
				q.Lte("Start", to),
			),
			q.And( // end inside interval
				q.Gte("End", from),
				q.Lte("End", to),
			),
		),
	)).Find(&downtimes)
	return downtimes, err
}
