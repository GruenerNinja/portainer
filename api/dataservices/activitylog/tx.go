package activitylog

import (
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portainer.ActivityLog, portainer.ActivityLogID]
}

func (service ServiceTx) Create(entry *portainer.ActivityLog) error {
	return service.Tx.CreateObject(BucketName, func(id uint64) (int, any) {
		entry.ID = portainer.ActivityLogID(id)
		return int(entry.ID), entry
	})
}
