# aws provider の S3 エンドポイントを Wasabi に向けて管理する
resource "aws_s3_bucket" "this" {
  bucket = var.bucket_name
}

# 署名付き URL でブラウザから直接 PUT / GET する
resource "aws_s3_bucket_cors_configuration" "this" {
  bucket = aws_s3_bucket.this.id

  cors_rule {
    allowed_origins = var.cors_origins
    allowed_methods = ["GET", "HEAD", "PUT"]
    allowed_headers = ["*"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3600
  }
}
