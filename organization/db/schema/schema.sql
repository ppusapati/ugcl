-- Organization Service Database Schema
-- This schema supports divisions, branches, and departments hierarchy

-- Divisions Table
CREATE TABLE divisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    head_user_id VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    UNIQUE(tenant_id, code)
);

-- Branches Table
CREATE TABLE branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    division_id UUID NOT NULL REFERENCES divisions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    branch_type VARCHAR(50), -- 'HQ', 'Regional', 'Local', 'Satellite'

    -- Location Details
    address_line1 TEXT,
    address_line2 TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100) DEFAULT 'India',
    postal_code VARCHAR(20),
    latitude NUMERIC(10, 8),
    longitude NUMERIC(11, 8),

    -- Contact Details
    phone VARCHAR(20),
    email VARCHAR(255),

    -- Management
    branch_manager_user_id VARCHAR(255),

    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    UNIQUE(tenant_id, code)
);

-- Departments Table
CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    division_id UUID NULL REFERENCES divisions(id) ON DELETE CASCADE, -- NULL = Business-level
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    department_type VARCHAR(50), -- 'Operational', 'Support', 'Administrative'
    head_user_id VARCHAR(255), -- Department head (GM, VP)
    parent_department_id UUID REFERENCES departments(id), -- For sub-departments
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    CHECK (
        -- Either business-level (division_id IS NULL)
        -- OR division-level (division_id IS NOT NULL)
        (division_id IS NULL AND parent_department_id IS NULL) OR
        (division_id IS NOT NULL)
    )
);

-- Create unique index for departments (handling NULL division_id)
CREATE UNIQUE INDEX idx_departments_unique_code ON departments (
    tenant_id,
    code,
    COALESCE(division_id::text, '')
);

-- Indexes for Performance
CREATE INDEX idx_divisions_tenant_id ON divisions(tenant_id);
CREATE INDEX idx_divisions_code ON divisions(tenant_id, code);
CREATE INDEX idx_divisions_is_active ON divisions(is_active) WHERE is_active = true;
CREATE INDEX idx_divisions_head_user ON divisions(head_user_id) WHERE head_user_id IS NOT NULL;

CREATE INDEX idx_branches_division_id ON branches(division_id);
CREATE INDEX idx_branches_tenant_id ON branches(tenant_id);
CREATE INDEX idx_branches_code ON branches(tenant_id, code);
CREATE INDEX idx_branches_is_active ON branches(is_active) WHERE is_active = true;
CREATE INDEX idx_branches_city ON branches(city);
CREATE INDEX idx_branches_state ON branches(state);
CREATE INDEX idx_branches_manager ON branches(branch_manager_user_id) WHERE branch_manager_user_id IS NOT NULL;

CREATE INDEX idx_departments_tenant_id ON departments(tenant_id);
CREATE INDEX idx_departments_division_id ON departments(division_id) WHERE division_id IS NOT NULL;
CREATE INDEX idx_departments_parent_id ON departments(parent_department_id) WHERE parent_department_id IS NOT NULL;
CREATE INDEX idx_departments_code ON departments(tenant_id, code);
CREATE INDEX idx_departments_is_active ON departments(is_active) WHERE is_active = true;
CREATE INDEX idx_departments_head_user ON departments(head_user_id) WHERE head_user_id IS NOT NULL;

-- Triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_divisions_updated_at
    BEFORE UPDATE ON divisions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_branches_updated_at
    BEFORE UPDATE ON branches
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_departments_updated_at
    BEFORE UPDATE ON departments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Views for Common Queries

-- Division Summary View
CREATE VIEW division_summary AS
SELECT
    d.id,
    d.tenant_id,
    d.code,
    d.name,
    d.description,
    d.head_user_id,
    d.is_active,
    d.display_order,
    COUNT(DISTINCT b.id) as branch_count,
    COUNT(DISTINCT dept.id) FILTER (WHERE dept.division_id = d.id) as department_count,
    d.created_at,
    d.updated_at
FROM divisions d
LEFT JOIN branches b ON d.id = b.division_id AND b.is_active = true
LEFT JOIN departments dept ON d.id = dept.division_id AND dept.is_active = true
GROUP BY d.id, d.tenant_id, d.code, d.name, d.description, d.head_user_id,
         d.is_active, d.display_order, d.created_at, d.updated_at;

-- Branch Summary View
CREATE VIEW branch_summary AS
SELECT
    b.id,
    b.division_id,
    d.name as division_name,
    d.code as division_code,
    b.tenant_id,
    b.code,
    b.name,
    b.branch_type,
    b.city,
    b.state,
    b.country,
    b.branch_manager_user_id,
    b.is_active,
    b.display_order,
    b.created_at,
    b.updated_at
FROM branches b
INNER JOIN divisions d ON b.division_id = d.id;

-- Department Hierarchy View
CREATE VIEW department_hierarchy AS
WITH RECURSIVE dept_tree AS (
    -- Base case: root departments
    SELECT
        id,
        tenant_id,
        division_id,
        code,
        name,
        description,
        department_type,
        head_user_id,
        parent_department_id,
        is_active,
        display_order,
        1 as level,
        ARRAY[id] as path,
        code as full_path
    FROM departments
    WHERE parent_department_id IS NULL

    UNION ALL

    -- Recursive case: child departments
    SELECT
        d.id,
        d.tenant_id,
        d.division_id,
        d.code,
        d.name,
        d.description,
        d.department_type,
        d.head_user_id,
        d.parent_department_id,
        d.is_active,
        d.display_order,
        dt.level + 1,
        dt.path || d.id,
        dt.full_path || ' > ' || d.code
    FROM departments d
    INNER JOIN dept_tree dt ON d.parent_department_id = dt.id
)
SELECT * FROM dept_tree;

-- Organization Hierarchy View (Full tree)
CREATE VIEW organization_full_hierarchy AS
SELECT
    d.id as division_id,
    d.code as division_code,
    d.name as division_name,
    b.id as branch_id,
    b.code as branch_code,
    b.name as branch_name,
    b.city,
    b.state,
    dept.id as department_id,
    dept.code as department_code,
    dept.name as department_name,
    dept.department_type,
    CASE
        WHEN dept.division_id IS NULL THEN 'Business-Level'
        ELSE 'Division-Level'
    END as department_scope
FROM divisions d
LEFT JOIN branches b ON d.id = b.division_id AND b.is_active = true
LEFT JOIN departments dept ON (dept.division_id = d.id OR dept.division_id IS NULL)
    AND dept.is_active = true
WHERE d.is_active = true;
