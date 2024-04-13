package resolvers

import (
	"database/sql"
	"github.com/les-cours/organization-service/api/users"
)

var instance *Server

func GetInstance(SQLDB *sql.DB, userService users.UserServiceClient) *Server {
	if instance != nil {
		return instance
	}

	instance = &Server{
		DB:          SQLDB,
		UserService: userService,
	}
	return instance
}
