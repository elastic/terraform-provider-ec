#!/bin/bash

eval "$(jq -r '@sh "ELASTIC_HTTP_METHOD=\(.elastic_http_method) ELASTIC_ENDPOINT=\(.kibana_endpoint) ELASTIC_USERNAME=\(.elastic_username) ELASTIC_PASSWORD=\(.elastic_password) ELASTIC_JSON_BODY=\(.elastic_json_body)"')"

# Define mapping
output=$(curl -s -X POST -u "$ELASTIC_USERNAME:$ELASTIC_PASSWORD" \
	-H "kbn-xsrf: true" -H 'Content-Type:application/json' -d "$ELASTIC_JSON_BODY"  \
   ${ELASTIC_ENDPOINT}/api/detection_engine/rules/_bulk_action | jq '.')

# Return response
SUCCESS=$( echo $output | jq -r '.success' )

jq -n --arg success "$SUCCESS" '{"success" : $success}'