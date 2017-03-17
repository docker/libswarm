package store

import (
	"github.com/docker/swarmkit/api"
	"github.com/docker/swarmkit/manager/state"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/pkg/errors"
)

const tableResource = "resource"

func init() {
	register(ObjectStoreConfig{
		Name: tableResource,
		Table: &memdb.TableSchema{
			Name: tableResource,
			Indexes: map[string]*memdb.IndexSchema{
				indexID: {
					Name:    indexID,
					Unique:  true,
					Indexer: resourceIndexerByID{},
				},
				indexName: {
					Name:    indexName,
					Unique:  true,
					Indexer: resourceIndexerByName{},
				},
				indexKind: {
					Name:    indexKind,
					Indexer: resourceIndexerByKind{},
				},
				indexCustom: {
					Name:         indexCustom,
					Indexer:      resourceCustomIndexer{},
					AllowMissing: true,
				},
			},
		},
		Save: func(tx ReadTx, snapshot *api.StoreSnapshot) error {
			var err error
			snapshot.Resources, err = FindResources(tx, All)
			return err
		},
		Restore: func(tx Tx, snapshot *api.StoreSnapshot) error {
			resources, err := FindResources(tx, All)
			if err != nil {
				return err
			}
			for _, r := range resources {
				if err := DeleteResource(tx, r.ID); err != nil {
					return err
				}
			}
			for _, r := range snapshot.Resources {
				if err := CreateResource(tx, r); err != nil {
					return err
				}
			}
			return nil
		},
		ApplyStoreAction: func(tx Tx, sa *api.StoreAction) error {
			switch v := sa.Target.(type) {
			case *api.StoreAction_Resource:
				obj := v.Resource
				switch sa.Action {
				case api.StoreActionKindCreate:
					return CreateResource(tx, obj)
				case api.StoreActionKindUpdate:
					return UpdateResource(tx, obj)
				case api.StoreActionKindRemove:
					return DeleteResource(tx, obj.ID)
				}
			}
			return errUnknownStoreAction
		},
		NewStoreAction: func(c state.Event) (api.StoreAction, error) {
			var sa api.StoreAction
			switch v := c.(type) {
			case state.EventCreateResource:
				sa.Action = api.StoreActionKindCreate
				sa.Target = &api.StoreAction_Resource{
					Resource: v.Resource,
				}
			case state.EventUpdateResource:
				sa.Action = api.StoreActionKindUpdate
				sa.Target = &api.StoreAction_Resource{
					Resource: v.Resource,
				}
			case state.EventDeleteResource:
				sa.Action = api.StoreActionKindRemove
				sa.Target = &api.StoreAction_Resource{
					Resource: v.Resource,
				}
			default:
				return api.StoreAction{}, errUnknownStoreAction
			}
			return sa, nil
		},
	})
}

type resourceEntry struct {
	*api.Resource
}

func (r resourceEntry) ID() string {
	return r.Resource.ID
}

func (r resourceEntry) Meta() api.Meta {
	return r.Resource.Meta
}

func (r resourceEntry) SetMeta(meta api.Meta) {
	r.Resource.Meta = meta
}

func (r resourceEntry) Copy() Object {
	return resourceEntry{r.Resource.Copy()}
}

func (r resourceEntry) EventCreate() state.Event {
	return state.EventCreateResource{Resource: r.Resource}
}

func (r resourceEntry) EventUpdate() state.Event {
	return state.EventUpdateResource{Resource: r.Resource}
}

func (r resourceEntry) EventDelete() state.Event {
	return state.EventDeleteResource{Resource: r.Resource}
}

func confirmExtension(tx Tx, r *api.Resource) error {
	// There must be an extension corresponding to the Kind field.
	extensions, err := FindExtensions(tx, ByName(r.Kind))
	if err != nil {
		return errors.Wrap(err, "failed to query extensions")
	}
	if len(extensions) == 0 {
		return errors.Errorf("object kind %s is unregistered", r.Kind)
	}
	return nil
}

