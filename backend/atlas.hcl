env "dev" {
  url = "postgres://postgres:postgres@localhost:5432/chat?sslmode=disable"
  migration {
    dir = "file://migrations"
  }
  schema {
    src = ["file://schema"]
  }
}

env "docker" {
  url = "postgres://postgres:postgres@db:5432/chat?sslmode=disable"
  migration {
    dir = "file://migrations"
  }
  schema {
    src = ["file://schema"]
  }
}


