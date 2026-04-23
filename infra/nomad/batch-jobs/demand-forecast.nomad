# Nomad — Demand Forecast Batch Job
# Responsibility: non-containerised and batch workloads that don't fit Kubernetes
# Nomad handles: ML training batch jobs, long-running data pipeline tasks,
# Java-based report generation (needs host access), legacy service wrappers

job "demand-forecast-daily" {
  datacenters = ["dc1"]
  type        = "batch"
  namespace   = "shopos-data"

  periodic {
    cron             = "0 2 * * *"   # 2 AM daily
    prohibit_overlap = true
    time_zone        = "UTC"
  }

  meta {
    owner   = "supply-chain-team"
    purpose = "Daily demand forecast model run for inventory planning"
    sla     = "must complete before 6 AM UTC"
  }

  group "forecast" {
    count = 1

    restart {
      attempts = 2
      interval = "30m"
      delay    = "5m"
      mode     = "fail"
    }

    ephemeral_disk {
      size    = 10000  # 10GB for model artifacts and training data
      sticky  = false
      migrate = false
    }

    task "fetch-training-data" {
      driver = "exec"

      config {
        command = "/usr/local/bin/python3"
        args = [
          "src/supply-chain/demand-forecast-service/scripts/fetch_training_data.py",
          "--start-date", "${NOMAD_META_forecast_start}",
          "--output-path", "${NOMAD_ALLOC_DIR}/data/training.parquet"
        ]
      }

      env {
        POSTGRES_HOST     = "${attr.unique.hostname}"
        CLICKHOUSE_HOST   = "clickhouse.shopos.internal"
        AWS_REGION        = "us-east-1"
        OUTPUT_BUCKET     = "shopos-ml-artifacts"
      }

      vault {
        policies  = ["shopos-data-read"]
        env       = true
        change_mode = "restart"
      }

      template {
        data = <<EOT
{{ with secret "database/creds/demand-forecast-ro" }}
POSTGRES_USER={{ .Data.username }}
POSTGRES_PASSWORD={{ .Data.password }}
{{ end }}
EOT
        destination = "secrets/db.env"
        env         = true
      }

      resources {
        cpu    = 2000
        memory = 4096
      }

      lifecycle {
        hook    = "prestart"
        sidecar = false
      }
    }

    task "train-and-forecast" {
      driver = "exec"

      config {
        command = "/usr/local/bin/python3"
        args = [
          "src/supply-chain/demand-forecast-service/main.py",
          "--training-data", "${NOMAD_ALLOC_DIR}/data/training.parquet",
          "--forecast-horizon", "30",
          "--model-output", "${NOMAD_ALLOC_DIR}/model/",
          "--forecast-output", "${NOMAD_ALLOC_DIR}/forecast.parquet"
        ]
      }

      env {
        MLFLOW_TRACKING_URI = "http://mlflow.shopos.internal:5000"
        EXPERIMENT_NAME     = "demand-forecast-prod"
        PYTHONPATH          = "/opt/shopos/src"
        OMP_NUM_THREADS     = "4"
      }

      resources {
        cpu    = 8000   # 8 cores for model training
        memory = 32768  # 32GB
      }

      artifact {
        source      = "s3::https://shopos-ml-artifacts.s3.amazonaws.com/demand-forecast/deps.tar.gz"
        destination = "local/deps"
      }
    }

    task "publish-forecast" {
      driver = "exec"

      config {
        command = "/usr/local/bin/python3"
        args = [
          "src/supply-chain/demand-forecast-service/scripts/publish_forecast.py",
          "--forecast-path", "${NOMAD_ALLOC_DIR}/forecast.parquet",
          "--target-table", "demand_forecasts"
        ]
      }

      resources {
        cpu    = 500
        memory = 1024
      }

      lifecycle {
        hook    = "poststop"
        sidecar = false
      }
    }
  }
}
