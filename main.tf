provider "aws" {
  access_key                  = "test"
  secret_key                  = "test"
  region                      = "us-east-1"
  s3_use_path_style           = false
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    apigateway     = "http://localhost:4566"
    apigatewayv2   = "http://localhost:4566"
    cloudformation = "http://localhost:4566"
    cloudwatch     = "http://localhost:4566"
    dynamodb       = "http://localhost:4566"
    ec2            = "http://localhost:4566"
    es             = "http://localhost:4566"
    elasticache    = "http://localhost:4566"
    firehose       = "http://localhost:4566"
    iam            = "http://localhost:4566"
    kinesis        = "http://localhost:4566"
    lambda         = "http://localhost:4566"
    rds            = "http://localhost:4566"
    redshift       = "http://localhost:4566"
    route53        = "http://localhost:4566"
    s3             = "http://s3.localhost.localstack.cloud:4566"
    secretsmanager = "http://localhost:4566"
    ses            = "http://localhost:4566"
    sns            = "http://localhost:4566"
    sqs            = "http://localhost:4566"
    ssm            = "http://localhost:4566"
    stepfunctions  = "http://localhost:4566"
    sts            = "http://localhost:4566"
  }
}

data "aws_vpc" "my_vpc" {
    default = true
}

resource "aws_subnet" "public_subnet_1" {
  vpc_id                  = data.aws_vpc.my_vpc.id
  cidr_block              = "172.31.1.0/24"  
  availability_zone       = "us-east-1a"  
}

resource "aws_subnet" "public_subnet_2" {
  vpc_id                  = data.aws_vpc.my_vpc.id
  cidr_block              = "172.31.2.0/24"  
  availability_zone       = "us-east-1b"  
}

resource "aws_subnet" "private_subnet_1" {
  vpc_id                  = data.aws_vpc.my_vpc.id
  cidr_block              = "172.31.3.0/24"  
  availability_zone       = "us-east-1a"  
}

resource "aws_subnet" "private_subnet_2" {
  vpc_id                  = data.aws_vpc.my_vpc.id
  cidr_block              = "172.31.4.0/24"  
  availability_zone       = "us-east-1b"  
}


resource "aws_instance" "public_instances_1" {
  count         = 2
  ami           = "ami-12345678"  
  instance_type = ["m4.large", "c5.large"][count.index]  

  subnet_id = aws_subnet.public_subnet_1.id
}

resource "aws_instance" "public_instances_2" {
  count         = 2
  ami           = "ami-12345678"  
  instance_type = ["m4.large", "c5.large"][count.index]  

  subnet_id = aws_subnet.public_subnet_2.id
}

resource "aws_instance" "private_subnet_1_instances" {
  count         = 2
  ami           = "ami-12345678"  
  instance_type = ["m4.large", "c5.large"][count.index]  

  subnet_id = aws_subnet.private_subnet_1.id


}

resource "aws_instance" "private_subnet_2_instances" {
  count         = 2
  ami           = "ami-12345678"  
  instance_type = ["m4.large", "c5.large"][count.index]  

  subnet_id = aws_subnet.private_subnet_2.id
}