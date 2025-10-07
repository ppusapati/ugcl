-- schema.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS contractors (
     id         BIGSERIAL PRIMARY KEY,
    uuid       UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    company_name VARCHAR(255) NOT NULL,
    company_type VARCHAR(100),
    gst VARCHAR(50),
    pan VARCHAR(50) UNIQUE,
    category VARCHAR(100),
    person_id BIGINT REFERENCES users(id),
    associated_project TEXT[], -- array of project IDs
    working_site TEXT[], -- array of site IDs
    contract_start_date TIMESTAMP,
    contract_end_date TIMESTAMP,
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT chk_contract_dates CHECK (contract_end_date IS NULL OR contract_end_date > contract_start_date)
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_contractors_person_id ON contractors(person_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_status ON contractors(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_company ON contractors(company_name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_category ON contractors(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_pan ON contractors(pan) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_gst ON contractors(gst) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_projects ON contractors USING GIN(associated_project) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_sites ON contractors USING GIN(working_site) WHERE deleted_at IS NULL;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_contractors_updated_at'
    ) THEN
        CREATE TRIGGER update_contractors_updated_at 
            BEFORE UPDATE ON contractors
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;