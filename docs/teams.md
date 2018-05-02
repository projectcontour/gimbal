# Teams with Gimbal

A key feature of Gimbal is team management. The idea is enable teams to configure and define their own Ingress resources without requiring an administrator to assist. To enable this, users will be isolated to one or more namespaces and should have capabilities to create Ingress routes and also view Services and Endpoints within their respective namespace.

## RBAC Rules 

A key component of any secure Kubernetes cluster are permissions implemented via Role-Based Access Control (RBAC). Following is a sample RBAC `ClusterRole` which can be assigned to users within a team namespace:

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