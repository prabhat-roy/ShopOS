# Bootstrapped via custom_data (cloud-init).

resource "azurerm_linux_virtual_machine" "this" {
  name                = "${var.name}-server"
  location            = var.location
  resource_group_name = var.resource_group_name
  size                = var.vm_size
  admin_username      = var.admin_username

  network_interface_ids = [azurerm_network_interface.this.id]

  admin_ssh_key {
    username   = var.admin_username
    public_key = var.ssh_pub_key
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
    disk_size_gb         = var.disk_size_gb
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "ubuntu-26_04-lts"
    sku       = "server"
    version   = "latest"
  }

  identity {
    type         = "UserAssigned"
    identity_ids = [var.identity_id]
  }

  custom_data = var.user_data == "" ? null : base64encode(var.user_data)

  tags = {
    Name        = "${var.name}-server"
    Environment = var.environment
  }
}
