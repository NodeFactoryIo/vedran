package repositories

import (
	"time"

	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
)

type PingRepository interface {
	FindByNodeID(nodeId string) (*models.Ping, error)
	Save(ping *models.Ping) error
	GetAll() (*[]models.Ping, error)
	ResetAllPings() error
	CalculateDowntime(nodeId string, pingTime time.Time) (time.Time, time.Duration, error)
}

type pingRepo struct {
	db *storm.DB
}

func NewPingRepo(db *storm.DB) PingRepository {
	return &pingRepo{
		db: db,
	}
}

func (r *pingRepo) FindByNodeID(nodeId string) (*models.Ping, error) {
	var ping models.Ping
	err := r.db.One("NodeId", nodeId, &ping)
	return &ping, err
}

func (r *pingRepo) Save(ping *models.Ping) error {
	return r.db.Save(ping)
}

func (r *pingRepo) GetAll() (*[]models.Ping, error) {
	var pings []models.Ping
	err := r.db.All(&pings)
	return &pings, err
}

func (r *pingRepo) ResetAllPings() error {
	pings, err := r.GetAll()
	if err != nil {
		return err
	}

	for _, ping := range *pings {
		newPing := models.Ping{
			NodeId:    ping.NodeId,
			Timestamp: time.Now(),
		}
		_ = r.Save(&newPing)
	}

	return nil
}

func (r *pingRepo) CalculateDowntime(nodeId string, pingTime time.Time) (time.Time, time.Duration, error) {
	lastPing, err := r.FindByNodeID(nodeId)
	if err != nil {
		if err.Error() == "not found" {
			return pingTime, time.Duration(0), nil
		}

		return pingTime, time.Duration(0), err
	}

	return lastPing.Timestamp, lastPing.Timestamp.Sub(pingTime), nil
}
