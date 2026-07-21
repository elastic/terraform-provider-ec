resource "azurerm_resource_group" "main" {
  name     = "tf-elastic-group"
  location = var.azure_region
}

data "azurerm_client_config" "current" {}

data "azurerm_subscription" "current" {}

#################
## Eventhub to collect the Logs
#################

resource "azurerm_eventhub_namespace" "elastic" {
  name                = "azureLogsToElastic"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "Standard"
  capacity            = 1

}

resource "azurerm_eventhub" "elastic" {
  name                = "azureLogsToElasticHub"
  namespace_name      = azurerm_eventhub_namespace.elastic.name
  resource_group_name = azurerm_resource_group.main.name
  partition_count     = 2
  message_retention   = 1
}

resource "azurerm_eventhub_authorization_rule" "elastic" {
  name                = "azureLogsToElasticHubRule"
  namespace_name      = azurerm_eventhub_namespace.elastic.name
  eventhub_name       = azurerm_eventhub.elastic.name
  resource_group_name = azurerm_resource_group.main.name
  listen              = true
  send                = true
  manage              = false
}

resource "azurerm_storage_account" "elastic" {
  name                     = "azurelogs2elastic"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = "GRS"
}

#################
## Create a vault + key for the Agent VM
#################

# resource "azurerm_key_vault" "elastic" {
#   name                        = "elastic-key-vault-tf"
#   location                    = azurerm_resource_group.main.location
#   resource_group_name         = azurerm_resource_group.main.name
#   enabled_for_disk_encryption = true
#   tenant_id                   = data.azurerm_client_config.current.tenant_id
#   soft_delete_retention_days  = 7
#   purge_protection_enabled    = false

#   sku_name = "standard"

#   access_policy {
#     tenant_id = data.azurerm_client_config.current.tenant_id
#     object_id = data.azurerm_client_config.current.object_id

#     key_permissions = [
#       "Create",
#       "Get",
#       "Purge",
#       "Recover"
#     ]

#     secret_permissions = [
#       "Get",
#     ]

#     storage_permissions = [
#       "Get",
#     ]
#   }
# }

# resource "azurerm_key_vault_key" "generated" {
#   name         = "elastic-agent-certificate"
#   key_vault_id = azurerm_key_vault.elastic.id
#   key_type     = "RSA"
#   key_size     = 2048

#   key_opts = [
#     "decrypt",
#     "encrypt",
#     "sign",
#     "unwrapKey",
#     "verify",
#     "wrapKey",
#   ]
# }