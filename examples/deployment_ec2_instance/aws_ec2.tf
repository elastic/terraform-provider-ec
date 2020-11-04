resource "aws_instance" "inst" {
  ami                    = var.ubuntu18_ami
  instance_type          = "t2.large"
  vpc_security_group_ids = [aws_security_group.group.id]
  key_name               = var.keypair
  subnet_id              = aws_default_subnet.default.id
}

# Use the default VPC and subnet
resource "aws_default_vpc" "default" {}
resource "aws_default_subnet" "default" {
  # Using zone A for convinience (this would render as us-east-1a, in case us-east-1 is the region variable)
  availability_zone = format("%sa", var.region)
}

# Creating a security group to allow all outbound(egress) and ssh inbound (ingress)
resource "aws_security_group" "group" {
  vpc_id = aws_default_vpc.default.id
  egress {
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 22
    protocol    = "tcp"
    to_port     = 22
    cidr_blocks = ["0.0.0.0/0"]
  }
}

output "instance" {
  value = aws_instance.inst.public_ip
}
