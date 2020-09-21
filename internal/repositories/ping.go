package repositories

import (
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