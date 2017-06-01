package models

type NewObjectRequest struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   Metadata `json:"metadata"`
}

type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}