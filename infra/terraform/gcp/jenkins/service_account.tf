resource "google_service_account" "jenkins" {
  account_id   = "${var.name}-sa"
  display_name = "Jenkins Service Account"
}

resource "google_project_iam_member" "jenkins_owner" {
  project = var.project_id
  role    = "roles/owner"
  member  = "serviceAccount:${google_service_account.jenkins.email}"
}
