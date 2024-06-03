#!/bin/env bash

jq --slurpfile plan_modifiers string_use_state_for_unknown.json 'def applies_to($item): $plan_modifiers[0].applies_to[] | any(. == $item; .); (.resources[] | select(.name == "elasticsearch_project") | .schema.attributes[] | select(applies_to(.name))).string.plan_modifiers |= $plan_modifiers[0].add' spec.json > /tmp/with-strings.json

mv /tmp/with-strings.json ./spec-mod.json
