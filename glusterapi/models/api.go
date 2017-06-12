package models

type CreateLVCommand struct {
	Size       string `json:"size"`
	MountPoint string `json:"mountPoint"`
	LvName     string `json:"lvName"`
}

type CreateVolumeCommand struct {
	Project string `json:"project"`
	Size    string `json:"size"`
}