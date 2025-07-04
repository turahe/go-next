package services

import (
	"wordpress-go-next/backend/pkg/database"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

var Enforcer *casbin.Enforcer

func InitCasbin() error {
	mconf := `
[request_definition]
 r = sub, obj, act

[policy_definition]
 p = sub, obj, act

[role_definition]
 g = _, _

[policy_effect]
 e = some(where (p.eft == allow))

[matchers]
 m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
	m, err := model.NewModelFromString(mconf)
	if err != nil {
		return err
	}
	adapter, err := gormadapter.NewAdapterByDB(database.DB)
	if err != nil {
		return err
	}
	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return err
	}
	Enforcer = e
	// Policies: admin can do anything, editor can manage posts/comments, moderator can delete comments, user can read and comment, guest can only read
	Enforcer.AddPolicy("admin", "/api/users", "GET")
	Enforcer.AddPolicy("admin", "/api/users", "PUT")
	Enforcer.AddPolicy("admin", "/api/users", "DELETE")
	Enforcer.AddPolicy("admin", "/api/posts", "POST")
	Enforcer.AddPolicy("admin", "/api/posts", "PUT")
	Enforcer.AddPolicy("admin", "/api/posts", "DELETE")
	Enforcer.AddPolicy("admin", "/api/comments", "DELETE")
	Enforcer.AddPolicy("editor", "/api/posts", "POST")
	Enforcer.AddPolicy("editor", "/api/posts", "PUT")
	Enforcer.AddPolicy("editor", "/api/posts", "DELETE")
	Enforcer.AddPolicy("editor", "/api/comments", "DELETE")
	Enforcer.AddPolicy("moderator", "/api/comments", "DELETE")
	Enforcer.AddPolicy("user", "/api/posts", "GET")
	Enforcer.AddPolicy("user", "/api/categories", "GET")
	Enforcer.AddPolicy("user", "/api/comments", "GET")
	Enforcer.AddPolicy("user", "/api/comments", "POST")
	Enforcer.AddPolicy("guest", "/api/posts", "GET")
	Enforcer.AddPolicy("guest", "/api/categories", "GET")
	return nil
}
