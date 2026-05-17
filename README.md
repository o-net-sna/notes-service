# Notes Service

A containerized Notes REST API project developed for the **System and Network Administration** course.  
The project demonstrates deployment and administration concepts including container orchestration, reverse proxying, monitoring, health checks and service integration.

Technologies used:

- Go
- PostgreSQL
- Docker
- Docker Compose
- NGINX
- Prometheus
- Grafana

---

## Architecture

The system consists of multiple services connected through a Docker network.

```
                        ┌──────────────────┐
                        │      Client      │
                        │ curl / browser   │
                        └────────┬─────────┘
                                 │
                                 ▼
                    ┌────────────────────────┐
                    │         NGINX          │
                    │ Reverse Proxy Layer    │
                    └────────┬───────────────┘
                             │
                             ▼
                 ┌───────────────────────────┐
                 │       Go Notes API        │
                 │ CRUD + health + metrics   │
                 └────────┬──────────────────┘
                          │
                          ▼
               ┌───────────────────────┐
               │      PostgreSQL       │
               │ Persistent storage    │
               └───────────────────────┘

                          │
                          │ /metrics
                          ▼

               ┌───────────────────────┐
               │     Prometheus        │
               │ metrics collection    │
               └────────┬──────────────┘
                        │
                        ▼
               ┌───────────────────────┐
               │       Grafana         │
               │ dashboards/monitoring │
               └───────────────────────┘
``` 
## Project Structure

notes-service/

├── docker-compose.yml \
├── Dockerfile \
├── .env.example \
│ \
├── nginx/ \
│   └── default.conf \
│ \
├── prometheus/ \
│   └── prometheus.yml \
│ \
├── monitoring-test/ \
│   └── test.go \
│ \
├── handlers/ \
├── repository/ \
├── models/ \
│\
└── main.go

### Main files

**docker-compose.yml:**

Defines all project services:
- PostgreSQL database
- Go backend
- NGINX reverse proxy
- Prometheus
- Grafana

Also configures:
- Docker networking
- Volumes
- Health checks
- Environment variables

**Dockerfile:**

Builds the backend service using a multi-stage build:

- Go build stage
- Lightweight Alpine runtime container

**nginx/default.conf:**

Configures NGINX reverse proxy:

- receives client requests
- forwards traffic to backend
- adds proxy headers
- manages request forwarding


**prometheus/prometheus.yml:**

Prometheus configuration:

- registers backend target
- periodically collects metrics
- scrapes /metrics

**monitoring-test/test.go:**

Simple load testing utility.

Generates requests to:

- /notes
- /healthz
- /whoami
- /metrics

Used to produce traffic for monitoring dashboards

**repository/**

Database layer

Contains:

- migrations
- note storage operations
 
**handlers/**

HTTP request handlers.

Contains API endpoints and request processing

**models/**

Data structures used in the application.

**main.go:**

Application entry point.

Responsible for:

- database initialization
- migrations
- route registration
- health checks
- metrics registration
- middleware setup

## Features

Implemented functionality:

- REST API using Go
- CRUD operations for notes
- PostgreSQL integration
- Docker containerization
- Docker Compose orchestration
- NGINX reverse proxy
- health checks
- Prometheus metrics
- Grafana dashboards
- monitoring middleware

## Running project
### Step 1
Clone repository:
```
git clone https://github.com/o-net-sna/notes-service.git
cd notes-service
```

### Step 2
Create environment file:
```
cp .env.example .env
```
Example:
```
POSTGRES_DB=notesdb
POSTGRES_USER=postgres
POSTGRES_PASSWORD=12345
BACKEND_REPLICAS=2
NGINX_PORT=8080
```

### Step 3
Run containers:
```
docker compose up --build
```
Check container status:
```
docker ps
```
Expected services:

- postgres
- backend
- nginx
- prometheus
- grafana

## Service URLs
### API
```
http://localhost:8080
```
### Prometheus
```
http://localhost:9090
```
### Grafana
```
http://localhost:3000
```

Login:
```
username: admin
password: admin
```

## API Endpoints

### Create note
```
POST /notes
```
Example:
```
curl.exe --% -X POST http://localhost:18080/notes -H "Content-Type: application/json" -d "{\"title\":\"Test\",\"content\":\"Hello world\"}"
```
### Get all notes
```
curl http://localhost:18080/notes
```
### Get note by id
```
curl http://localhost:18080/notes/1
```
### Delete note
```
curl -X DELETE http://localhost:18080/notes/1
```
### Health check
```
curl http://localhost:18080/healthz
```
Exected:
```
204 No Content
```

### Service identification
```
curl http://localhost:18080/whoami
```
Example:
```
{
    "hostname":"274e9c73f234",
    "service":"notes-service"
}
```

### Monitoring

Prometheus collects metrics from:
```
http://backend:8000/metrics
```

Metrics endpoint is exposed through API:
```
http://localhost:8080/metrics
```
Example:
```
curl http://localhost:8080/metrics
```
Available custom metrics:

- http_requests_total
- http_request_duration_seconds
- http_requests_in_flight

### Prometheus
Open:
```
http://localhost:9090
```

Useful checks:

Status -> Targets

Expected:
```
notes-service -> UP
```
This confirms successful metric collection

### Grafana
Open:
```
http://localhost:3000
```
Login:
```
admin/admin
```

Go to the Dashboards folder to view visualised data
Dashboard displays:

- total requests
- request rate
- request duration
- active requests

Dashboard preview:

<img width="700" src="https://github.com/user-attachments/assets/7e576c5a-9e3a-456d-b9b7-8d22c81d347c">


### Load Testing
To generate traffic for monitoring:

Run:
```
go run monitoring-test/test.go
```
The script:

- creates test notes
- performs concurrent requests
- collects endpoint statistics
- cleans up created resources

Use this together with Grafana to observe live changes

### Health Checks

All services use health checks:

- PostgreSQL readiness
- backend availability
- NGINX availability

This ensures services start in the correct order

## Project Goal
The purpose of this project is not only backend development but also demonstrating practical System Administration concepts:

- service orchestration
- reverse proxy configuration
- monitoring
- health checking
- container networking
- infrastructure deployment
