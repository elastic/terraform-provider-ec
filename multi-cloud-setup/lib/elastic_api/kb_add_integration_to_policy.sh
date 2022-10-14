#!/bin/bash

eval "$(jq -r '@sh "ELASTIC_HTTP_METHOD=\(.elastic_http_method) ELASTIC_ENDPOINT=\(.kibana_endpoint) ELASTIC_USERNAME=\(.elastic_username) ELASTIC_PASSWORD=\(.elastic_password) ELASTIC_JSON_BODY=\(.elastic_json_body)"')"

# Define mapping
output=$(curl -s -X POST -u "$ELASTIC_USERNAME:$ELASTIC_PASSWORD" \
	-H "kbn-xsrf: true" -H 'Content-Type:application/json' -d "$ELASTIC_JSON_BODY" \
   ${ELASTIC_ENDPOINT}/api/fleet/package_policies | jq '.')

# Return response
ID=$( echo $output | jq -r '.item.id' )
SUCCESS=$( echo $output | jq -r '.success' )
ERROR=$( echo $output | jq -r '.error' )
MESSAGE=$( echo $output | jq -r '.message' )

jq -n --arg id "$ID" --arg success "$SUCCESS" --arg error "$ERROR" --arg message "$MESSAGE" '{"id": $id, "success" : $success, "error": $error, "message": $message}'