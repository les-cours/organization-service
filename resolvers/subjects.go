package resolvers

import (
	"context"
	"database/sql"
	"errors"
	"github.com/les-cours/organization-service/api/orgs"
	"github.com/les-cours/organization-service/api/users"
	"github.com/les-cours/organization-service/utils"
	"log"
)

func (s *Server) GetSubjects(ctx context.Context, in *orgs.GetSubjectsRequest) (*orgs.Subjects, error) {

	rows, err := s.DB.Query(`
SELECT 
    s.subject_id,s.title,s.title_ar 
from subjects as s INNER JOIN public.grades_subjects gs on s.subject_id = gs.subject_id
WHERE grade_id = $1;`, in.GetGradID())

	if err != nil {
		s.Logger.Error(err.Error())
		return nil, ErrInternal
	}

	var subjects = new(orgs.Subjects)

	var subject *orgs.Subject
	for rows.Next() {
		err = rows.Scan(&subject.SubjectID, &subject.Title, &subject.ArabicTitle)
		if err != nil {
			log.Println("err when scan subjects")
			return nil, err
		}
		subjects.Subjects = append(subjects.Subjects, subject)
	}
	return subjects, nil
}

func (s *Server) GetSubject(ctx context.Context, in *orgs.GetSubjectRequest) (*orgs.Subject, error) {

	var subject *orgs.Subject
	var teachers []*users.Teacher
	var grads []*orgs.Grad

	subject.SubjectID = in.SubjectID

	err := s.DB.QueryRow(`
SELECT 
    title,title_ar
from subjects 
WHERE subject_id = $1;`, in.GetSubjectID()).Scan(&subject.Title, &subject.ArabicTitle)

	if err != nil {
		log.Println("err when query subject")
		return nil, err
	}

	var rows *sql.Rows

	//grads
	rows, err = s.DB.Query(`
SELECT 
    g.grade_id, g.title,g.title_ar 
FROM grades as g
    INNER JOIN 
    public.grades_subjects gs 
        on g.grade_id = gs.grade_id
WHERE gs.subject_id = $1;
        `, in.SubjectID)

	var grad *orgs.Grad
	for rows.Next() {
		err = rows.Scan(&grad.GradID, &grad.Title, &grad.ArabicTitle)
		if err != nil {
			return nil, err
		}
		grads = append(grads, grad)
	}

	//teachers

	go func() {
		var res *users.Teachers
		res, err = s.UserService.GetTeacherBySubject(ctx, &users.GetTeacherBySubjectRequest{
			SubjectID: in.SubjectID,
		})
		if err != nil {

		}
		teachers = res.Teachers
	}()

	subject.Teachers = teachers
	subject.Grads = grads
	return subject, nil
}

func (s *Server) AddSubject(ctx context.Context, in *orgs.SubjectAddRequest) (*orgs.Subject, error) {

	subjectID := utils.GenerateUUIDString()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	stmt, err := tx.Prepare(`
    INSERT INTO subjects (subject_id, title, title_ar)
    VALUES ($1, $2, $3)
  `)
	if err != nil {
		tx.Rollback()
		log.Println("error preparing create subject statement:", err)
		return nil, err
	}

	defer stmt.Close()

	_, err = stmt.Exec(subjectID, in.Title, in.ArabicTitle)
	if err != nil {
		tx.Rollback()
		log.Println("error creating subject:", err)
		return nil, err
	}

	// add subject to the grads :

	for _, gradID := range in.GradsIDs {
		_, err = tx.Exec(`INSERT INTO grades_subjects (subject_id, grade_id) values ($1,$2)`, subjectID, gradID)
		if err != nil {
			tx.Rollback()
			log.Println("error assigning grade to subject:", err)
			return nil, err
		}
	}

	// add teachers to the subject

	for _, teacherID := range in.TeachersIDs {
		_, err = tx.Exec(`INSERT INTO teacher_subjects (subject_id, teacher_id) values ($1,$2)`, subjectID, teacherID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	subjects, err := s.GetSubject(ctx, &orgs.GetSubjectRequest{
		SubjectID: subjectID,
	})
	if err != nil {
		return nil, err
	}
	return subjects, nil
}

func (s *Server) UpdateSubject(ctx context.Context, in *orgs.SubjectUpdateRequest) (*orgs.Subject, error) {
	stmt, err := s.DB.Prepare(`
    UPDATE subjects
    SET title = $1, title_ar = $2
    WHERE subject_id = $3;
  `)
	if err != nil {
		log.Println("error preparing update subject statement:", err)
		return nil, err
	}
	defer stmt.Close() // Ensure statement is closed even on errors

	_, err = stmt.Exec(in.Title, in.ArabicTitle, in.SubjectID)
	if err != nil {
		log.Println("error updating subject:", err)
		return nil, err
	}

	subject, err := s.GetSubject(ctx, &orgs.GetSubjectRequest{SubjectID: in.SubjectID})
	if err != nil {
		log.Println("error getting subject:", err)
		return nil, err
	}

	return subject, nil
}

func (s *Server) DeleteSubject(ctx context.Context, in *orgs.DeleteSubjectsRequest) (*orgs.OperationStatus, error) {
	stmt, err := s.DB.Prepare(`
    DELETE FROM subjects
    WHERE subject_id = $1;
  `)
	if err != nil {
		log.Println("error preparing delete subject statement:", err)
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(in.SubjectID)
	if err != nil {
		log.Println("error deleting subject:", err)
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("error checking rows affected:", err)
		return nil, err
	}

	if rowsAffected == 0 {
		log.Println("No subject found with ID:", in.SubjectID)
		return nil, errors.New("no subject found with ID")
	}

	log.Println("Subject deleted successfully")
	return &orgs.OperationStatus{Status: true}, nil
}

func (s *Server) DeleteSubjects(ctx context.Context, in *orgs.MultiSubjectsDeleteRequest) (*orgs.OperationStatus, error) {

	for _, id := range in.SubjectsIDs {
		_, err := s.DeleteSubject(ctx, &orgs.DeleteSubjectsRequest{SubjectID: id})
		if err != nil {
			return nil, err
		}
	}

	return &orgs.OperationStatus{Status: true}, nil
}

func (s *Server) GetSubjectsByGrad(ctx context.Context, in *orgs.IDRequest) (*orgs.Subjects, error) {

	rows, err := s.DB.Query(`
SELECT subjects.subject_id, subjects.title , subjects.title_ar 
from subjects
    INNER JOIN  grades_subjects gs on subjects.subject_id = gs.subject_id
    WHERE gs.grade_id = $1;
    `, in.Id)

	if err != nil {
		s.Logger.Error(err.Error())
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound("subjects")
		}
		return nil, ErrInternal
	}

	var subjects = new(orgs.Subjects)
	for rows.Next() {
		subject := &orgs.Subject{}
		err = rows.Scan(&subject.SubjectID, &subject.Title, &subject.ArabicTitle)
		if err != nil {
			s.Logger.Error(err.Error())
			return nil, ErrInternal
		}
		subjects.Subjects = append(subjects.Subjects, subject)
	}

	return subjects, nil
}
