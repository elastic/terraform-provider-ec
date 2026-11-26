resource "ec_elasticsearch_project" "my_project" {
  name      = "my_project"
  region_id = "aws-us-east-1"

  search_lake = {
    search_power = 100
    boost_window = 7

    # Minimum number of search VCUs the Search tier keeps allocated, so it
    # never scales below this floor. At least 2; the maximum is enforced by
    # the API. Remove the attribute to unset the floor.
    min_search_vcu = 64
  }
}
