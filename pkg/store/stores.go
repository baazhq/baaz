package store

type InternalStore struct {
	ClusterNameNodeGroupName map[string][]string
}

type Store interface {
	AddNodeGroup(clusterName string, nodeGroupName string)
	GetNodeGroups(clusterName string) []string
}

// Constructor for InternalStore
func NewInternalStore() Store {
	return &InternalStore{
		ClusterNameNodeGroupName: make(map[string][]string),
	}
}

// Method to add a node group to the store
func (store *InternalStore) AddNodeGroup(clusterName string, nodeGroupName string) {
	if store.ClusterNameNodeGroupName == nil {
		store.ClusterNameNodeGroupName = make(map[string][]string)
	}
	store.ClusterNameNodeGroupName[clusterName] = append(store.ClusterNameNodeGroupName[clusterName], nodeGroupName)
}

// Method to get all node groups for a given cluster name
func (store *InternalStore) GetNodeGroups(clusterName string) []string {
	return store.ClusterNameNodeGroupName[clusterName]
}
