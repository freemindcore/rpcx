package plugin

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

//ZooKeeperRegisterPlugin a register plugin which can register services into zookeeper for cluster
type ZooKeeperRegisterPlugin struct {
	ServiceAddress   string
	ZooKeeperServers []string
	BasePath         string
	Conn             *zk.Conn
}

// Start starts to connect zookeeper cluster
func (plugin *ZooKeeperRegisterPlugin) Start() (err error) {
	conn, _, err := zk.Connect(plugin.ZooKeeperServers, time.Second)
	plugin.Conn = conn
	return
}

//Close closes zookeeper connection.
func (plugin *ZooKeeperRegisterPlugin) Close() {
	plugin.Conn.Close()
}

func mkdirs(conn *zk.Conn, path string) (err error) {
	if path == "" {
		return errors.New("path should not been empty")
	}
	if path == "/" {
		return nil
	}
	if path[0] != '/' {
		return errors.New("path must start with /")
	}

	//check whether this path exists
	exist, _, err := conn.Exists(path)
	if exist {
		return nil
	}
	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)
	_, err = conn.Create(path, []byte(""), flags, acl)
	if err == nil { //created successfully
		return
	}

	//create parent
	paths := strings.Split(path[1:], "/")
	createdPath := ""
	for _, p := range paths {
		createdPath = createdPath + "/" + p
		exist, _, err = conn.Exists(createdPath)
		if !exist {
			path, err = conn.Create(createdPath, []byte(""), flags, acl)
			if err != nil {
				return
			}
		}
	}

	return nil
}

// Register handles registering event.
// this service is registered at BASE/serviceName/thisIpAddress node
func (plugin *ZooKeeperRegisterPlugin) Register(name string, rcvr interface{}) (err error) {
	nodePath := plugin.BasePath + "/" + name
	err = mkdirs(plugin.Conn, nodePath)
	if err != nil {
		return err
	}

	nodePath = nodePath + "/" + plugin.ServiceAddress
	//delete existed node
	exists, _, err := plugin.Conn.Exists(nodePath)
	if exists {
		err = plugin.Conn.Delete(nodePath, -1)
		fmt.Printf("delete: ok\n")
	}

	//create Ephemeral node
	flags := int32(zk.FlagEphemeral)
	acl := zk.WorldACL(zk.PermAll)
	_, err = plugin.Conn.Create(nodePath, []byte(""), flags, acl)
	return
}

// Unregister a service from zookeeper but this service still exists in this node.
func (plugin *ZooKeeperRegisterPlugin) Unregister(name string) {
	nodePath := plugin.BasePath + "/" + name + "/" + plugin.ServiceAddress

	//delete existed node
	exists, _, _ := plugin.Conn.Exists(nodePath)
	if exists {
		err := plugin.Conn.Delete(nodePath, -1)
		if err != nil {
			fmt.Printf("delete: false because of %v\n", err)
		} else {
			fmt.Printf("delete: ok\n")
		}

	}
}

// Name return name of this plugin.
func (plugin *ZooKeeperRegisterPlugin) Name() string {
	return "ZooKeeperRegisterPlugin"
}

// Description return description of this plugin.
func (plugin *ZooKeeperRegisterPlugin) Description() string {
	return "a register plugin which can register services into zookeeper for cluster"
}
