resource "azurerm_linux_virtual_machine" "jenkins" {
  name                = "${var.name}-server"
  location            = azurerm_resource_group.jenkins.location
  resource_group_name = azurerm_resource_group.jenkins.name
  size                = var.vm_size
  admin_username      = var.admin_username

  network_interface_ids = [azurerm_network_interface.jenkins.id]

  admin_ssh_key {
    username   = var.admin_username
    public_key = file(var.ssh_pub_key_path)
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
    disk_size_gb         = var.disk_size_gb
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "ubuntu-24_04-lts"
    sku       = "server"
    version   = "latest"
  }

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.jenkins.id]
  }

  custom_data = base64encode(file("${path.root}/../../../../scripts/bash/jenkins-install.sh"))

  tags = {
    Name        = "${var.name}-server"
    Environment = var.environment
  }
}

resource "null_resource" "jenkins_setup" {
  depends_on = [azurerm_linux_virtual_machine.jenkins]

  connection {
    type        = "ssh"
    host        = azurerm_public_ip.jenkins.ip_address
    user        = var.admin_username
    private_key = file(var.private_key_path)
    timeout     = "10m"
  }

  provisioner "remote-exec" {
    inline = [
      "echo 'Waiting for Jenkins user_data script to complete...'",
      "until [ -f /var/lib/jenkins/jenkins-setup-complete ]; do echo 'Still setting up...'; sleep 30; done",
      "echo 'Jenkins setup complete!'",
      "curl -s -o /dev/null -w 'Jenkins HTTP status: %%{http_code}' -u admin:admin http://localhost:8080/api/json",
    ]
  }
}
