-- name: CreateEmployee :one
INSERT INTO employees (
    employee_code,
    user_id,
    division_id,
    branch_id,
    department_id,
    designation,
    job_title,
    job_grade,
    employee_type,
    manager_id,
    department_head_id,
    date_of_joining,
    date_of_confirmation,
    pan,
    aadhaar,
    uan,
    esic_number,
    bank_name,
    bank_account_number,
    bank_ifsc,
    current_address,
    permanent_address,
    emergency_contact_name,
    emergency_contact_phone,
    emergency_contact_relation,
    work_location,
    office_phone,
    extension,
    status,
    skills,
    certifications,
    highest_qualification,
    metadata,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
    $31, $32, $33, $34
)
RETURNING *;

-- name: GetEmployeeByID :one
SELECT * FROM employees
WHERE uuid = $1 AND deleted_at IS NULL;

-- name: GetEmployeeByCode :one
SELECT * FROM employees
WHERE employee_code = $1 AND deleted_at IS NULL;

-- name: GetEmployeeByUserID :one
SELECT * FROM employees
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: UpdateEmployee :one
UPDATE employees
SET
    employee_code = COALESCE(sqlc.narg('employee_code'), employee_code),
    division_id = COALESCE(sqlc.narg('division_id'), division_id),
    branch_id = COALESCE(sqlc.narg('branch_id'), branch_id),
    department_id = COALESCE(sqlc.narg('department_id'), department_id),
    designation = COALESCE(sqlc.narg('designation'), designation),
    job_title = COALESCE(sqlc.narg('job_title'), job_title),
    job_grade = COALESCE(sqlc.narg('job_grade'), job_grade),
    employee_type = COALESCE(sqlc.narg('employee_type'), employee_type),
    manager_id = COALESCE(sqlc.narg('manager_id'), manager_id),
    department_head_id = COALESCE(sqlc.narg('department_head_id'), department_head_id),
    date_of_joining = COALESCE(sqlc.narg('date_of_joining'), date_of_joining),
    date_of_confirmation = COALESCE(sqlc.narg('date_of_confirmation'), date_of_confirmation),
    date_of_leaving = COALESCE(sqlc.narg('date_of_leaving'), date_of_leaving),
    pan = COALESCE(sqlc.narg('pan'), pan),
    aadhaar = COALESCE(sqlc.narg('aadhaar'), aadhaar),
    uan = COALESCE(sqlc.narg('uan'), uan),
    esic_number = COALESCE(sqlc.narg('esic_number'), esic_number),
    bank_name = COALESCE(sqlc.narg('bank_name'), bank_name),
    bank_account_number = COALESCE(sqlc.narg('bank_account_number'), bank_account_number),
    bank_ifsc = COALESCE(sqlc.narg('bank_ifsc'), bank_ifsc),
    current_address = COALESCE(sqlc.narg('current_address'), current_address),
    permanent_address = COALESCE(sqlc.narg('permanent_address'), permanent_address),
    emergency_contact_name = COALESCE(sqlc.narg('emergency_contact_name'), emergency_contact_name),
    emergency_contact_phone = COALESCE(sqlc.narg('emergency_contact_phone'), emergency_contact_phone),
    emergency_contact_relation = COALESCE(sqlc.narg('emergency_contact_relation'), emergency_contact_relation),
    work_location = COALESCE(sqlc.narg('work_location'), work_location),
    office_phone = COALESCE(sqlc.narg('office_phone'), office_phone),
    extension = COALESCE(sqlc.narg('extension'), extension),
    status = COALESCE(sqlc.narg('status'), status),
    termination_reason = COALESCE(sqlc.narg('termination_reason'), termination_reason),
    skills = COALESCE(sqlc.narg('skills'), skills),
    certifications = COALESCE(sqlc.narg('certifications'), certifications),
    highest_qualification = COALESCE(sqlc.narg('highest_qualification'), highest_qualification),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_by = sqlc.narg('updated_by')
WHERE uuid = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: DeleteEmployee :exec
UPDATE employees
SET deleted_at = CURRENT_TIMESTAMP
WHERE uuid = $1 AND deleted_at IS NULL;

-- name: ListEmployees :many
SELECT * FROM employees
WHERE deleted_at IS NULL
  AND (sqlc.narg('employee_code')::text IS NULL OR employee_code ILIKE '%' || sqlc.narg('employee_code')::text || '%')
  AND (sqlc.narg('designation')::text IS NULL OR designation ILIKE '%' || sqlc.narg('designation')::text || '%')
  AND (sqlc.narg('department_id')::uuid IS NULL OR department_id = sqlc.narg('department_id')::uuid)
  AND (sqlc.narg('branch_id')::uuid IS NULL OR branch_id = sqlc.narg('branch_id')::uuid)
  AND (sqlc.narg('division_id')::uuid IS NULL OR division_id = sqlc.narg('division_id')::uuid)
  AND (sqlc.narg('manager_id')::uuid IS NULL OR manager_id = sqlc.narg('manager_id')::uuid)
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text)
  AND (sqlc.narg('employee_type')::text IS NULL OR employee_type = sqlc.narg('employee_type')::text)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CountEmployees :one
SELECT COUNT(*) FROM employees
WHERE deleted_at IS NULL
  AND (sqlc.narg('employee_code')::text IS NULL OR employee_code ILIKE '%' || sqlc.narg('employee_code')::text || '%')
  AND (sqlc.narg('designation')::text IS NULL OR designation ILIKE '%' || sqlc.narg('designation')::text || '%')
  AND (sqlc.narg('department_id')::uuid IS NULL OR department_id = sqlc.narg('department_id')::uuid)
  AND (sqlc.narg('branch_id')::uuid IS NULL OR branch_id = sqlc.narg('branch_id')::uuid)
  AND (sqlc.narg('division_id')::uuid IS NULL OR division_id = sqlc.narg('division_id')::uuid)
  AND (sqlc.narg('manager_id')::uuid IS NULL OR manager_id = sqlc.narg('manager_id')::uuid)
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text)
  AND (sqlc.narg('employee_type')::text IS NULL OR employee_type = sqlc.narg('employee_type')::text);

-- name: GetEmployeesByDepartment :many
SELECT * FROM employees
WHERE department_id = $1 AND deleted_at IS NULL
ORDER BY designation, employee_code;

-- name: GetEmployeesByManager :many
SELECT * FROM employees
WHERE manager_id = $1 AND deleted_at IS NULL
ORDER BY designation, employee_code;

-- name: GetEmployeesByDivision :many
SELECT * FROM employees
WHERE division_id = $1 AND deleted_at IS NULL
ORDER BY department_id, designation, employee_code;

-- name: GetEmployeesByBranch :many
SELECT * FROM employees
WHERE branch_id = $1 AND deleted_at IS NULL
ORDER BY department_id, designation, employee_code;
