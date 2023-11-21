# Kubernetes resource cache

This project is uses an inmemory cache with kubernetes dynamic watchers to cache custom resources instead of going to api for every request. Resources are cache based on namespace and a cache enabled label.

# Build

Build the kubectl-resourcecache plugin and install it in your GOPATH:

```
make cli
```

Test out with custom resource

```
make test
```

# Usage

**Note:** This example uses [cert-manager](https://cert-manager.io/docs/installation/), make sure you have the CRDs installed in your cluster.

1. Install the kubectl plugin.
2. Apply test data
  ```sh
  kubectl apply -f testdata/test-certissuer.yaml
  ```
3. Use `resource` command in the plugin and pass in the group, version, kind and namespace of the resource
   ```sh
   kubectl resourcecache resource 
   ```

   Output: 
   ```
   Group: cert-manager.io
   Version: v1
   Kind: issuers
   Namespace: default
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  101584μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  256μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  211μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  297μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  188μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  188μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  216μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  380μs
   ...
   ...
   ```

4. Use `external` command in the plugin and pass in the group, version, kind and namespace of the resource 
   ```sh
   kubectl resourcecache extenal 
   ```

   Output:
   ```
   URL: http://worldtimeapi.org/api/timezone/Asia/Kolkata
   CABundle: certs.crt
   RefreshInterval: 3
   Fetched data from external url
   resources found from url: http://worldtimeapi.org/api/timezone/Asia/Kolkata body: map[abbreviation:IST client_ip:122.167.123.74 datetime:2023-11-21T19:33:33.420928+05:30 day_of_week:2 day_of_year:325 dst:false dst_from:<nil> dst_offset:0 dst_until:<nil> raw_offset:19800 timezone:Asia/Kolkata unixtime:1.700575413e+09 utc_datetime:2023-11-21T14:03:33.420928+00:00 utc_offset:+05:30 week_number:47]
   Time taken:  3212682μs
   resources found from url: http://worldtimeapi.org/api/timezone/Asia/Kolkata body: map[abbreviation:IST client_ip:122.167.123.74 datetime:2023-11-21T19:33:33.420928+05:30 day_of_week:2 day_of_year:325 dst:false dst_from:<nil> dst_offset:0 dst_until:<nil> raw_offset:19800 timezone:Asia/Kolkata unixtime:1.700575413e+09 utc_datetime:2023-11-21T14:03:33.420928+00:00 utc_offset:+05:30 week_number:47]
   Time taken:  129μs
   resources found from url: http://worldtimeapi.org/api/timezone/Asia/Kolkata body: map[abbreviation:IST client_ip:122.167.123.74 datetime:2023-11-21T19:33:33.420928+05:30 day_of_week:2 day_of_year:325 dst:false dst_from:<nil> dst_offset:0 dst_until:<nil> raw_offset:19800 timezone:Asia/Kolkata unixtime:1.700575413e+09 utc_datetime:2023-11-21T14:03:33.420928+00:00 utc_offset:+05:30 week_number:47]
   Time taken:  133μs
   Fetched data from external url
   resources found from url: http://worldtimeapi.org/api/timezone/Asia/Kolkata body: map[abbreviation:IST client_ip:122.167.123.74 datetime:2023-11-21T19:33:36.294082+05:30 day_of_week:2 day_of_year:325 dst:false dst_from:<nil> dst_offset:0 dst_until:<nil> raw_offset:19800 timezone:Asia/Kolkata unixtime:1.700575416e+09 utc_datetime:2023-11-21T14:03:36.294082+00:00 utc_offset:+05:30 week_number:47]
   Time taken:  62μs
   resources found from url: http://worldtimeapi.org/api/timezone/Asia/Kolkata body: map[abbreviation:IST client_ip:122.167.123.74 datetime:2023-11-21T19:33:36.294082+05:30 day_of_week:2 day_of_year:325 dst:false dst_from:<nil> dst_offset:0 dst_until:<nil> raw_offset:19800 timezone:Asia/Kolkata unixtime:1.700575416e+09 utc_datetime:2023-11-21T14:03:36.294082+00:00 utc_offset:+05:30 week_number:47]
   Time taken:  49μs
   resources found from url: http://worldtimeapi.org/api/timezone/Asia/Kolkata body: map[abbreviation:IST client_ip:122.167.123.74 datetime:2023-11-21T19:33:36.294082+05:30 day_of_week:2 day_of_year:325 dst:false dst_from:<nil> dst_offset:0 dst_until:<nil> raw_offset:19800 timezone:Asia/Kolkata unixtime:1.700575416e+09 utc_datetime:2023-11-21T14:03:36.294082+00:00 utc_offset:+05:30 week_number:47]
   Time taken:  47μs
   Fetched data from external url
   resources found from url: http://worldtimeapi.org/api/timezone/Asia/Kolkata body: map[abbreviation:IST client_ip:122.167.123.74 datetime:2023-11-21T19:33:39.293144+05:30 day_of_week:2 day_of_year:325 dst:false dst_from:<nil> dst_offset:0 dst_until:<nil> raw_offset:19800 timezone:Asia/Kolkata unixtime:1.700575419e+09 utc_datetime:2023-11-21T14:03:39.293144+00:00 utc_offset:+05:30 week_number:47]
   Time taken:  171μs
   resources found from url: http://worldtimeapi.org/api/timezone/Asia/Kolkata body: map[abbreviation:IST client_ip:122.167.123.74 datetime:2023-11-21T19:33:39.293144+05:30 day_of_week:2 day_of_year:325 dst:false dst_from:<nil> dst_offset:0 dst_until:<nil> raw_offset:19800 timezone:Asia/Kolkata unixtime:1.700575419e+09 utc_datetime:2023-11-21T14:03:39.293144+00:00 utc_offset:+05:30 week_number:47]
   Time taken:  41μs
   resources found from url: http://worldtimeapi.org/api/timezone/Asia/Kolkata body: map[abbreviation:IST client_ip:122.167.123.74 datetime:2023-11-21T19:33:39.293144+05:30 day_of_week:2 day_of_year:325 dst:false dst_from:<nil> dst_offset:0 dst_until:<nil> raw_offset:19800 timezone:Asia/Kolkata unixtime:1.700575419e+09 utc_datetime:2023-11-21T14:03:39.293144+00:00 utc_offset:+05:30 week_number:47]
   Time taken:  40μs
   Fetched data from external url
   resources found from url: http://worldtimeapi.org/api/timezone/Asia/Kolkata body: map[abbreviation:IST client_ip:122.167.123.74 datetime:2023-11-21T19:33:42.291786+05:30 day_of_week:2 day_of_year:325 dst:false dst_from:<nil> dst_offset:0 dst_until:<nil> raw_offset:19800 timezone:Asia/Kolkata unixtime:1.700575422e+09 utc_datetime:2023-11-21T14:03:42.291786+00:00 utc_offset:+05:30 week_number:47]
   Time taken:  38μs
   resources found from url: http://worldtimeapi.org/api/timezone/Asia/Kolkata body: map[abbreviation:IST client_ip:122.167.123.74 datetime:2023-11-21T19:33:42.291786+05:30 day_of_week:2 day_of_year:325 dst:false dst_from:<nil> dst_offset:0 dst_until:<nil> raw_offset:19800 timezone:Asia/Kolkata unixtime:1.700575422e+09 utc_datetime:2023-11-21T14:03:42.291786+00:00 utc_offset:+05:30 week_number:47]
   Time taken:  92μs
   resources found from url: http://worldtimeapi.org/api/timezone/Asia/Kolkata body: map[abbreviation:IST client_ip:122.167.123.74 datetime:2023-11-21T19:33:42.291786+05:30 day_of_week:2 day_of_year:325 dst:false dst_from:<nil> dst_offset:0 dst_until:<nil> raw_offset:19800 timezone:Asia/Kolkata unixtime:1.700575422e+09 utc_datetime:2023-11-21T14:03:42.291786+00:00 utc_offset:+05:30 week_number:47]
   Time taken:  92μs
   ```