#!/bin/bash

eval "$(jq -r '@sh "ELASTIC_ENDPOINT=\(.elastic_endpoint) ELASTIC_USERNAME=\(.elastic_username) ELASTIC_PASSWORD=\(.elastic_password) API_KEY_BODY=\(.api_key_body)"')"

output=$(curl -s -X POST -u "$ELASTIC_USERNAME:$ELASTIC_PASSWORD" \
   -H 'Content-Type:application/json' -d "$API_KEY_BODY" \
   ${ELASTIC_ENDPOINT}/_security/api_key | jq '.')

ENCODED=$( echo $output | jq -r '.encoded' )
jq -n --arg encoded "$ENCODED" '{"encoded" : $encoded}'
