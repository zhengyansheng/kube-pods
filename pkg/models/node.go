package models

type Node struct {
	BaseModel
	Name       string `json:"name"`
	Status     string `json:"status"`
	Version    string `json:"version"`
	InternalIP string `json:"internal_ip"`
	Capacity   string `json:"capacity"`
}
