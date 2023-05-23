package store

import "sync"

type InternalStore struct {
	locker              *sync.RWMutex
	clusterResourceName map[string][]string
}

type Store interface {
	Add(clusterName string, resourceName string)
	List(clusterName string) []string
}

// Constructor for InternalStore
func NewInternalStore() Store {
	return &InternalStore{
		locker:              &sync.RWMutex{},
		clusterResourceName: make(map[string][]string),
	}
}

// AddNodeGroup to add a node group to the store
func (store *InternalStore) Add(clusterName string, nodeGroupName string) {
	store.locker.Lock()
	defer store.locker.Unlock()
	if store.clusterResourceName == nil {
		store.clusterResourceName = make(map[string][]string)
	}
	store.clusterResourceName[clusterName] = append(store.clusterResourceName[clusterName], nodeGroupName)
}

// GetNodeGroups to get all node groups for a given cluster name
func (store *InternalStore) List(clusterName string) []string {
	store.locker.RLock()
	defer store.locker.RUnlock()
	return store.clusterResourceName[clusterName]
}
