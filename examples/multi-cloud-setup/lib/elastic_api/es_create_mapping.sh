#!/bin/bash

eval "$(jq -r '@sh "ELASTIC_HTTP_METHOD=\(.elastic_http_method) ELASTIC_ENDPOINT=\(.elastic_endpoint) ELASTIC_USERNAME=\(.elastic_username) ELASTIC_PASSWORD=\(.elastic_password) ELASTIC_JSON_BODY=\(.elastic_json_body) ELASTIC_INDEX_NAME=\(.elastic_index_name)"')"

# Create index
# output=$(curl -s -X ${ELASTIC_HTTP_METHOD} -u "$ELASTIC_USERNAME:$ELASTIC_PASSWORD" \
#   ${ELASTIC_ENDPOINT}/${ELASTIC_INDEX_NAME})

# Define mapping
output=$(curl -s -X ${ELASTIC_HTTP_METHOD} -u "$ELASTIC_USERNAME:$ELASTIC_PASSWORD" \
   -H 'Content-Type:application/json' -d "$ELASTIC_JSON_BODY" \
   ${ELASTIC_ENDPOINT}/${ELASTIC_INDEX_NAME} | jq '.')

# Return response
ACKNOWLEDGED=$( echo $output | jq -r '.acknowledged' )
jq -n --arg acknowledged "$ACKNOWLEDGED" '{"acknowledged" : $acknowledged}'
