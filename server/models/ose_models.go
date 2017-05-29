package models

type PolicyBindingResponse struct {
	Kind         string `json:"kind"`
	APIVersion   string `json:"apiVersion"`
	Metadata     struct {
			     Name            string `json:"name"`
			     Namespace       string `json:"namespace"`
			     ResourceVersion string `json:"resourceVersion"`
		     } `json:"metadata"`
	PolicyRef    struct {
			     Name string `json:"name"`
		     } `json:"policyRef"`
	RoleBindings []struct {
		Name        string `json:"name"`
		RoleBinding struct {
				    Metadata   struct {
						       Name            string `json:"name"`
						       Namespace       string `json:"namespace"`
						       ResourceVersion string `json:"resourceVersion"`
					       } `json:"metadata"`
				    UserNames  []string `json:"userNames"`
				    GroupNames interface{} `json:"groupNames"`
				    Subjects   []struct {
					    Kind      string `json:"kind"`
					    Namespace string `json:"namespace,omitempty"`
					    Name      string `json:"name"`
				    } `json:"subjects"`
				    RoleRef    struct {
						       Name string `json:"name"`
					       } `json:"roleRef"`
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

type ProjectRequest struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   Metadata `json:"metadata"`
}

type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}