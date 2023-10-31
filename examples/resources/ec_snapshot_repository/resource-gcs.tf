resource "ec_snapshot_repository" "this" {
  name = "my-snapshot-repository"
  generic = {
    type = "gcs"
    settings = jsonencode({
      bucket   = "my_bucket"
      client   = "my_alternate_client"
      compress = false
    })
  }
}
