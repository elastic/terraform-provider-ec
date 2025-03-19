#!/bin/bash

eval "$(jq -r '@sh "ELASTIC_HTTP_METHOD=\(.elastic_http_method) ELASTIC_ENDPOINT=\(.kibana_endpoint) ELASTIC_USERNAME=\(.elastic_username) ELASTIC_PASSWORD=\(.elastic_password) SO_FILE=\(.so_file)"')"

# Define mapping
output=$(curl -s -X PUT -u "$ELASTIC_USERNAME:$ELASTIC_PASSWORD" \
	-H "kbn-xsrf: true"  \
   ${ELASTIC_ENDPOINT}/api/detection_engine/rules/prepackaged | jq '.')

# Return response
RULES=$( echo $output | jq -r '.rules_installed' )

jq -n --arg rules "$RULES" '{"rules" : $rules}'