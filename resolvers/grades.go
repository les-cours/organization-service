package resolvers

import (
	"context"
	"github.com/les-cours/organization-service/api/orgs"
	"github.com/les-cours/organization-service/utils"
	"log"
)

func (s *Server) GetGrads(ctx context.Context, in *orgs.GetGradsRequest) (*orgs.Grads, error) {

	rows, err := s.DB.Query(`
SELECT 
    g.grade_id,g.title,g.title_ar 
from grades as g
WHERE  department_id = $1;`, in.GetDepartmentID())

	if err != nil {
		s.Logger.Error(err.Error())
		return nil, err
	}

	grads := &orgs.Grads{}

	for rows.Next() {
		grad := &orgs.Grad{}
		err = rows.Scan(&grad.GradID, &grad.Title, &grad.ArabicTitle)
		if err != nil {
			s.Logger.Error(err.Error())
			return nil, err
		}
		subjects, err := s.GetSubjectsByGrad(ctx, &orgs.IDRequest{
			Id: grad.GradID,
		})
		if err != nil {
			s.Logger.Error(err.Error())
			return nil, ErrInternal
		}
		grad.Subjects = subjects
		grads.Grads = append(grads.Grads, grad)
	}
	return grads, nil
}

func (s *Server) GetGrad(ctx context.Context, in *orgs.GetGradRequest) (*orgs.Grad, error) {

	var grad *orgs.Grad
	grad.GradID = in.GradID

	err := s.DB.QueryRow(`
SELECT 
    title,title_ar
from grades 
WHERE grade_id = $1;`, in.GetGradID()).Scan(&grad.Title, &grad.ArabicTitle)

	if err != nil {
		s.Logger.Error(err.Error())
		return nil, ErrInternal
	}

	return grad, nil
}

func (s *Server) AddGrad(ctx context.Context, in *orgs.GradAddRequest) (*orgs.Grad, error) {

	existDepartment := s.existDepartment(in.DepartmentID)

	if !existDepartment {
		return nil, ErrExistInput("department")
	}

	gradID := utils.GenerateUUIDString()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	stmt, err := tx.Prepare(`
    INSERT INTO grades (grade_id, title, title_ar,department_id)
    VALUES ($1, $2, $3,$4)
  `)
	if err != nil {
		tx.Rollback()
		log.Println("error preparing create grad statement:", err)
		return nil, err
	}

	defer stmt.Close()

	_, err = stmt.Exec(gradID, in.Title, in.ArabicTitle, in.DepartmentID)
	if err != nil {
		tx.Rollback()
		log.Println("error creating grad:", err)
		return nil, err
	}

	grads, err := s.GetGrad(ctx, &orgs.GetGradRequest{
		GradID: gradID,
	})
	if err != nil {
		return nil, err
	}
	return grads, nil
}

func (s *Server) UpdateGrad(ctx context.Context, in *orgs.GradUpdateRequest) (*orgs.Grad, error) {
	stmt, err := s.DB.Prepare(`
    UPDATE grades
    SET title = $1, title_ar = $2
    WHERE grade_id = $3;
  `)
	if err != nil {
		log.Println("error preparing update grad statement:", err)
		return nil, err
	}
	defer stmt.Close() // Ensure statement is closed even on errors

	_, err = stmt.Exec(in.Title, in.ArabicTitle, in.GradID)
	if err != nil {
		log.Println("error updating grad:", err)
		return nil, err
	}

	grad, err := s.GetGrad(ctx, &orgs.GetGradRequest{GradID: in.GradID})
	if err != nil {
		log.Println("error getting grad:", err)
		return nil, err
	}

	return grad, nil
}
