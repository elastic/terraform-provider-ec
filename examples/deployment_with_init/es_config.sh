#!/bin/bash

log (){
  echo "\n######## $1 ########\n"
}

create_users() {
  # Create a personal superuser
  log "creating additional superuser"
  curl -s --user ${elastic-user}:${elastic-password} -XPOST ${es-url}/_security/user/omer -d\
   '{"password" : "Password1","roles" : [ "superuser"],"full_name" : "Omer Kushmaro","email" : "omer.kushmaro@elastic.co"}'\
   -H 'Accept: application/json' -H 'Content-type: application/json'

  # Create an ingest user
  log "creating an ingest user"
  curl -s --user ${elastic-user}:${elastic-password} -XPOST ${es-url}/_security/user/filebeat -d\
  '{"password" : "Ingest123!","roles" : [ "ingest_admin"],"full_name" : "File Beat","email" : "application@elastic.co"}'\
  -H 'Accept: application/json' -H 'Content-type: application/json'
}

create_indices() {
  # Creating my importnat index and mapping
  log "creating pre-defined index and mapping"
  curl -s --user ${elastic-user}:${elastic-password} -XPUT ${es-url}/my-index-000001 -d\
   '{"settings": {"number_of_shards": 2,"number_of_replicas": 2},"mappings": {"properties": {"field1": { "type": "text" }}}}'\
   -H 'Accept: application/json' -H 'Content-type: application/json'
}

_main() {
 create_indices
 create_users
}

_main "$@"