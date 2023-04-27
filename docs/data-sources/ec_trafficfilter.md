---
page_title: "Elastic Cloud: ec_trafficfilter"
description: |-
  Filters for available traffic filters that match the given criteria
---

# Data Source: ec_trafficfilter

Use this data source to filter for an existing traffic filter that has been created via one of the provided filters. 

## Example Usage

```hcl
data "ec_trafficfilter" "name" {
  name = "example-filter"
}

data "ec_trafficfilter" "id" {
  id = "41d275439f884ce89359039e53eac516"
}

data "ec_trafficfilter" "region" {
  region = "us-east-1"
}
```

## Argument Reference

* `name` (Optional) - The name of the traffic filter. Has to match exactly.
* `id` (Optional) - The id of the traffic filter. Has to match exactly.
* `region` (Optional) - Region where the traffic filter is. For Elastic Cloud Enterprise (ECE) installations, use `"ece-region`.

## Attributes Reference
See the [API guide](https://www.elastic.co/guide/en/cloud/current/definitions.html#TrafficFilterRulesets) for the available fields.
