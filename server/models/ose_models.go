package models

type PolicyBindingResponse struct {
	RoleBindings []struct {
		Name        string `json:"name"`
		RoleBinding struct {
				    UserNames []string `json:"userNames"`
			    } `json:"roleBinding"`
	} `json:"roleBindings"`
}

type ResourceQuotaResponse struct {
	Items []struct {
		Metadata struct {
				 Name string `json:"name"`
			 } `json:"metadata"`
		Spec     struct {
				 Hard struct {
					      CPU    string `json:"cpu"`
					      Memory string `json:"memory"`
				      } `json:"hard"`
			 } `json:"spec"`
	} `json:"items"`
}