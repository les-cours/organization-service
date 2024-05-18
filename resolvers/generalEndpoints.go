package resolvers

import (
	"context"
	"github.com/les-cours/organization-service/api/orgs"
)

func (s *Server) GetCities(ctx context.Context, empty *orgs.Empty) (*orgs.Cities, error) {

	rows, err := s.DB.Query(`SELECT  id ,city_name ,city_name_ar  from cities;`)

	if err != nil {
		s.Logger.Error(err.Error())
		return nil, ErrInternal
	}

	cities := new(orgs.Cities)
	for rows.Next() {
		city := &orgs.City{}
		err = rows.Scan(&city.Id, &city.Name, &city.ArabicName)
		if err != nil {
			s.Logger.Error(err.Error())
			return nil, ErrInternal
		}

		cities.Cities = append(cities.Cities, city)
	}
	return cities, nil
}
