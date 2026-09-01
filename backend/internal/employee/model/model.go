package model

import "time"

type Employee struct {
	PublicID string `json:"id"`
	EmployeeNumber string `json:"employee_number"`
	FirstName string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
	LastName string `json:"last_name"`
	Gender string `json:"gender,omitempty"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	NationalID string `json:"national_id,omitempty"`
	PassportNumber string `json:"passport_number,omitempty"`
	Nationality string `json:"nationality,omitempty"`
	KRAPIN string `json:"kra_pin,omitempty"`
	NSSFNumber string `json:"nssf_number,omitempty"`
	SHIFNumber string `json:"shif_number,omitempty"`
	NHIFNumber string `json:"nhif_number,omitempty"`
	EmploymentDate time.Time `json:"employment_date"`
	TerminationDate *time.Time `json:"termination_date,omitempty"`
	JobTitle string `json:"job_title,omitempty"`
	DepartmentID string `json:"department_id,omitempty"`
	Department string `json:"department,omitempty"`
	EmploymentType string `json:"employment_type,omitempty"`
	Status string `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Salary struct {
	PublicID string `json:"id"`
	BasicSalary string `json:"basic_salary"`
	PayFrequency string `json:"pay_frequency"`
	EffectiveFrom time.Time `json:"effective_from"`
	EffectiveTo *time.Time `json:"effective_to,omitempty"`
}

type EmployeeDetail struct {
	Employee
	CurrentSalary *Salary `json:"current_salary,omitempty"`
}

type CreateInput struct {
	EmployeeNumber string `json:"employee_number"`
	FirstName string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName string `json:"last_name"`
	Gender string `json:"gender"`
	DateOfBirth string `json:"date_of_birth"`
	NationalID string `json:"national_id"`
	PassportNumber string `json:"passport_number"`
	Nationality string `json:"nationality"`
	KRAPIN string `json:"kra_pin"`
	NSSFNumber string `json:"nssf_number"`
	SHIFNumber string `json:"shif_number"`
	NHIFNumber string `json:"nhif_number"`
	EmploymentDate string `json:"employment_date"`
	JobTitle string `json:"job_title"`
	DepartmentID string `json:"department_id"`
	EmploymentType string `json:"employment_type"`
	BasicSalary string `json:"basic_salary"`
	PayFrequency string `json:"pay_frequency"`
	EffectiveFrom string `json:"effective_from"`
}

type UpdateInput struct {
	FirstName string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName string `json:"last_name"`
	Gender string `json:"gender"`
	DateOfBirth string `json:"date_of_birth"`
	NationalID string `json:"national_id"`
	PassportNumber string `json:"passport_number"`
	Nationality string `json:"nationality"`
	KRAPIN string `json:"kra_pin"`
	NSSFNumber string `json:"nssf_number"`
	SHIFNumber string `json:"shif_number"`
	NHIFNumber string `json:"nhif_number"`
	EmploymentDate string `json:"employment_date"`
	TerminationDate string `json:"termination_date"`
	JobTitle string `json:"job_title"`
	DepartmentID string `json:"department_id"`
	EmploymentType string `json:"employment_type"`
	Status string `json:"status"`
}