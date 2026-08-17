package tag

import (
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "tags"

// Service represents a service for managing environment(endpoint) data.
type Service struct {
	// Embedding the generic base service supplies Read, ReadAll, Update, and Delete
	// for Tag values, similar to extending a generic repository in Java.
	dataservices.BaseDataService[portainer.Tag, portainer.TagID]
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portainer.Tag, portainer.TagID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	// The transaction-scoped service reuses the caller's transaction instead of
	// opening and committing a separate one.
	return ServiceTx{
		BaseDataServiceTx: dataservices.BaseDataServiceTx[portainer.Tag, portainer.TagID]{
			Bucket:     BucketName,
			Connection: service.Connection,
			Tx:         tx,
		},
	}
}

// CreateTag creates a new tag.
func (service *Service) Create(tag *portainer.Tag) error {
	// CreateObject allocates the numeric ID; the callback stores it on the tag and
	// returns the key/value pair that should be persisted.
	return service.Connection.CreateObject(
		BucketName,
		func(id uint64) (int, any) {
			tag.ID = portainer.TagID(id)
			return int(tag.ID), tag
		},
	)
}

// UpdateTagFunc updates a tag inside a transaction avoiding data races.
func (service *Service) UpdateTagFunc(ID portainer.TagID, updateFunc func(tag *portainer.Tag)) error {
	id := service.Connection.ConvertToKey(int(ID))
	tag := &portainer.Tag{}

	return service.Connection.UpdateObjectFunc(BucketName, id, tag, func() {
		updateFunc(tag)
	})
}
