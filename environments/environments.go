package environments

import "os"

var PostgresUrl = os.Getenv("POSTGRES_URL")
