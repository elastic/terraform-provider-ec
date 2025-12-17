#!/bin/env bash

jq --slurpfile plan_modifiers string_use_state_for_unknown.json 'def applies_to($item): $plan_modifiers[0].applies_to[] | any(. == $item; .); (.resources[] | .schema.attributes[] | select(applies_to(.name))).string.plan_modifiers |= $plan_modifiers[0].add' spec.json > /tmp/with-strings.json

jq --slurpfile product_type_custom_type product_type_custom_type.json '(.resources[] | select(.name == "security_project") | .schema.attributes[] | select(.name == "product_types")).list_nested.custom_type |= $product_type_custom_type[0]' /tmp/with-strings.json > /tmp/with-custom-type.json

mv /tmp/with-custom-type.json ./spec-mod.json