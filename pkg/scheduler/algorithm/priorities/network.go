package priorities

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
)

func NetworkPriorityMap(pod *v1.Pod, meta interface{}, nodeInfo *schedulercache.NodeInfo) (schedulerapi.HostPriority, error) {
	type Length struct {
		Len float64 `json:"len"`
	}

	node := nodeInfo.Node()
	if node == nil {
		return schedulerapi.HostPriority{}, fmt.Errorf("node not found")
	}

	addrs := node.Status.Addresses
	var hostIp string
	for _, addr := range addrs {
		if addr.Type == v1.NodeInternalIP {
			hostIp = addr.Address
			break
		}
	}
	url := "http://" + hostIp + ":8080/as_len"

	resp, err := http.Get(url)
	if err != nil {
		klog.Error(err)
	}
	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Error(err)
	}

	var l Length
	_ = json.Unmarshal(respByte, &l)

	var score int
	score = 100 - int(l.Len) // as pathが小さいほどスコアを上げたい, as pathが100超えると壊れる
	// klog.Infof("l, s: %+v %+v", l.Len, score)

	return schedulerapi.HostPriority{
		Host:  node.Name,
		Score: score,
	}, nil
}
