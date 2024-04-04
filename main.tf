terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}
variable "do_token" {
  type  = string
}

variable "ssh_private_key" {
  type  = string
}

variable "ssh_public_key" {
  type  = string
}

variable "mw_api_key" {
  type  = string
}

variable "mw_target" {
  type  = string
}


provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_ssh_key" "default" {
  name       = "terraform-key"
  public_key = var.ssh_public_key
}

resource "digitalocean_droplet" "docker" {
  name   = "terraform-droplet"
  region = "nyc3"
  size   = "s-1vcpu-1gb"
  image  = "ubuntu-20-04-x64"
  ssh_keys = [digitalocean_ssh_key.default.fingerprint]

  connection {
    host = self.ipv4_address
    user = "root"
    type = "ssh"
    private_key = var.ssh_private_key
    timeout = "100m"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo sleep 60",
      "sudo apt-get update",
      "sudo apt install bash",
      "sudo apt-get install -y docker.io docker-compose",
      "sudo MW_AGENT_DOCKER_IMAGE=ghcr.io/middleware-labs/terraform-agent MW_API_KEY=${var.mw_api_key} MW_TARGET=${var.mw_target} bash -c \"$(curl -L https://install.middleware.io/scripts/docker-install.sh)\"",
      "sudo sleep 40",
      "docker ps -a --filter ancestor=ghcr.io/middleware-labs/terraform-agent:latest"
    ]
  }

}
