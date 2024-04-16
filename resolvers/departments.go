package resolvers

import (
	"context"
	"errors"
	"github.com/les-cours/organization-service/api/orgs"
	"github.com/les-cours/organization-service/utils"
	"log"
)

func (s *Server) GetDepartments(ctx context.Context, in *orgs.GetDepartmentsRequest) (*orgs.Departments, error) {

	rows, err := s.DB.Query(`
SELECT 
    department_id,name,arabic_name 
from departments 
WHERE schools = $1;`, in.GetSchoolID())

	if err != nil {
		log.Println("err whene query departments")
		return nil, err
	}

	var departments *orgs.Departments

	var departmentID, name, arabicName string
	for rows.Next() {
		err = rows.Scan(&departmentID, &name, &arabicName)
		if err != nil {
			log.Println("err when scan departments")
			return nil, err
		}
		departments.Departments = append(departments.Departments, &orgs.Department{
			DepartmentID: departmentID,
			Name:         name,
			ArabicName:   arabicName,
			Schools:      in.GetSchoolID(),
		})

	}
	return departments, nil
}

func (s *Server) GetDepartment(ctx context.Context, in *orgs.GetDepartmentRequest) (*orgs.Department, error) {
	var department *orgs.Department
	department.DepartmentID = in.DepartmentID
	err := s.DB.QueryRow(`
SELECT 
    schools,name,arabic_name 
from departments 
WHERE department_id = $1;`, in.GetDepartmentID()).Scan(&department.Schools, &department.Name, &department.ArabicName)

	if err != nil {
		log.Println("err when query department")
		return nil, err
	}

	return department, nil
}

func (s *Server) AddDepartment(ctx context.Context, in *orgs.DepartmentAddRequest) (*orgs.Department, error) {
	stmt, err := s.DB.Prepare(`
    INSERT INTO departments (department_id, name, arabic_name, schools)
    VALUES ($1, $2, $3,$4)
    RETURNING department_id;
  `)
	if err != nil {
		log.Println("error preparing create department statement:", err)
		return nil, err
	}
	defer stmt.Close() // Ensure statement is closed even on errors

	var newDepartment *orgs.Department
	err = stmt.QueryRow(utils.GenerateUUIDString(), in.Name, in.ArabicName, in.Schools).Scan(&newDepartment)
	if err != nil {
		log.Println("error creating department:", err)
		return nil, err
	}

	log.Println("Department created successfully")
	return newDepartment, nil
}

func (s *Server) UpdateDepartment(ctx context.Context, in *orgs.DepartmentUpdateRequest) (*orgs.Department, error) {
	stmt, err := s.DB.Prepare(`
    UPDATE departments
    SET name = $1, arabic_name = $2
    WHERE department_id = $3;
  `)
	if err != nil {
		log.Println("error preparing update department statement:", err)
		return nil, err
	}
	defer stmt.Close() // Ensure statement is closed even on errors

	_, err = stmt.Exec(in.Name, in.ArabicName, in.DepartmentID)
	if err != nil {
		log.Println("error updating department:", err)
		return nil, err
	}

	department, err := s.GetDepartment(ctx, &orgs.GetDepartmentRequest{DepartmentID: in.DepartmentID})
	if err != nil {
		log.Println("error getting department:", err)
		return nil, err
	}

	return department, nil
}

func (s *Server) DeleteDepartment(ctx context.Context, in *orgs.DeleteDepartmentsRequest) (*orgs.OperationStatus, error) {
	stmt, err := s.DB.Prepare(`
    DELETE FROM departments
    WHERE department_id = $1;
  `)
	if err != nil {
		log.Println("error preparing delete department statement:", err)
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(in.DepartmentID)
	if err != nil {
		log.Println("error deleting department:", err)
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("error checking rows affected:", err)
		return nil, err
	}

	if rowsAffected == 0 {
		log.Println("No department found with ID:", in.DepartmentID)
		return nil, errors.New("no department found with ID")
	}

	log.Println("Department deleted successfully")
	return &orgs.OperationStatus{Status: true}, nil
}

func (s *Server) DeleteDepartments(ctx context.Context, in *orgs.MultiDepartmentsDeleteRequest) (*orgs.OperationStatus, error) {

	for _, id := range in.DepartmentsIDs {
		_, err := s.DeleteDepartment(ctx, &orgs.DeleteDepartmentsRequest{DepartmentID: id})
		if err != nil {
			return nil, err
		}
	}

	return &orgs.OperationStatus{Status: true}, nil
}

func (s *Server) existDepartment(depID string) bool {

	return true
}
