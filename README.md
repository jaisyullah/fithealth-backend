# FitHealth Backend - Starter

## Prasyarat
- Docker & Docker Compose

## Setup & run (lokal)
1. Salin `.env.example` ke `.env` dan sesuaikan.
2. Jalankan:
   docker compose up --build

Server akan tersedia pada http://localhost:8080

## Endpoint
- POST /v1/observations
  - Body (contoh):
  {
    "deviceId": "watch-001",
    "patientId": "local-patient-123",
    "observations": [
      {"type":"heart_rate","value":"72","unit":"beats/min","timestamp":"2025-11-09T05:00:00+07:00"}
    ]
  }

- GET /health

## Notes
- Timestamp akan disimpan dalam UTC.
- Worker mengambil job dari Redis list `obs_queue`, mapping ke FHIR Observation, lalu POST ke SATUSEHAT_FHIR_URL + /Observation menggunakan token dari SATUSEHAT_TOKEN_URL.
- Perlu menyesuaikan: error handling lebih robust, retries dengan backoff, tracing, metrics, pengamanan endpoints (API key/JWT), validasi consent.
