locals {
  map = {
    af-south-1: "ami-022666956ad401a1"
    ap-northeast-1: "ami-015f1a68ce825a8d2"
    ap-northeast-2: "ami-0be9734c9e68b99f4"
    ap-northeast-3: "ami-01cb3e73f8ef13fdc"
    ap-south-1: "ami-00aaac1f2ef4ce965"
    ap-southeast-1: "ami-0012ffabeb7413479"
    ap-southeast-2: "ami-03ec1fe05b3849c74"
    ca-central-1: "ami-04c56d394d31cdeac"
    eu-central-1: "ami-0980c5102b5ef10cc"
    me-south-1: "ami-03cc0b5db8321f2e5"
    ap-east-1: "ami-0c7e5903bee96ef81"
    eu-north-1: "ami-0663a4867a210287a"
    eu-south-1: "ami-035e213233577516f"
    eu-west-1: "ami-0213344887e47003a"
    eu-west-2: "ami-0add0a5a0cf9afc6c"
    eu-west-3: "ami-01019e7343a5f361d"
    sa-east-1: "ami-0312c74c38dc7bae6"
    us-east-1: "ami-0db6c6238a40c0681"
    us-east-2: "ami-03b6c8bd55e00d5ed"
    us-west-1: "ami-0f5868930cb63c89c"
    us-west-2: "ami-038a0ccaaedae6406"
  }

  ami = lookup(local.map, var.aws_region)
}