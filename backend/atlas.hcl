env "local" {
  url = "postgres://postgres:password@localhost:5432/chat?sslmode=disable"
  dev = "postgres://postgres:password@localhost:5432/chat_dev?sslmode=disable"
  migration {
    dir = "file://ent/migrate/migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "docker" {
  url = "postgres://postgres:password@db:5432/chat?sslmode=disable"
  dev = "postgres://postgres:password@db:5432/chat_dev?sslmode=disable"
  migration {
    dir = "file://ent/migrate/migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
