package resolvers

import (
	"database/sql"
	"github.com/les-cours/organization-service/api/users"
	"go.uber.org/zap"
)

var instance *Server

func GetInstance(db *sql.DB, userService users.UserServiceClient, logger *zap.Logger) *Server {
	if instance != nil {
		return instance
	}

	instance = &Server{
		DB:          db,
		UserService: userService,
		Logger:      logger,
	}
	return instance
}
