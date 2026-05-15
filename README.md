# notes-service


*Metrics monitoring Guide*
Build the app via 'docker compose up --build'
In a separate terminal, run ./monitoring-test/test.go. This file generates HTTP queries to different endpoints.
To view statistics on the metrics, Grafana app at http://localhost:3000. Login as admin (password="admin") and go to the Dashboards folder to view visualised data.