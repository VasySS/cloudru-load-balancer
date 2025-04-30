# load-balancer

[![License](https://img.shields.io/github/license/vasyss/cloudru-load-balancer)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vasyss/segoya-backend)](go.mod)

## Getting started

### Prerequisites

- **Docker** & **Docker Compose** installed
- Ensure the following ports are available:
  - **5432** (Postgres)
  - **8081** and **8082** for dummy backends (if needed)
  - **8080** for the load balancer

### Installation

1. **Clone the Repository**:

```sh
git clone https://github.com/VasySS/cloudru-load-balancer.git
cd cloudru-load-balancer
```

2. Create **.env** in root folder and set required fields (look at **.env.example** for reference)

3. Run the command to start the app and all services needed for it:

```sh
# with dummy backends for testing:
docker compose up -d --build
# only load balancer and postgres:
docker compose up -d --build balancer postgres
```

### Started services

| Service                             | URL                                             |
| ----------------------------------- | ----------------------------------------------- |
| Load balancer and rate limiter      | http://localhost:8080                           |
| Postgres                            | postgres://postgres:postgrespass@localhost:5432 |
| First dummy backend (if specified)  | http://localhost:8081                           |
| Second dummy backend (if specified) | http://localhost:8082                           |

## Example of running a load test

[![asciicast](https://asciinema.org/a/FQjXm3NFjJ61Nhb1pmsYw7FOr.svg)](https://asciinema.org/a/FQjXm3NFjJ61Nhb1pmsYw7FOr)
