// repository/employee_repository.go
package repository

import (
	"context"
	"database/sql"
	"fmt"

	db "p9e.in/ugcl/employee/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
)

// IEmployeeRepository defines employee data access methods
type IEmployeeRepository interface {
	Create(ctx context.Context, employeeCode string, userID pgtype.UUID, divisionID pgtype.UUID, branchID pgtype.UUID, departmentID pgtype.UUID, designation string, jobTitle pgtype.Text, jobGrade pgtype.Text, employeeType pgtype.Text, managerID pgtype.UUID, departmentHeadID pgtype.UUID, dateOfJoining pgtype.Timestamp, dateOfConfirmation pgtype.Timestamp, pan pgtype.Text, aadhaar pgtype.Text, uan pgtype.Text, esicNumber pgtype.Text, bankName pgtype.Text, bankAccountNumber pgtype.Text, bankIfsc pgtype.Text, currentAddress pgtype.Text, permanentAddress pgtype.Text, emergencyContactName pgtype.Text, emergencyContactPhone pgtype.Text, emergencyContactRelation pgtype.Text, workLocation pgtype.Text, officePhone pgtype.Text, extension pgtype.Text, status pgtype.Text, skills []string, certifications []string, highestQualification pgtype.Text, metadata []byte, createdBy pgtype.UUID) (*db.Employee, error)
	GetByID(ctx context.Context, id uuid.UUID) (*db.Employee, error)
	GetByCode(ctx context.Context, code string) (*db.Employee, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*db.Employee, error)
	Update(ctx context.Context, employeeCode pgtype.Text, divisionID pgtype.UUID, branchID pgtype.UUID, departmentID pgtype.UUID, designation pgtype.Text, jobTitle pgtype.Text, jobGrade pgtype.Text, employeeType pgtype.Text, managerID pgtype.UUID, departmentHeadID pgtype.UUID, dateOfJoining pgtype.Timestamp, dateOfConfirmation pgtype.Timestamp, dateOfLeaving pgtype.Timestamp, pan pgtype.Text, aadhaar pgtype.Text, uan pgtype.Text, esicNumber pgtype.Text, bankName pgtype.Text, bankAccountNumber pgtype.Text, bankIfsc pgtype.Text, currentAddress pgtype.Text, permanentAddress pgtype.Text, emergencyContactName pgtype.Text, emergencyContactPhone pgtype.Text, emergencyContactRelation pgtype.Text, workLocation pgtype.Text, officePhone pgtype.Text, extension pgtype.Text, status pgtype.Text, terminationReason pgtype.Text, skills []string, certifications []string, highestQualification pgtype.Text, metadata []byte, updatedBy pgtype.UUID, id uuid.UUID) (*db.Employee, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, employeeCode pgtype.Text, designation pgtype.Text, departmentID pgtype.UUID, branchID pgtype.UUID, divisionID pgtype.UUID, managerID pgtype.UUID, status pgtype.Text, employeeType pgtype.Text, offset int32, limit int32) ([]db.Employee, int64, error)
	CountEmployees(ctx context.Context, employeeCode pgtype.Text, designation pgtype.Text, departmentID pgtype.UUID, branchID pgtype.UUID, divisionID pgtype.UUID, managerID pgtype.UUID, status pgtype.Text, employeeType pgtype.Text) (int64, error)
	GetByDepartment(ctx context.Context, departmentID uuid.UUID) ([]db.Employee, error)
	GetByManager(ctx context.Context, managerID uuid.UUID) ([]db.Employee, error)
}

type EmployeeRepository struct {
	queries *db.Queries
}

// NewEmployeeRepository creates a new employee repository with fx
func NewEmployeeRepository(q *db.Queries) IEmployeeRepository {
	return &EmployeeRepository{
		queries: q,
	}
}

// Create creates a new employee
func (r *EmployeeRepository) Create(ctx context.Context, employeeCode string, userID pgtype.UUID, divisionID pgtype.UUID, branchID pgtype.UUID, departmentID pgtype.UUID, designation string, jobTitle pgtype.Text, jobGrade pgtype.Text, employeeType pgtype.Text, managerID pgtype.UUID, departmentHeadID pgtype.UUID, dateOfJoining pgtype.Timestamp, dateOfConfirmation pgtype.Timestamp, pan pgtype.Text, aadhaar pgtype.Text, uan pgtype.Text, esicNumber pgtype.Text, bankName pgtype.Text, bankAccountNumber pgtype.Text, bankIfsc pgtype.Text, currentAddress pgtype.Text, permanentAddress pgtype.Text, emergencyContactName pgtype.Text, emergencyContactPhone pgtype.Text, emergencyContactRelation pgtype.Text, workLocation pgtype.Text, officePhone pgtype.Text, extension pgtype.Text, status pgtype.Text, skills []string, certifications []string, highestQualification pgtype.Text, metadata []byte, createdBy pgtype.UUID) (*db.Employee, error) {
	employee, err := r.queries.CreateEmployee(ctx, employeeCode, userID, divisionID, branchID, departmentID, designation, jobTitle, jobGrade, employeeType, managerID, departmentHeadID, dateOfJoining, dateOfConfirmation, pan, aadhaar, uan, esicNumber, bankName, bankAccountNumber, bankIfsc, currentAddress, permanentAddress, emergencyContactName, emergencyContactPhone, emergencyContactRelation, workLocation, officePhone, extension, status, skills, certifications, highestQualification, metadata, createdBy)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return nil, fmt.Errorf("employee with code already exists")
			}
		}
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}
	return &employee, nil
}

