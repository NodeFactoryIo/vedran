package repositories

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"time"
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
		q.Eq("id", nodeID),
		q.Or(
			q.And( // start inside interval
				q.Gte("start", from),
				q.Lte("start", to),
			),
			q.And( // end inside interval
				q.Gte("end", from),
				q.Lte("end", to),
			),
		),
	)).Find(downtimes)
	return downtimes, err
}

