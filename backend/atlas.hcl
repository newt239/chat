env "dev" {
  url = env("DATABASE_URL")
  migration {
    dir = "file://migrations"
  }
  schema {
    src = ["file://schema"]
  }
}