// GetByID retrieves an employee by UUID
func (r *EmployeeRepository) GetByID(ctx context.Context, id uuid.UUID) (*db.Employee, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(id.String()); err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	employee, err := r.queries.GetEmployeeByID(ctx, pgUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}
	return &employee, nil
}

// GetByCode retrieves an employee by employee code
func (r *EmployeeRepository) GetByCode(ctx context.Context, code string) (*db.Employee, error) {
	employee, err := r.queries.GetEmployeeByCode(ctx, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee by code: %w", err)
	}
	return &employee, nil
}

// GetByUserID retrieves an employee by user ID
func (r *EmployeeRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*db.Employee, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(userID.String()); err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	employee, err := r.queries.GetEmployeeByUserID(ctx, pgUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("employee not found for user")
		}
		return nil, fmt.Errorf("failed to get employee by user ID: %w", err)
	}
	return &employee, nil
}

// Update updates an employee
func (r *EmployeeRepository) Update(ctx context.Context, employeeCode pgtype.Text, divisionID pgtype.UUID, branchID pgtype.UUID, departmentID pgtype.UUID, designation pgtype.Text, jobTitle pgtype.Text, jobGrade pgtype.Text, employeeType pgtype.Text, managerID pgtype.UUID, departmentHeadID pgtype.UUID, dateOfJoining pgtype.Timestamp, dateOfConfirmation pgtype.Timestamp, dateOfLeaving pgtype.Timestamp, pan pgtype.Text, aadhaar pgtype.Text, uan pgtype.Text, esicNumber pgtype.Text, bankName pgtype.Text, bankAccountNumber pgtype.Text, bankIfsc pgtype.Text, currentAddress pgtype.Text, permanentAddress pgtype.Text, emergencyContactName pgtype.Text, emergencyContactPhone pgtype.Text, emergencyContactRelation pgtype.Text, workLocation pgtype.Text, officePhone pgtype.Text, extension pgtype.Text, status pgtype.Text, terminationReason pgtype.Text, skills []string, certifications []string, highestQualification pgtype.Text, metadata []byte, updatedBy pgtype.UUID, id uuid.UUID) (*db.Employee, error) {
	var pgID pgtype.UUID
	if err := pgID.Scan(id.String()); err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	employee, err := r.queries.UpdateEmployee(ctx, employeeCode, divisionID, branchID, departmentID, designation, jobTitle, jobGrade, employeeType, managerID, departmentHeadID, dateOfJoining, dateOfConfirmation, dateOfLeaving, pan, aadhaar, uan, esicNumber, bankName, bankAccountNumber, bankIfsc, currentAddress, permanentAddress, emergencyContactName, emergencyContactPhone, emergencyContactRelation, workLocation, officePhone, extension, status, terminationReason, skills, certifications, highestQualification, metadata, updatedBy, pgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}
	return &employee, nil
}

// Delete soft deletes an employee
func (r *EmployeeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(id.String()); err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	err := r.queries.DeleteEmployee(ctx, pgUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("employee not found")
		}
		return fmt.Errorf("failed to delete employee: %w", err)
	}
	return nil
}

// List lists employees with filters
func (r *EmployeeRepository) List(ctx context.Context, employeeCode pgtype.Text, designation pgtype.Text, departmentID pgtype.UUID, branchID pgtype.UUID, divisionID pgtype.UUID, managerID pgtype.UUID, status pgtype.Text, employeeType pgtype.Text, offset int32, limit int32) ([]db.Employee, int64, error) {
	// Get count
	total, err := r.queries.CountEmployees(ctx, employeeCode, designation, departmentID, branchID, divisionID, managerID, status, employeeType)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count employees: %w", err)
	}

	// Get list
	employees, err := r.queries.ListEmployees(ctx, employeeCode, designation, departmentID, branchID, divisionID, managerID, status, employeeType, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list employees: %w", err)
	}

	return employees, total, nil
}

// CountEmployees counts employees with filters
func (r *EmployeeRepository) CountEmployees(ctx context.Context, employeeCode pgtype.Text, designation pgtype.Text, departmentID pgtype.UUID, branchID pgtype.UUID, divisionID pgtype.UUID, managerID pgtype.UUID, status pgtype.Text, employeeType pgtype.Text) (int64, error) {
	count, err := r.queries.CountEmployees(ctx, employeeCode, designation, departmentID, branchID, divisionID, managerID, status, employeeType)
	if err != nil {
		return 0, fmt.Errorf("failed to count employees: %w", err)
	}
	return count, nil
}

// GetByDepartment lists employees by department
func (r *EmployeeRepository) GetByDepartment(ctx context.Context, departmentID uuid.UUID) ([]db.Employee, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(departmentID.String()); err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	employees, err := r.queries.GetEmployeesByDepartment(ctx, pgUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list employees by department: %w", err)
	}
	return employees, nil
}

// GetByManager lists employees by manager
func (r *EmployeeRepository) GetByManager(ctx context.Context, managerID uuid.UUID) ([]db.Employee, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(managerID.String()); err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	employees, err := r.queries.GetEmployeesByManager(ctx, pgUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list employees by manager: %w", err)
	}
	return employees, nil
}
