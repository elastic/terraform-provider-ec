########################################
## Activity Logs
########################################

# data "azurerm_monitor_diagnostic_categories" "subscription" {
#   resource_id = data.azurerm_subscription.current.id
# }

#[ "AuditEvent", "Administrative", "Security",  "ServiceHealth", "Alert", "Recommendation", "Policy", "Autoscale", "ResourceHealth"]
resource "azurerm_monitor_diagnostic_setting" "elastic" {
  name                              = "AllLogsToElastic"
  target_resource_id                = data.azurerm_subscription.current.id
  eventhub_name                     = azurerm_eventhub.elastic.name
  eventhub_authorization_rule_id    = "${azurerm_eventhub_namespace.elastic.id}/authorizationRules/RootManageSharedAccessKey"
  
  dynamic "log" {
    for_each = [ "Administrative", "Security",  "ServiceHealth", "Alert", "Recommendation", "Policy", "Autoscale", "ResourceHealth"]
    content {
        category = log.value
        enabled  = true

        retention_policy {
            enabled = false
        }
    }
  }
}