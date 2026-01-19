-- employees schema

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS employees (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    employee_code VARCHAR(50) UNIQUE NOT NULL,
    user_id UUID NOT NULL,  -- References identity.users

    -- Organizational hierarchy
    division_id UUID,
    branch_id UUID,
    department_id UUID,

    -- Job information
    designation VARCHAR(100) NOT NULL,
    job_title VARCHAR(150),
    job_grade VARCHAR(50),
    employee_type VARCHAR(50) DEFAULT 'PERMANENT',  -- PERMANENT, TEMPORARY, PROBATION

    -- Reporting structure
    manager_id UUID,  -- References employees.uuid
    department_head_id UUID,  -- References employees.uuid

    -- Employment dates
    date_of_joining TIMESTAMP NOT NULL,
    date_of_confirmation TIMESTAMP,
    date_of_leaving TIMESTAMP,

    -- Personal information
    pan VARCHAR(20) UNIQUE,
    aadhaar VARCHAR(20),
    uan VARCHAR(50),  -- Universal Account Number (PF)
    esic_number VARCHAR(50),

    -- Bank details
    bank_name VARCHAR(100),
    bank_account_number VARCHAR(50),
    bank_ifsc VARCHAR(20),

    -- Address
    current_address TEXT,
    permanent_address TEXT,

    -- Emergency contact
    emergency_contact_name VARCHAR(100),
    emergency_contact_phone VARCHAR(20),
    emergency_contact_relation VARCHAR(50),

    -- Work location
    work_location VARCHAR(150),
    office_phone VARCHAR(20),
    extension VARCHAR(10),

    -- Status
    status VARCHAR(50) DEFAULT 'ACTIVE',  -- ACTIVE, INACTIVE, ON_LEAVE, TERMINATED, RESIGNED
    termination_reason TEXT,

    -- Skills and qualifications
    skills TEXT[],  -- Array of skills
    certifications TEXT[],  -- Array of certifications
    highest_qualification VARCHAR(100),

    -- Metadata
    metadata JSONB,

    -- Audit fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMP,

    -- Constraints
    CONSTRAINT chk_employment_dates CHECK (
        date_of_leaving IS NULL OR date_of_leaving > date_of_joining
    ),
    CONSTRAINT chk_confirmation_date CHECK (
        date_of_confirmation IS NULL OR date_of_confirmation >= date_of_joining
    )
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_employees_user_id ON employees(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_employee_code ON employees(employee_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_division ON employees(division_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_branch ON employees(branch_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_department ON employees(department_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_manager ON employees(manager_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_status ON employees(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_designation ON employees(designation) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_employee_type ON employees(employee_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_pan ON employees(pan) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_skills ON employees USING GIN(skills) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_certifications ON employees USING GIN(certifications) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_joining_date ON employees(date_of_joining) WHERE deleted_at IS NULL;

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
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_employees_updated_at'
    ) THEN
        CREATE TRIGGER update_employees_updated_at
            BEFORE UPDATE ON employees
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- Comments
COMMENT ON TABLE employees IS 'Stores employee information and organizational hierarchy';
COMMENT ON COLUMN employees.employee_code IS 'Unique identifier for employee (e.g., EMP001)';
COMMENT ON COLUMN employees.user_id IS 'References the user account in identity.users';
COMMENT ON COLUMN employees.manager_id IS 'References the employee who is this employee''s manager';
COMMENT ON COLUMN employees.division_id IS 'References the division in organization module';
COMMENT ON COLUMN employees.branch_id IS 'References the branch in organization module';
COMMENT ON COLUMN employees.department_id IS 'References the department in organization module';
