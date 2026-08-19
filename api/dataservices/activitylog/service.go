package activitylog

import (
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
)

const (
	BucketName = "activity_logs"
	Retention  = 7 * 24 * time.Hour
)

type Service struct {
	dataservices.BaseDataService[portainer.ActivityLog, portainer.ActivityLogID]
}

func NewService(connection portainer.Connection) (*Service, error) {
	if err := connection.SetServiceName(BucketName); err != nil {
		return nil, err
	}

	return &Service{BaseDataService: dataservices.BaseDataService[portainer.ActivityLog, portainer.ActivityLogID]{
		Bucket: BucketName, Connection: connection,
	}}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{BaseDataServiceTx: dataservices.BaseDataServiceTx[portainer.ActivityLog, portainer.ActivityLogID]{
		Bucket: BucketName, Connection: service.Connection, Tx: tx,
	}}
}

func (service *Service) Create(entry *portainer.ActivityLog) error {
	return service.Connection.UpdateTx(func(tx portainer.Transaction) error {
		serviceTx := service.Tx(tx)
		cutoff := time.Now().Add(-Retention).Unix()
		old, err := serviceTx.ReadAll(func(entry portainer.ActivityLog) bool { return entry.Timestamp < cutoff })
		if err != nil {
			return err
		}
		for _, entry := range old {
			if err := serviceTx.Delete(entry.ID); err != nil {
				return err
			}
		}
		return serviceTx.Create(entry)
	})
}
