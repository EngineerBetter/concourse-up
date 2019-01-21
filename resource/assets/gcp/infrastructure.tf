variable "zone" {
  type = "string"
	default = "{{ .Zone }}"
}
variable "tags" {
  type = "string"
	default = "{{ .Tags }}"
}
variable "project" {
  type = "string"
	default = "{{ .Project }}"
}
variable "gcpcredentialsjson" {
  type = "string"
	default = "{{ .GCPCredentialsJSON }}"
}
variable "externalip" {
  type = "string"
	default = "{{ .ExternalIP }}"
}

variable "deployment" {
  type = "string"
	default = "{{ .Deployment }}"
}
variable "region" {
  type = "string"
	default = "{{ .Region }}"
}

variable "db_tier" {
  type = "string"
	default = "{{ .DBTier }}"
}

variable "db_username" {
  type = "string"
	default = "{{ .DBUsername }}"
}
variable "db_password" {
  type = "string"
	default = "{{ .DBPassword }}"
}

variable "db_name" {
  type = "string"
  default = "{{ .DBName }}"
}

{{if .DNSManagedZoneName }}
variable "dns_managed_zone_name" {
  type = "string"
  default = "{{ .DNSManagedZoneName }}"
}

variable "dns_record_set_prefix" {
  type = "string"
  default = "{{ .DNSRecordSetPrefix }}"
}
{{end}}

variable "source_access_ip" {
  type = "string"
  default = "{{ .ExternalIP }}"
}

provider "google" {
    credentials = "{{ .GCPCredentialsJSON }}"
    project = "{{ .Project }}"
    region = "${var.region}"
}


terraform {
	backend "gcs" {
		bucket = "{{ .ConfigBucket }}"
		region = "{{ .Region }}"
	}
}

{{if .DNSManagedZoneName }}
data "google_dns_managed_zone" "dns_zone" {
  name = "${var.dns_managed_zone_name}"
}

resource "google_dns_record_set" "dns" {
  managed_zone = "${data.google_dns_managed_zone.dns_zone.name}"
  name = "${var.dns_record_set_prefix}.${data.google_dns_managed_zone.dns_zone.dns_name}"
  type    = "A"
  ttl     = 60

  rrdatas = ["${google_compute_address.atc_ip.address}"]
}
{{end}}

// route for nat
resource "google_compute_route" "nat" {
  name                   = "${var.deployment}-nat-route"
  dest_range             = "0.0.0.0/0"
  network                = "${google_compute_network.default.name}"
  next_hop_instance      = "${google_compute_instance.nat-instance.name}"
  next_hop_instance_zone = "${var.zone}"
  priority               = 800
  tags                   = ["no-ip"]
  project                = "${var.project}"
}

// nat
resource "google_compute_instance" "nat-instance" {
  name         = "${var.deployment}-nat-instance"
  machine_type = "n1-standard-1"
  zone         = "${var.zone}"
  project      = "${var.project}"

  tags = ["nat", "internal"]

  boot_disk {
    initialize_params {
      image = "ubuntu-1804-bionic-v20181222"
    }
  }

  network_interface {
    subnetwork = "${google_compute_subnetwork.private.name}"
    subnetwork_project = "${var.project}"
    access_config {
      // Ephemeral IP
    }
  }

  can_ip_forward = true

  metadata_startup_script = <<EOT
#!/bin/bash

netif="$(ip r | awk '/default/ {print $5}')"

echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
sudo sysctl -p

sudo iptables -t nat -A POSTROUTING -o "$netif" -j MASQUERADE
EOT
}

resource "google_compute_network" "default" {
  name                    = "${var.deployment}-bosh-network"
  project                 = "${var.project}"
  auto_create_subnetworks = "false"
}

resource "google_compute_subnetwork" "public" {
  name          = "${var.deployment}-bosh-${var.region}-subnet-public"
  ip_cidr_range = "10.0.0.0/24"
  network       = "${google_compute_network.default.self_link}"
  project       = "${var.project}"
}
resource "google_compute_subnetwork" "private" {
  name          = "${var.deployment}-bosh-${var.region}-subnet-private"
  ip_cidr_range = "10.0.1.0/24"
  network       = "${google_compute_network.default.self_link}"
  project       = "${var.project}"
}

resource "google_compute_firewall" "director" {
  name = "${var.deployment}-director"
  description = "Firewall for external access to BOSH director"
  network     = "${google_compute_network.default.self_link}"
  target_tags = ["external"]
  source_ranges = ["${var.source_access_ip}/32", "${google_compute_instance.nat-instance.network_interface.0.access_config.0.nat_ip}/32"]
  allow {
    protocol = "tcp"
    ports = ["6868", "25555", "22"]
  }
}

resource "google_compute_firewall" "nat" {
  name = "${var.deployment}-nat"
  description = "Firewall for external access to NAT"
  network     = "${google_compute_network.default.self_link}"
  target_tags = ["nat"]
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "tcp"
    ports = ["80", "443"]
  }
}

resource "google_compute_firewall" "atc-one" {
  name = "${var.deployment}-atc-one"
  description = "Firewall for external access to concourse atc"
  network     = "${google_compute_network.default.self_link}"
  target_tags = ["web"]
  source_tags = ["web", "worker", "external", "internal"]
  source_ranges = [{{ .AllowIPs }}]
  allow {
    protocol = "tcp"
    ports = ["80"]
  }
}

