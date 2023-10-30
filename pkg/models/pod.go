package models

type Pod struct {
	BaseModel
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Phase      string `json:"phase"`
	PodIP      string `json:"pod_ip"`
	PodVersion string `json:"pod_version"`
	NodeName   string `json:"node_name"`
	HostIP     string `json:"host_ip"`
}
