# Discovery naming conventions

To load balance and route traffic to backend systems, Gimbal must
discover the backends and synchronize them to the Gimbal cluster. This is done
by the Gimbal discovery components -- currently the Kubernetes Discoverer and the
OpenStack Discoverer.

During the discovery process, Gimbal translates the discovered backends into
Kubernetes Services and Endpoints. The _Discovered Name_ of each Service and
Endpoint is formed by concatenating:

```
${backend-name}-${service-name}
```

The name of a service port is not specified, and is handled independently by each
Discoverer implementation.

## Kubernetes Service naming requirements

Kubernetes Service names must adhere to the [rfc1035 DNS Label](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/identifiers.md) specification:

> An alphanumeric (a-z, and 0-9) string, with a maximum length of 63 characters,
> with the '-' character allowed anywhere except the first or last character,
> suitable for use as a hostname or segment in a domain name.

### Handling names that contain invalid characters

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
   characters allowed. The _Discovered Name_ has a total of two components, and
   thus each component is allocated a maximum of 31 characters (one of the 63 is
   used for the separator).

2. Starting at the last _Discovered Name Component_, check whether it is longer
   than the allocated number of characters. If it is, take the SHA256 hash of
   the _Component_. Truncate the excess characters of the _Component_, and
   replace with the first 6 characters of the hash.

3. If the resulting _Discovered Name_ is still longer than 63 characters, move
   onto the next _Component_ and shorten using the approach described above.

4. The shortening process produces the final _Discovered Name_.

#### Example

- `${backend-name}`: `us-east-cluster`
- `${service-name}`: `the-really-long-kube-service-name-that-is-exactly-63-characters`

1. The _Discovered Name_ has a total of two components. Thus, allocate 31
   characters to each component (62/2, as one char is allocated for the
   separator).

2. Check if the last _Component_ of the _Discovered Name_ goes over 31 characters:

    ```
    "the-really-long-kube-service-name-that-is-exactly-63-characters" has 63 characters.
    ```

3. Shorten the last _Component_ as it is longer than 31 characters:

  ```
  Take SHA256 of "the-really-long-kube-service-name-that-is-exactly-63-characters"
  SHA256 hash = 1feeec450b150bcfd731a5e7399890e6c61088fec088eb61f83baa75ca3bd2d9
  Short hash = 1feeec

  Result = "the-really-long-kube-serv1feeec"
  ```

4. Check if the resulting _Discovered Name_ is shorter than 63 characters:

    ```
    "us-east-cluster-the-really-long-kube-serv1feeec" has 47 characters.
    Thus, we have arrived at our shortened Discovered Name.
    ```

Discovered name after shortening: `us-east-cluster-the-really-long-kube-serv1feeec`

## Discoverer Specifics

The specifics of each discoverer are documented below.

### Kubernetes discoverer

- `${backend-name}`: The value of the `--backend-name` flag provided to the
  discoverer. Must begin with a lowercase letter.
- `${service-name}`: The name of the backend service, verbatim.

Service port names are copied verbatim from the backend service.

### OpenStack discoverer

- `${backend-name}`: The value of the `--backend-name` flag provided to the
  discoverer. Must begin with a lowercase letter.
- `${service-name}`: `${name}-${id}` of the LBaaS Load Balancer. Both are
  lowercased during the discovery process.

Service port names are set to `port-${port-number}`.

#### Why is the `${service-name}` a composite name?

The `${service-name}` produced by the OpenStack discoverer is composed of the
name and ID of the LBaaS Load Balancer. This is required because names are not
guaranteed to be unique in an OpenStack project. By appending the ID, we ensure
that we are referncing a single Load Balancer in the OpenStack cluster.

#### What happens to Load Balancers that do not have a name?

Names in OpenStack are optional. In this scenario, the `${service-name}` will
be the ID of the LBaaS Load Balancer.
