package resolvers

import (
	"database/sql"
	"github.com/les-cours/organization-service/api/orgs"
	"github.com/les-cours/organization-service/api/users"
)

type Server struct {
	DB          *sql.DB
	UserService users.UserServiceClient
	orgs.UnimplementedOrgServiceServer
}
