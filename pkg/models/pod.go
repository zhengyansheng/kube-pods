package models

import (
	"github.com/zhengyansheng/common"
	v1 "k8s.io/api/core/v1"
)

type pod struct {
	BaseModel
	PodName    string      `json:"pod_name"`
	Namespace  string      `json:"namespace"`
	DeployName string      `json:"deploy_name"`
	Phase      v1.PodPhase `json:"phase"`
	PodIP      string      `json:"pod_ip"`
	PodVersion string      `json:"pod_version"`
	NodeName   string      `json:"node_name"`
	HostIP     string      `json:"host_ip"`
}

func NewPod(p *v1.Pod) *pod {
	return &pod{
		PodName:    p.Name,
		Namespace:  p.Namespace,
		DeployName: p.ObjectMeta.Name,
		Phase:      p.Status.Phase,
		PodIP:      p.Status.PodIP,
		PodVersion: "",
		NodeName:   p.Spec.NodeName,
		HostIP:     p.Status.HostIP,
		BaseModel: BaseModel{
			CreateTime: p.ObjectMeta.CreationTimestamp.Format("2006-01-02 15:04:05"),
			ClsName:    "k8s_online",
		},
	}
}

func (p *pod) Notify() {
	common.Indent(p)
}
