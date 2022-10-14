resource "azurerm_resource_group" "main" {
  name     = "tf-elastic-group"
  location = var.azure_location
}