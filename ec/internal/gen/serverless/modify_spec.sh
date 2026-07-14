#!/bin/env bash

jq --slurpfile plan_modifiers string_use_state_for_unknown.json 'def applies_to($item): $plan_modifiers[0].applies_to[] | any(. == $item; .); (.resources[] | .schema.attributes[] | select(applies_to(.name))).string.plan_modifiers |= $plan_modifiers[0].add' spec.json >/tmp/with-strings.json

jq --slurpfile product_type_custom_type product_type_custom_type.json '(.resources[] | select(.name == "security_project") | .schema.attributes[] | select(.name == "product_types")).list_nested.custom_type |= $product_type_custom_type[0]' /tmp/with-strings.json >/tmp/with-custom-type.json

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
| (.resources[] | select(.name | endswith("_project")).schema.attributes) |= [.[] | select(.name != "traffic_filters")]' /tmp/with-custom-type.json >/tmp/with-traffic-filters.json

# Restructure the linked block: split the API-computed status out of the
# practitioner-controlled projects map element so that removing a linked project
# is detected as a change (a Computed attribute inside the element type would
# otherwise cause the framework to preserve removed elements). status moves to
# a top-level computed `statuses` map keyed by project ID. linked is also made
# Optional-only so that removing the whole block unlinks all projects.
#
# `statuses` is a top-level attribute (sibling of `linked`) rather than a child
# of `linked` because a SingleNestedAttribute containing any Computed child is
# treated as partially provider-owned, which causes the framework to preserve
# omitted `projects` map keys and prevents unlinking.
jq '
(.resources[] | select(.name | endswith("_project")) | .schema.attributes[] | select(.name=="linked") | .single_nested.computed_optional_required) = "optional"
| (.resources[] | select(.name | endswith("_project")) | .schema.attributes[] | select(.name=="linked") | .single_nested.attributes) |= [
    .[] | if .name == "projects" then
      (.map_nested.nested_object.attributes |= [.[] | select(.name != "status")])
    else . end
  ]
| (.resources[] | select(.name | endswith("_project")) | .schema.attributes[] | select(.name=="linked") | .single_nested.attributes) += [{
    "name": "statuses",
    "map": {
      "computed_optional_required": "computed",
      "element_type": { "string": {} },
      "description": "Status of each linked project, keyed by project ID. Populated by the provider from the API."
    }
  }]
' /tmp/with-traffic-filters.json >/tmp/with-linked.json

mv /tmp/with-linked.json ./spec-mod.json
