# -------------------------------------------------------------
# Create EC2 instance + Elastic Agent
# -------------------------------------------------------------

data "template_file" "install_agent" {
  template = file("../../lib/scripts/agent_install.sh")
  vars = {
    elastic_version = var.elastic_version
    elasticsearch_username = ec_deployment.elastic_deployment.elasticsearch_username
    elasticsearch_password = ec_deployment.elastic_deployment.elasticsearch_password
    kibana_endpoint = ec_deployment.elastic_deployment.kibana[0].https_endpoint
    integration_server_endpoint = ec_deployment.elastic_deployment.integrations_server[0].https_endpoint
    policy_id = data.external.elastic_create_policy.result.id
  }
}

resource "aws_security_group" "elastic-agent" {
  name        = "elastic-agent"
  description = "Allow traffic for elastic-agent"

  ingress {
    description      = "TLS from VPC"
    from_port        = 443
    to_port          = 443
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "SSH"
    from_port        = 22
    to_port          = 22
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "ICMP"
    from_port        = -1
    to_port          = -1
    protocol         = "icmp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "Other ports"
    from_port        = 8000
    to_port          = 9500
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "Other ports"
    from_port        = 6780
    to_port          = 6800
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  tags = {
    Name = "elastic-agent"
  }
}

resource "aws_instance" "elastic-agent" {
  ami           = local.ami # us-west-2
  instance_type = "t2.micro"
  associate_public_ip_address = true
  security_groups = [ aws_security_group.elastic-agent.name ]
  key_name = "felix-london"

  tags = {
    Name = "elastic-agent"
  }

  user_data = "${data.template_file.install_agent.rendered}"
}