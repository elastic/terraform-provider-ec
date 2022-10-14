#!/bin/bash

eval "$(jq -r '@sh "TRANSFORM_NAME=\(.transform_name) ELASTIC_ENDPOINT=\(.elastic_endpoint) ELASTIC_USERNAME=\(.elastic_username) ELASTIC_PASSWORD=\(.elastic_password) ELASTIC_JSON_BODY=\(.elastic_json_body)"')"

# Define mapping
output=$(curl -s -X PUT -u "$ELASTIC_USERNAME:$ELASTIC_PASSWORD" \
   -H 'Content-Type:application/json' -d "$ELASTIC_JSON_BODY" \
   ${ELASTIC_ENDPOINT}/_transform/${TRANSFORM_NAME} | jq '.')

# Return response
ACKNOWLEDGED=$( echo $output | jq -r '.acknowledged' )
jq -n --arg acknowledged "$ACKNOWLEDGED" '{"acknowledged" : $acknowledged}'


