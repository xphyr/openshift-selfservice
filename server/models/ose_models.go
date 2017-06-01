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

type ProjectResponse struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
			   Name            string `json:"name"`
			   SelfLink        string `json:"selfLink"`
			   UID             string `json:"uid"`
			   ResourceVersion string `json:"resourceVersion"`
			   Annotations     struct {
						   Description             string `json:"openshift.io/description"`
						   DisplayName             string `json:"openshift.io/display-name"`
						   Requester               string `json:"openshift.io/requester"`
						   SaSccMcs                string `json:"openshift.io/sa.scc.mcs"`
						   SaSccSupplementalGroups string `json:"openshift.io/sa.scc.supplemental-groups"`
						   SaSccUIDRange           string `json:"openshift.io/sa.scc.uid-range"`
						   BillingNr               string `json:"openshift.io/kontierung-element"`
						   MegaId                  string `json:"openshift.io/MEGAID"`
					   } `json:"annotations"`
		   } `json:"metadata"`
	Spec       struct {
			   Finalizers []string `json:"finalizers"`
		   } `json:"spec"`
	Status     struct {
			   Phase string `json:"phase"`
		   } `json:"status"`
}

type NewObjectRequest struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   Metadata `json:"metadata"`
}

type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}