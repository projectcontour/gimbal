# Teams with Gimbal

A key feature of Gimbal is team management. Teams should be able to configure and define their own Ingress resources within the Gimbal cluster without requiring an administrator to assist. To enable this, users should be allowed access only to specified namespaces in the Gimbal cluster. Within their respective namespaces, team members should be granted specific authorization to create Ingress routes and to view Services and Endpoints. 

## RBAC rules 

A key component of any secure Kubernetes cluster are permissions implemented with Role Based Access Control (RBAC). 

Sample RBAC `ClusterRole` that can be assigned to users in a team namespace:

```yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: team-ingress
rules:
- apiGroups:
  - ""
  resources:
  - services
  - endpoints
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - "*"
```