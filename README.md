# General idea
We at [@SchweizerischeBundesbahnen](https://github.com/SchweizerischeBundesbahnen) have a lot of projects who need changes on their projects all the time. As those settings are (and also should ;-)) limited to the administrator roles, we had to do a lot of manual changes like:
- Creating new projects with certain attributes
- Updating project quotas
- Creating service-accounts
- Update project billing information

So we built this tool which allows users to do certain things in self service. The tool checks permissions & certain conditions.

# Installation
```bash
# Create a project & a service-account
oc new-project ose-selfservice
oc create serviceaccount ose-selfservice

# Add a new role to your cluster-policy:
oc edit clusterPolicy default

###
- name: ose:selfservice
  role:
    metadata:
      creationTimestamp: null
      name: ose:selfservice
    rules:
    - apiGroups: null
      attributeRestrictions: null
      resources:
      - policybindings
      verbs:
      - get
      - list
      - update
    - apiGroups: null
      attributeRestrictions: null
      resources:
      - resourcequotas
      verbs:
      - get
      - list
      - update
    - apiGroups: null
      attributeRestrictions: null
      resources:
      - namespaces
      verbs:
      - get
      - update
    - apiGroups: null
      attributeRestrictions: null
      resources:
      - serviceaccounts
      verbs:
      - create
    - apiGroups: null
      attributeRestrictions: null
      resources:
      - serviceaccounts
      verbs:
      - create
    - apiGroups: null
      attributeRestrictions: null
      resources:
      - projectrequests
      verbs:
      - create
###

# Add the role to the service-account
oc adm policy add-cluster-role-to-user ose:selfservice system:serviceaccount:ose-selfservice:ose-selfservice
```

Just create a 'oc new-app' from building the dockerfile.

## Parameters
**Param**|**Description**|**Example**
:-----:|:-----:|:-----:
LDAP\_URL|Your LDAP|ldap://xzw.ch:389
LDAP\_BIND\_DN|LDAP Bind|cn=root
LDAP\_BIND\_CRED|LDAP Credentials|secret
LDAP\_SEARCH\_BASE|LDAP Search Base|ou=passport-ldapauth
LDAP\_FILTER|LDAP Filter|(uid={{username}})

# NEW
SESSION|_KEY|A secret password to encrypt session information|secret
OPENSHIFT\_API\_URL|Your OpenShift API Url|https://master01.ch:8443
OPENSHIFT\_TOKEN|The token from the service-account| 
GIN\_MODE|Mode of the Webframework|debug/release
MAX\_CPU|How many CPU can a user assign to his project|30
MAX\_MEMORY|How many GB memory can a user assign to his project|50

