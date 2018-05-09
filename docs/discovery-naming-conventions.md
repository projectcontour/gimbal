# Discovery Naming Conventions

In order to load balance and route traffic to backend systems, Gimbal must
discover the backends and sync them to the Gimbal cluster. This is done
by the Gimbal discovery components, such as the Kubernetes discoverer and the
OpenStack discoverer.

During the discovery process, Gimbal translates the discovered backends into
Kubernetes Services and Endpoints. The name of the discovered Services and
Endpoints is called the _Discovered Name_, and is built from the following
_Components_:

```
${discoverer-prefix}-${service-name}-${backend-name}
```

The name of service ports is not specified, and is handled independently by each
discoverer implementation.

## Kubernetes Service Naming Requirements

Kubernetes Service names must adhere to the [rfc1035 DNS Label](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/identifiers.md) specification:

> An alphanumeric (a-z, and 0-9) string, with a maximum length of 63 characters,
> with the '-' character allowed anywhere except the first or last character,
> suitable for use as a hostname or segment in a domain name.

### Handling of invalid characters

Backend systems other than Kubernetes (e.g. OpenStack) might support names that
contain characters that are invalid according to the Kubernetes Service naming
requirements. In these scenarios, discoverers will skip the service and inform
the user via logs and metrics that the backend service name is not compatible
with Gimbal.

Even thought this might affect the user experience and cause friction, Gimbal
prefers to be predictable and safe when it comes to managing backend services
and endpoints, instead of using "magic" (replacing or removing characters, etc)
to massage invalid names into valid ones.

### Handling Discovered Names that are longer than 63 characters

When the _Discovered Name_ is longer than 63 characters, it is shortened using
the following process:

1. Each component of the _Discovered Name_ gets the same number of maximum
   characters allowed. The _Discovered Name_ has a total of three components,
   and thus each component is allocated a maximum of 21 characters.

2. Take the SHA256 hash of the _Discovered Name_ before shortening, and take the
   first 6 characters of the resulting hash. This is called the _short hash_.

3. Starting at the last _Discovered Name Component_, check whether it is longer
   than the allocated number of characters. If it is, truncate the end of the
   _Component_ and append `-${short-hash}`.

4. If the resulting _Discovered Name_ is still longer than 63 characters, move
   onto the next _Component_ and shorten using the _short hash_. Continue until
   the resulting _Discovered Name_ is 63 characters or less.

5. In the case that the _Discovered Name_ is longer than 63 characters after the
   shortening process, the SHA256 hash produced in step 2 will be used as the
   _Discovered Name_, prepended with `g-` to ensure the name starts with a
   letter. This scenario will only happen if *all* _Components_ of the initial
   _Discovered Name_ are longer than 21 characters, which is a rare case.

#### Example

- `${discoverer-prefix}`: `k8s`
- `${service-name}`: `the-really-long-kube-service-name-that-is-exactly-63-characters`
- `${backend-name}`: `us-east-cluster`

1. The _Discovered Name_ has a total of three components. Thus, allocate 21 (61/3) characters to each.

2. Take the SHA256 hash of `k8s-the-really-long-kube-service-name-that-is-exactly-63-characters-us-east-cluster`.

    ```
    SHA256 hash = cdbeb21c3261255a91ded1844754791ef1ec06ecbe5b2421ff2abe221950fc72
    Short hash = cdbeb2
    ```

3. Check if the last _Component_ of the _Discovered Name_ goes over 21 characters:

    ```
    "us-east-cluster" has 15 characters. Move onto next component.
    ```

4. Check if the next _Component_ goes over 21 characters:

    ```
    "the-really-long-kube-service-name-that-is-exactly-63-characters" has 63 characters.
    Shorten using the short hash by truncating and appending the short hash.
    Result: "the-really-lon-cdbeb2".
    ```

5. Check if the resulting _Discovered Name_ is shorter than 63 characters:

    ```
    "k8s-the-really-lon-cdbeb2-us-east-cluster" has 41 characters.
    Thus, we have arrived at our shortened Discovered Name.
    ```

Discovered name after shortening: `k8s-the-really-lon-cdbeb2-us-east-cluster`

## Discoverer Specifics

The specifics of each discoverer are documented below.

### Kubernetes discoverer

- `${discoverer-prefix}`: `k8s`
- `${service-name}`: The name of the backend service, verbatim.
- `${backend-name}`: The value of the `--cluster-name` flag provided to the
  discoverer.

Service port names are copied verbatim from the backend service.

### OpenStack discoverer

- `${discoverer-prefix}`: `os`
- `${service-name}`: `${name}-${id}` of the LBaaS Load Balancer. Both are
  lowercased during the discovery process.
- `${backend-name}`: The value of the `--cluster-name` flag provided to the
  discoverer.

Service port names are set to `${discoverer-prefix}-${port-number}`.

#### Why is the `${service-name}` a composite name?

The `${service-name}` produced by the OpenStack discoverer is composed of the
name and ID of the LBaaS Load Balancer. This is required because names are not
guaranteed to be unique in an OpenStack project.

#### What happens to Load Balancers that do not have a name?

Names in OpenStack are optional. In this scenario, the `${service-name}` will
be the ID of the LBaaS Load Balancer.
