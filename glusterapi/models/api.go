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

type GrowVolumeCommand struct {
	PvName   string `json:"pvName"`
	GrowSize string `json:"growSize"`
}

type VolInfo struct {
	TotalKiloBytes int `json:"totalKiloBytes"`
	UsedKiloBytes  int `json:"usedKiloBytes"`
}
