package models

type PolicyBindingResponse struct {
	Kind string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Items []struct {
		Metadata struct {
				 Name string `json:"name"`
				 Namespace string `json:"namespace"`
				 SelfLink string `json:"selfLink"`
				 UID string `json:"uid"`
				 ResourceVersion string `json:"resourceVersion"`
			 } `json:"metadata"`
		PolicyRef struct {
				 Name string `json:"name"`
			 } `json:"policyRef"`
		RoleBindings []struct {
			Name string `json:"name"`
			RoleBinding struct {
				     Metadata struct {
						      Name string `json:"name"`
						      Namespace string `json:"namespace"`
						      UID string `json:"uid"`
						      ResourceVersion string `json:"resourceVersion"`
					      } `json:"metadata"`
				     UserNames []string `json:"userNames"`
				     GroupNames []string `json:"groupNames"`
				     Subjects []struct {
					     Kind string `json:"kind"`
					     Name string `json:"name"`
				     } `json:"subjects"`
				     RoleRef struct {
						      Name string `json:"name"`
					      } `json:"roleRef"`
			     } `json:"roleBinding"`
		} `json:"roleBindings"`
	} `json:"items"`
}