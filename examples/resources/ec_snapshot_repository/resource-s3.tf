resource "ec_snapshot_repository" "this" {
  name = "my-snapshot-repository"
  s3 = {
    bucket     = "my-bucket"
    access_key = "my-access-key"
    secret_key = "my-secret-key"
  }
}