// CreateResource adds a new resource object to the store.
// Returns ErrExist if the ID is already taken.
func CreateResource(tx Tx, r *api.Resource) error {
	if err := confirmExtension(tx, r); err != nil {
		return err
	}
	return tx.create(tableResource, resourceEntry{r})
}

// UpdateResource updates an existing resource object in the store.
// Returns ErrNotExist if the object doesn't exist.
func UpdateResource(tx Tx, r *api.Resource) error {
	if err := confirmExtension(tx, r); err != nil {
		return err
	}
	return tx.update(tableResource, resourceEntry{r})
}

// DeleteResource removes a resource object from the store.
// Returns ErrNotExist if the object doesn't exist.
func DeleteResource(tx Tx, id string) error {
	return tx.delete(tableResource, id)
}

// GetResource looks up a resource object by ID.
// Returns nil if the object doesn't exist.
func GetResource(tx ReadTx, id string) *api.Resource {
	r := tx.get(tableResource, id)
	if r == nil {
		return nil
	}
	return r.(resourceEntry).Resource
}

// FindResources selects a set of resource objects and returns them.
func FindResources(tx ReadTx, by By) ([]*api.Resource, error) {
	checkType := func(by By) error {
		switch by.(type) {
		case byIDPrefix, byName, byKind, byCustom, byCustomPrefix:
			return nil
		default:
			return ErrInvalidFindBy
		}
	}

	resourceList := []*api.Resource{}
	appendResult := func(o Object) {
		resourceList = append(resourceList, o.(resourceEntry).Resource)
	}

	err := tx.find(tableResource, by, checkType, appendResult)
	return resourceList, err
}

type resourceIndexerByID struct{}

func (ri resourceIndexerByID) FromArgs(args ...interface{}) ([]byte, error) {
	return fromArgs(args...)
}

func (ri resourceIndexerByID) FromObject(obj interface{}) (bool, []byte, error) {
	r, ok := obj.(resourceEntry)
	if !ok {
		panic("unexpected type passed to FromObject")
	}

	// Add the null character as a terminator
	val := r.Resource.ID + "\x00"
	return true, []byte(val), nil
}

func (ri resourceIndexerByID) PrefixFromArgs(args ...interface{}) ([]byte, error) {
	return prefixFromArgs(args...)
}

type resourceIndexerByName struct{}

func (ri resourceIndexerByName) FromArgs(args ...interface{}) ([]byte, error) {
	return fromArgs(args...)
}

func (ri resourceIndexerByName) FromObject(obj interface{}) (bool, []byte, error) {
	r, ok := obj.(resourceEntry)
	if !ok {
		panic("unexpected type passed to FromObject")
	}

	// Add the null character as a terminator
	val := r.Resource.Annotations.Name + "\x00"
	return true, []byte(val), nil
}

type resourceIndexerByKind struct{}

func (ri resourceIndexerByKind) FromArgs(args ...interface{}) ([]byte, error) {
	return fromArgs(args...)
}

func (ri resourceIndexerByKind) FromObject(obj interface{}) (bool, []byte, error) {
	r, ok := obj.(resourceEntry)
	if !ok {
		panic("unexpected type passed to FromObject")
	}

	// Add the null character as a terminator
	val := r.Resource.Kind + "\x00"
	return true, []byte(val), nil
}

type resourceCustomIndexer struct{}

func (ri resourceCustomIndexer) FromArgs(args ...interface{}) ([]byte, error) {
	return fromArgs(args...)
}

func (ri resourceCustomIndexer) FromObject(obj interface{}) (bool, [][]byte, error) {
	r, ok := obj.(resourceEntry)
	if !ok {
		panic("unexpected type passed to FromObject")
	}

	return customIndexer(r.Kind, &r.Annotations)
}

func (ri resourceCustomIndexer) PrefixFromArgs(args ...interface{}) ([]byte, error) {
	return prefixFromArgs(args...)
}
