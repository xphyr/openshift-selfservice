package models

type EditQuotasCommand struct {
	ProjectName string `json:projectName`
	CPU int `json:cpu`
	Memory int `json:memory`
}

type Reply struct {
	Status bool `json:status`
	Message string  `json:message`
}
