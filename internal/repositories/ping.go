package repositories

import (
	"time"

	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
)

type PingRepo struct {
	db *storm.DB
}

func NewPingRepo(db *storm.DB) *PingRepo {
	return &PingRepo{
		db: db,
	}
}

func (r *PingRepo) FindByNodeID(nodeId string) (*models.Ping, error) {
	var ping *models.Ping
	err := r.db.One("NodeId", nodeId, ping)
	return ping, err
}

func (r *PingRepo) Save(ping *models.Ping) error {
	return r.db.Save(ping)
}

func (r PingRepo) GetAll() (*[]models.Ping, error) {
	var pings []models.Ping
	err := r.db.All(&pings)
	return &pings, err
}

func (r PingRepo) CalculateDowntime(nodeId string, pingTime time.Time) (time.Time, time.Duration, error) {
	lastPing, err := r.FindByNodeID(nodeId)
	if err != nil {
		return pingTime, time.Duration(0), err
	}

	return lastPing.Timestamp, lastPing.Timestamp.Sub(pingTime), nil
}
