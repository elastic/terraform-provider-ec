---
page_title: "Elastic Cloud: ec_azure_privatelink_endpoint"
description: |-
  Retrieves infomation about the Azure Private Link configuration for a given region.
---

# Data Source: ec_azure_privatelink_endpoint

Use this data source to retrieve information about the Azure Private Link configuration for a given region. Further documentation on how to establish a PrivateLink connection can be found in the ESS [documentation](https://www.elastic.co/guide/en/cloud/current/ec-traffic-filtering-vnet.html).

~> **NOTE:** This data source provides data relevant to the Elasticsearch Service (ESS) only, and should not be used for ECE.

## Example Usage

```hcl
data "ec_azure_privatelink_endpoint" "eastus" {
  region = "eastus"
}
```

## Argument Reference

* `region` (Required) - Region to retrieve the Private Link configuration for.

## Attributes Reference

* `service_alias` - The service alias to establish a connection to.
* `domain_name` - The domain name to used in when configuring a private hosted zone in the VNet connection.
