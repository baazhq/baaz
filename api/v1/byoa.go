package v1

type ApplicationType string

type NodeGroupName string

// druid
const (
	Druid       ApplicationType = "druid"
	DataNodes   NodeGroupName   = "datanodes"
	QueryNodes  NodeGroupName   = "querynodes"
	MasterNodes NodeGroupName   = "masternodes"
)

// zookeeper
const (
	Zookeeper ApplicationType = "zookeeper"
	ZkNodes   NodeGroupName   = "zookeepernodes"
)
