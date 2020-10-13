resource "null_resource" "bootstrap-elasticsearch" {
  provisioner "local-exec" {
    command = data.template_file.elasticsearch-configuration.rendered
  }
}

data "template_file" "elasticsearch-configuration" {
  template   = file("es_config.sh")
  depends_on = [ec_deployment.example_minimal]
  vars = {
    # Created servers and appropriate AZs
    elastic-user     = ec_deployment.example_minimal.elasticsearch_username
    elastic-password = ec_deployment.example_minimal.elasticsearch_password
    es-url           = ec_deployment.example_minimal.elasticsearch[0].https_endpoint
  }
}