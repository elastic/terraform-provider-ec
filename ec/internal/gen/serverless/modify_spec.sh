#!/bin/env bash

jq --slurpfile plan_modifiers string_use_state_for_unknown.json 'def applies_to($item): $plan_modifiers[0].applies_to[] | any(. == $item; .); (.resources[] | .schema.attributes[] | select(applies_to(.name))).string.plan_modifiers |= $plan_modifiers[0].add' spec.json > /tmp/with-strings.json

jq --slurpfile product_type_custom_type product_type_custom_type.json '(.resources[] | select(.name == "security_project") | .schema.attributes[] | select(.name == "product_types")).list_nested.custom_type |= $product_type_custom_type[0]' /tmp/with-strings.json > /tmp/with-custom-type.json

# Add traffic_filter_ids to all project resources and remove the generated traffic_filters attribute
# (only traffic_filter_ids is used in hand-written code; traffic_filters is unused API cruft)
jq '(.resources[] | select(.name | endswith("_project")) | .schema.attributes) += [{
  "name": "traffic_filter_ids",
  "set": {
    "computed_optional_required": "optional",
    "element_type": {
      "string": {}
    },
    "description": "Set of traffic filter IDs to associate with this project"
  }
}]
| (.resources[] | select(.name | endswith("_project")).schema.attributes) |= [.[] | select(.name != "traffic_filters")]' /tmp/with-custom-type.json > /tmp/with-traffic-filters.json

mv /tmp/with-traffic-filters.json ./spec-mod.json
