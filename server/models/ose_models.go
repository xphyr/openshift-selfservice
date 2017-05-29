package models

type PolicyBindingResponse struct {
	RoleBindings []struct {
		Name string `json:"name"`
		RoleBinding struct {
			     UserNames []string `json:"userNames"`
		     } `json:"roleBinding"`
	} `json:"roleBindings"`
}

