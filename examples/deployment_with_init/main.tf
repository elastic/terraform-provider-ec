resource "null_resource" "bootstrap-elasticsearch" {
  provisioner "local-exec" {
    # Created servers and appropriate AZs
    command = templatefile("es_config.sh", {
      elastic-user     = ec_deployment.example_minimal.elasticsearch_username
      elastic-password = ec_deployment.example_minimal.elasticsearch_password
      es-url           = ec_deployment.example_minimal.elasticsearch[0].https_endpoint
    })
  }
}
