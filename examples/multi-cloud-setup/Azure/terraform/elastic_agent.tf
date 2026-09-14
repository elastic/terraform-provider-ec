# -------------------------------------------------------------
# Create VM + Elastic Agent
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

resource "azurerm_virtual_machine_extension" "elastic-agent" {
  name                 = var.elastic_agent_vm_name
  virtual_machine_id = "${azurerm_linux_virtual_machine.agent.id}"
  publisher            = "Microsoft.Azure.Extensions"
  type                 = "CustomScript"
  type_handler_version = "2.0"

  settings = <<SETTINGS
    {"script": "${base64encode(data.template_file.install_agent.rendered)}"}
  SETTINGS
}

resource "azurerm_linux_virtual_machine" "agent" {
  depends_on = [ec_deployment.elastic_deployment, data.external.elastic_create_policy] ## We want to have the elastic deployment before we install the agent

  name                = var.elastic_agent_vm_name
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  size                = "Standard_F2"
  admin_username      = "adminuser"
  admin_password      = "${ec_deployment.elastic_deployment.elasticsearch_password}"
  disable_password_authentication = false
  network_interface_ids = [ 
    azurerm_network_interface.main.id,
    azurerm_network_interface.internal.id,
  ]

  # admin_ssh_key {
  #   username   = "adminuser"
  #   public_key = azurerm_key_vault_key.generated.public_key_pem
  # }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}

resource "azurerm_virtual_network" "main" {
  name                = "elastic-agent-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
}

resource "azurerm_subnet" "internal" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_public_ip" "pip" {
  name                = "elastic-agent-pip"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  allocation_method   = "Dynamic"
}

resource "azurerm_network_interface" "main" {
  name                = "elastic-agent-nic1"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurerm_subnet.internal.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.pip.id
  }
}

resource "azurerm_network_interface" "internal" {
  name                      = "elastic-agent-nic2"
  resource_group_name       = azurerm_resource_group.main.name
  location                  = azurerm_resource_group.main.location

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.internal.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_network_security_group" "agent" {
  name                = "elastic-agent-sg"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  security_rule {
    access                     = "Allow"
    direction                  = "Inbound"
    name                       = "tls"
    priority                   = 100
    protocol                   = "Tcp"
    source_port_range          = "*"
    source_address_prefix      = "*"
    destination_port_range     = "443"
    destination_address_prefix = azurerm_network_interface.main.private_ip_address
  }
}

resource "azurerm_network_interface_security_group_association" "main" {
  network_interface_id      = azurerm_network_interface.internal.id
  network_security_group_id = azurerm_network_security_group.agent.id
}