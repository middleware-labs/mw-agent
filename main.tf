terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

resource "digitalocean_droplet" "terraform" {
  name   = "terraform-droplet"
  region = "nyc3"
  size   = "s-1vcpu-1gb"
  image  = "ubuntu-23-04-x64"
}

provisioner "remote-exec" {
  inline = [
    "sleep 60",
    'sudo apt-get update',
    'MW_AGENT_DOCKER_IMAGE=ghcr.io/middleware-labs/terraform-agent MW_API_KEY=e3g4p2dkmoee09mmlwql64q2oornpyupht4c MW_TARGET=https://gh5k31l.middleware.io:443 bash -c "$(curl -L https://install.middleware.io/scripts/docker-install.sh)"',
  ]
}
