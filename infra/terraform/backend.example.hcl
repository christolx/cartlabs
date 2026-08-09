bucket = "cartlabs-tf-state"
key    = "demo/terraform.tfstate"
region = "auto"

endpoints = {
  s3 = "https://S3_ENDPOINT"
}

use_path_style              = true
use_lockfile                = true
skip_credentials_validation = true
skip_region_validation      = true
skip_requesting_account_id  = true
