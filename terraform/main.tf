/* Backend */

terraform {
  backend "s3" {
    bucket = "remotebackend"
    key     = "differencebetween/terraform.tfstate"
    region  = "us-west-1"
    profile = "jds"
  }
}

/* Providers */

provider "aws" {
  region = var.region
  profile = "jds"
}
