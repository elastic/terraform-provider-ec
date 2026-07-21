# -------------------------------------------------------------
# Create Audit Log Route to Elastic Agent
# -------------------------------------------------------------


resource "google_pubsub_topic" "audit" {
  name = var.google_pubsub_audit_topic

  labels = {
    elastic-log = "audit"
  }
}

resource "google_logging_project_sink" "audit" {
  name = var.google_pubsub_audit_topic

  # Can export to pubsub, cloud storage, or bigquery
  destination = "pubsub.googleapis.com/${google_pubsub_topic.audit.id}"

  filter = var.google_pubsub_audit_filter

  # Use a unique writer (creates a unique service account used for writing)
  unique_writer_identity = true

  depends_on = [google_pubsub_topic.audit]
}

# -------------------------------------------------------------
# Create Firewall Log Route to Elastic Agent
# -------------------------------------------------------------


resource "google_pubsub_topic" "firewall" {
  name = var.google_pubsub_firewall_topic

  labels = {
    elastic-log = "firewall"
  }
}

resource "google_logging_project_sink" "firewall" {
  name = var.google_pubsub_firewall_topic

  # Can export to pubsub, cloud storage, or bigquery
  destination = "pubsub.googleapis.com/${google_pubsub_topic.firewall.id}"

  filter = var.google_pubsub_firewall_filter

  # Use a unique writer (creates a unique service account used for writing)
  unique_writer_identity = true

  depends_on = [google_pubsub_topic.firewall]
}


# -------------------------------------------------------------
# Create VPC Flow Log Route to Elastic Agent
# -------------------------------------------------------------


resource "google_pubsub_topic" "vpcflow" {
  name = var.google_pubsub_vpcflow_topic

  labels = {
    elastic-log = "vpcflow"
  }
}

resource "google_logging_project_sink" "vpcflow" {
  name = var.google_pubsub_vpcflow_topic

  # Can export to pubsub, cloud storage, or bigquery
  destination = "pubsub.googleapis.com/${google_pubsub_topic.vpcflow.id}"

  filter = var.google_pubsub_vpcflow_filter

  # Use a unique writer (creates a unique service account used for writing)
  unique_writer_identity = true

  depends_on = [google_pubsub_topic.vpcflow]
}



# -------------------------------------------------------------
# Create DNS Log Route to Elastic Agent
# -------------------------------------------------------------


resource "google_pubsub_topic" "dns" {
  name = var.google_pubsub_dns_topic

  labels = {
    elastic-log = "dns"
  }
}

resource "google_logging_project_sink" "dns" {
  name = var.google_pubsub_dns_topic

  # Can export to pubsub, cloud storage, or bigquery
  destination = "pubsub.googleapis.com/${google_pubsub_topic.dns.id}"

  filter = var.google_pubsub_dns_filter

  # Use a unique writer (creates a unique service account used for writing)
  unique_writer_identity = true

  depends_on = [google_pubsub_topic.dns]
}


# -------------------------------------------------------------
# Create Loadbalancer Log Route to Elastic Agent
# -------------------------------------------------------------


resource "google_pubsub_topic" "lb" {
  name = var.google_pubsub_lb_topic

  labels = {
    elastic-log = "loadbalancer"
  }
}

resource "google_logging_project_sink" "lb" {
  name = var.google_pubsub_lb_topic

  # Can export to pubsub, cloud storage, or bigquery
  destination = "pubsub.googleapis.com/${google_pubsub_topic.lb.id}"

  filter = var.google_pubsub_lb_filter

  # Use a unique writer (creates a unique service account used for writing)
  unique_writer_identity = true

  depends_on = [google_pubsub_topic.lb]
}


# -------------------------------------------------------------
# Role bindings
# -------------------------------------------------------------

resource "google_project_iam_binding" "pubsub_writer_logs" {
  project            = var.google_cloud_project
  role               = "roles/pubsub.editor"

  members = [
    google_logging_project_sink.audit.writer_identity,
    google_logging_project_sink.firewall.writer_identity,
    google_logging_project_sink.vpcflow.writer_identity,
    google_logging_project_sink.dns.writer_identity,
    google_logging_project_sink.lb.writer_identity,
  ]

  depends_on = [
    google_logging_project_sink.audit,
    google_logging_project_sink.firewall,
    google_logging_project_sink.vpcflow,
    google_logging_project_sink.dns,
    google_logging_project_sink.lb
  ]
}