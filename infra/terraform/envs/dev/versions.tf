terraform {
  required_version = ">= 1.11"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 8.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.7"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }

  # bucket は -backend-config で渡す
  backend "gcs" {
    prefix = "envs/dev"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# API トークンは環境変数 CLOUDFLARE_API_TOKEN で渡す
provider "cloudflare" {}

# Wasabi を S3 互換 API で操作する。アクセスキーは環境変数 AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY で渡す
provider "aws" {
  region                      = "ap-northeast-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  skip_metadata_api_check     = true
  skip_region_validation      = true
  s3_use_path_style           = true

  endpoints {
    s3 = "https://s3.ap-northeast-1.wasabisys.com"
  }
}
