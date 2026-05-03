output "instance_name" {
  value = google_compute_instance.this.name
}

output "public_ip" {
  value = google_compute_address.this.address
}