resource "google_compute_firewall" "atc-two" {
  name = "${var.deployment}-atc-two"
  description = "Firewall for external access to concourse atc"
  network     = "${google_compute_network.default.self_link}"
  target_tags = ["web"]
  source_ranges = ["${google_compute_instance.nat-instance.network_interface.0.access_config.0.nat_ip}/32", "${google_compute_address.atc_ip.address}/32", {{ .AllowIPs }}]
  allow {
    protocol = "tcp"
    ports = ["443", "8443"]
  }
}

resource "google_compute_firewall" "vms" {
  name = "${var.deployment}-vms"
  description = "Concourse UP VMs firewall"
  network     = "${google_compute_network.default.self_link}"
  target_tags = ["web", "external", "internal", "worker"]
  source_ranges = ["10.0.0.0/16"]
  allow {
    protocol = "tcp"
    ports = ["6868","4222", "25250", "25555", "25777","2222", "5555", "7777", "7788", "7799", "22"]
  }
  allow {
    protocol = "udp"
    ports = ["53"]
  }
  allow {
    protocol = "icmp"
  }
}

resource "google_compute_firewall" "atc-three" {
  name = "${var.deployment}-atc-three"
  description = "Firewall for external access to concourse atc"
  network     = "${google_compute_network.default.self_link}"
  target_tags = ["web"]
  source_ranges = ["${google_compute_instance.nat-instance.network_interface.0.access_config.0.nat_ip}/32", "${google_compute_address.atc_ip.address}/32", {{ .AllowIPs }}]
  allow {
    protocol = "tcp"
    ports = ["3000", "8844"]
  }
}

resource "google_compute_firewall" "internal" {
  name        = "${var.deployment}-int"
  description = "BOSH CI Internal Traffic"
  network     = "${google_compute_network.default.self_link}"
  source_tags = ["internal"]
  target_tags = ["internal"]

  allow {
    protocol = "tcp"
  }

  allow {
    protocol = "udp"
  }

  allow {
    protocol = "icmp"
  }
}

resource "google_compute_firewall" "sql" {
  name        = "${var.deployment}-sql"
  description = "BOSH CI External Traffic"
  network     = "${google_compute_network.default.self_link}"
  direction = "EGRESS"
  allow {
    protocol = "tcp"
    ports    = ["5432"]
  }
  destination_ranges = ["${google_sql_database_instance.director.first_ip_address}/32"]
}

resource "google_service_account" "bosh" {
  account_id   = "${var.deployment}-bosh"
  display_name = "bosh"
}
resource "google_service_account_key" "bosh" {
  service_account_id = "${google_service_account.bosh.name}"
  public_key_type = "TYPE_X509_PEM_FILE"
}

resource "google_project_iam_member" "bosh" {
  project = "${var.project}"
  role    = "roles/owner"
  member  = "serviceAccount:${google_service_account.bosh.email}"
}
resource "google_compute_address" "atc_ip" {
  name = "${var.deployment}-atc-ip"
}

resource "google_compute_address" "director" {
  name = "${var.deployment}-director-ip"
}

resource "google_sql_database_instance" "director" {
  name = "${var.db_name}"
  database_version = "POSTGRES_9_6"
  region       = "${var.region}"

  settings {
    tier = "${var.db_tier}"
    user_labels {
      deployment = "${var.deployment}"
    }

    ip_configuration {
      authorized_networks = {
        name = "atc_conf"
        value = "${google_compute_address.atc_ip.address}/32"}

    authorized_networks = {
      name = "bosh"
      value = "${google_compute_address.director.address}/32"
    }
    }
  }
}

resource "google_sql_database" "director" {
  name      = "udb"
  instance  = "${google_sql_database_instance.director.name}"
}

resource "google_sql_user" "director" {
  name     = "${var.db_username}"
  instance = "${google_sql_database_instance.director.name}"
  host     = "*"
  password = "${var.db_password}"
}
output "network" {
value = "${google_compute_network.default.name}"
}
output "director_firewall_name" {
value = "${google_compute_firewall.director.name}"
}

output "private_subnetwork_name" {
value = "${google_compute_subnetwork.private.name}"
}
output "public_subnetwork_name" {
value = "${google_compute_subnetwork.public.name}"
}

output "public_subnetwork_cidr" {
value = "${google_compute_subnetwork.public.ip_cidr_range}"
}
output "private_subnetwork_cidr" {
value = "${google_compute_subnetwork.private.ip_cidr_range}"
}

output "private_subnetwor_internal_gw" {
value = "${google_compute_subnetwork.private.gateway_address}"
}
output "public_subnetwor_internal_gw" {
value = "${google_compute_subnetwork.public.gateway_address}"
}

output "atc_public_ip" {
value = "${google_compute_address.atc_ip.address}"
}

output "director_account_creds" {
  value = "${base64decode(google_service_account_key.bosh.private_key)}"
}

output "director_public_ip" {
  value = "${google_compute_address.director.address}"
}

output "bosh_db_address" {
  value = "${google_sql_database_instance.director.first_ip_address}"
}

output "db_name" {
  value = "${google_sql_database_instance.director.name}"
}

output "nat_gateway_ip" {
  value = "${google_compute_instance.nat-instance.network_interface.0.access_config.0.nat_ip}"
}
output "server_ca_cert" {
  value = "${google_sql_database_instance.director.server_ca_cert.0.cert}"
}
