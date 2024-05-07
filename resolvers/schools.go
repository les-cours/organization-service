package resolvers

import (
	"context"
	"github.com/les-cours/organization-service/api/orgs"
)

func (s *Server) GetSchool(ctx context.Context, in *orgs.IDRequest) (*orgs.Departments, error) {

	rows, err := s.DB.Query(`
SELECT 
    department_id,title,title_ar, description, description_ar
from departments 
WHERE schools = $1;`, in.Id)

	if err != nil {
		s.Logger.Error(err.Error())
		return nil, ErrInternal
	}

	departments := &orgs.Departments{}

	for rows.Next() {
		department := &orgs.Department{}

		err = rows.Scan(&department.DepartmentID, &department.Title, &department.ArabicTitle, &department.Description, &department.DescriptionAr)
		if err != nil {
			s.Logger.Error(err.Error())
			return nil, ErrInternal
		}

		grades, err := s.GetGrads(ctx, &orgs.GetGradsRequest{
			DepartmentID: department.DepartmentID,
		})

		department.Grades = grades

		if err != nil {
			s.Logger.Error(err.Error())
			return nil, ErrInternal
		}

		departments.Departments = append(departments.Departments, &orgs.Department{
			DepartmentID: department.DepartmentID,
			Title:        department.Title,
			ArabicTitle:  department.ArabicTitle,
			Grades:       department.Grades,
		})

	}
	return departments, nil
}
