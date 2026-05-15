# notes-service


**Metrics monitoring Guide**

Build the app via 'docker compose up --build'
In a separate terminal, run ./monitoring-test/test.go. This file generates HTTP queries to different endpoints.
To view statistics on the metrics, Grafana app at http://localhost:3000. Login as admin (password="admin") and go to the Dashboards folder to view visualised data.

Dashboard preview:

<img width="700"  alt="Screenshot from 2026-05-15 04-17-49" src="https://github.com/user-attachments/assets/7e576c5a-9e3a-456d-b9b7-8d22c81d347c" />
