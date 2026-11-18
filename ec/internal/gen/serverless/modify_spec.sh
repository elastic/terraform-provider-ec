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

# Audit logging is nested under monitoring. The generator marks the whole
# chain computed_optional because the fields exist on read. Force the blocks
# the practitioner sets to optional so an omitted block is a removal, and
# keep enabled computed_optional: a default on a non-computed attribute fails
# schema validation. destination stays required. destination.project_type is
# required and limited to the types project-api accepts. status stays computed.
# ignore_filters is an unordered array of {id}, so replace it with a set of
# ids and keep the size limit the list validator had.
jq '
(.resources[] | select(.name | endswith("_project")) | .schema.attributes[] | select(.name=="monitoring") | .single_nested) |= (
  .computed_optional_required = "optional"
  | (.attributes[] | select(.name=="logging") | .single_nested) |= (
      .computed_optional_required = "optional"
      | (.attributes[] | select(.name=="audit") | .single_nested) |= (
          .computed_optional_required = "optional"
          | (.attributes[] | select(.name=="enabled") | .bool) |= (
              .computed_optional_required = "computed_optional"
              | .default = {"static": true}
            )
          | (.attributes[] | select(.name=="destination") | .single_nested.attributes[] | select(.name=="project_type") | .string) |= (
              .computed_optional_required = "required"
              | .validators = [{
                  "custom": {
                    "imports": [{"path": "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"}],
                    "schema_definition": "stringvalidator.OneOf(\n\"observability\",\n\"security\",\n)"
                  }
                }]
            )
          | .attributes = (
              [.attributes[] | select(.name != "ignore_filters")]
              + [{
                  "name": "ignore_filter_ids",
                  "set": {
                    "computed_optional_required": "optional",
                    "element_type": {"string": {}},
                    "description": "IDs of logging ignore filters applied to reduce the volume of logs delivered for this category. The predefined IDs are listed in the [Elastic Cloud Serverless API reference](https://www.elastic.co/docs/api/doc/elastic-cloud-serverless/operation/operation-listignorefilters).",
                    "validators": [{
                      "custom": {
                        "imports": [{"path": "github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"}],
                        "schema_definition": "setvalidator.SizeAtMost(10)"
                      }
                    }]
                  }
                }]
            )
        )
    )
)
' /tmp/with-linked.json >/tmp/with-audit.json

# search_lake.min_search_vcu is the Minimum Search VCU floor on Elasticsearch
# projects. The generator marks it computed_optional because it is present on
# read, but a computed attribute could never be unset from configuration: a
# null config value would be planned from prior state. Make it optional-only so
# removing the line plans null, which Patch sends as JSON null. Validate only
# the lower bound: the maximum can differ between organizations and change
# over time, so the project API enforces it and limit changes need no provider
# release.
#
# metadata.annotations arrived with the same spec bump and is not exposed yet;
# drop it from the three project schemas so Read does not have to fill an
# attribute nothing maps (NewMetadataValue fails on a missing attribute). The
# same bump made metadata.tags optional in the API (it used to be the only
# metadata key); keep the Terraform attribute required and the block
# description as they were so the metadata schema does not change.
jq '
(.resources[] | select(.name=="elasticsearch_project") | .schema.attributes[] | select(.name=="search_lake") | .single_nested.attributes[] | select(.name=="min_search_vcu") | .int64) |= (
  .computed_optional_required = "optional"
  | .validators = [{
      "custom": {
        "imports": [{"path": "github.com/hashicorp/terraform-plugin-framework-validators/int64validator"}],
        "schema_definition": "int64validator.AtLeast(2)"
      }
    }]
  | .description += " Remove the attribute from a configured `search_lake` block, or set it to `null`, to unset the minimum. See [Minimum Search VCUs](https://www.elastic.co/docs/deploy-manage/deploy/elastic-cloud/project-settings#minimum-search-vcus) for how the setting affects scaling and billing."
)
| (.resources[] | select(.name | endswith("_project")) | .schema.attributes[] | select(.name=="metadata") | .single_nested) |= (
    .description = "Metadata request for a project with tags."
    | .attributes = [.attributes[] | select(.name != "annotations")]
    | (.attributes[] | select(.name=="tags") | .map.computed_optional_required) = "required"
  )
' /tmp/with-audit.json >./spec-mod.json

# jq's |= on a select() that matches nothing exits 0 and leaves the input
# untouched, so assert the rewrites above landed rather than letting a moved
# or renamed field silently regenerate the default schema.
jq -e '
([.resources[] | select(.name=="elasticsearch_project") | .schema.attributes[] | select(.name=="search_lake") | .single_nested.attributes[] | select(.name=="min_search_vcu") | .int64]
  | length == 1 and .[0].computed_optional_required == "optional" and (.[0].validators | length == 1))
and ([.resources[] | select(.name | endswith("_project")) | .schema.attributes[] | select(.name=="metadata") | .single_nested.attributes[] | select(.name=="annotations")] | length == 0)
and ([.resources[] | select(.name | endswith("_project")) | .schema.attributes[] | select(.name=="metadata") | .single_nested.attributes[] | select(.name=="tags") | .map.computed_optional_required] | all(. == "required"))
' ./spec-mod.json >/dev/null || { echo "modify_spec.sh: min_search_vcu or metadata rewrite did not apply; check the paths in this script against spec.json" >&2; exit 1; }
