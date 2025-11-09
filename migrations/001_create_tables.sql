CREATE TABLE IF NOT EXISTS observations_raw (
  id BIGSERIAL PRIMARY KEY,
  device_id TEXT,
  patient_id TEXT,
  obs_type TEXT,
  value DOUBLE PRECISION,
  unit TEXT,
  observed_at TIMESTAMPTZ,
  received_at TIMESTAMPTZ DEFAULT now(),
  status TEXT DEFAULT 'pending',
  retry_count INT DEFAULT 0,
  last_error TEXT
);

CREATE TABLE IF NOT EXISTS fhir_transactions (
  id BIGSERIAL PRIMARY KEY,
  observation_raw_id BIGINT REFERENCES observations_raw(id),
  fhir_payload JSONB,
  response_code INT,
  response_body TEXT,
  sent_at TIMESTAMPTZ DEFAULT now(),
  status TEXT
);
