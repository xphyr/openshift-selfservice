# General idea
We at [@SchweizerischeBundesbahnen](https://github.com/SchweizerischeBundesbahnen) have a lot of projects who need changes on their projects all the time. As those settings are (and also should ;-)) limited to the administrator roles, we had to do a lot of manual changes like:
- Creating new projects with certain attributes
- Updating project quotas
- Creating service-accounts
- Update project billing information

Persistent storage:
- Create gluster volumes
- Create PV, PVC, Gluster Endpoint & Service in OpenShift

So we built this tool which allows users to do certain things in self service. The tool checks permissions & certain conditions.

# Components
- The Self-Service-Portal (as a container)
- The GlusterFS-API server

# Installation & Documentation
## Self-Service Portal
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
      - persistentvolumes
      - persistentvolumeclaims
      verbs:
      - create
    - apiGroups: null
      attributeRestrictions: null
      resources:
      - services
      - endpoints
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

Just create a 'oc new-app' from the dockerfile.

### Parameters
**Param**|**Description**|**Example**
:-----:|:-----:|:-----:
LDAP\_URL|Your LDAP|ldap.xzw.ch
LDAP\_BIND\_DN|LDAP Bind|cn=root
LDAP\_BIND\_CRED|LDAP Credentials|secret
LDAP\_SEARCH\_BASE|LDAP Search Base|ou=passport-ldapauth
LDAP\_FILTER|LDAP Filter|(uid=%s)
SESSION\_KEY|A secret password to encrypt session information|secret
OPENSHIFT\_API\_URL|Your OpenShift API Url|https://master01.ch:8443
OPENSHIFT\_TOKEN|The token from the service-account| 
GIN\_MODE|Mode of the Webframework|debug/release
MAX\_CPU|How many CPU can a user assign to his project|30
MAX\_MEMORY|How many GB memory can a user assign to his project|50
GLUSTER\_API\_URL|The URL of your Gluster-API|http://glusterserver01:80
GLUSTER\_SECRET|The basic auth password you configured on the gluster api|secret
GLUSTER\_IPS|IP addresses of the gluster endpoints|192.168.1.1,192.168.1.2

## The GlusterFS api
Use/see the service unit file in ./install/

## Monitoring endpoints
The gluster api has two public endpoints for monitoring purposes. Call them this way:

The first endpoint returns usage statistics:
```bash
curl <yourserver>:<port>/volume/<volume-name>
{"totalKiloBytes":123520,"usedKiloBytes":5472}
```

The check endpoint returns if the current %-usage is below the defined threshold:
```bash

# Successful response
curl -i <yourserver>:<port>/volume/<volume-name>/check\?threshold=20
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 12 Jun 2017 14:23:53 GMT
Content-Length: 38

{"message":"Usage is below threshold"}

# Error response
curl -i <yourserver>:<port>/volume/<volume-name>/check\?threshold=3

HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8
Date: Mon, 12 Jun 2017 14:23:37 GMT
Content-Length: 70
{"message":"Error used 4.430051813471502 is bigger than threshold: 3"}
```
