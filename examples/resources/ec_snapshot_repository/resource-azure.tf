resource "ec_snapshot_repository" "this" {
  name = "my-snapshot-repository"
  generic = {
    type = "azure"
    settings = jsonencode({
      container = "my_container"
      client    = "my_alternate_client"
      compress  = false
    })
  }
}
