#!/bin/bash

eval "$(jq -r '@sh "ELASTIC_HTTP_METHOD=\(.elastic_http_method) ELASTIC_ENDPOINT=\(.kibana_endpoint) ELASTIC_USERNAME=\(.elastic_username) ELASTIC_PASSWORD=\(.elastic_password) SO_FILE=\(.so_file)"')"

# Define mapping
output=$(curl -s -X ${ELASTIC_HTTP_METHOD} -u "$ELASTIC_USERNAME:$ELASTIC_PASSWORD" \
	-H "kbn-xsrf: true" --form file=@${SO_FILE} \
   ${ELASTIC_ENDPOINT}/api/saved_objects/_import | jq '.')

# Return response
SUCCESS=$( echo $output | jq -r '.success' )
ERROR=$( echo $output | jq -r '.error' )
MESSAGE=$( echo $output | jq -r '.message' )

jq -n --arg success "$SUCCESS" --arg error "$ERROR" --arg message "$MESSAGE" '{"success" : $success, "error": $error, "message": $message}'