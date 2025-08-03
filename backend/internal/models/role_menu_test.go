package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRoleMenuModel(t *testing.T) {
	roleMenu := RoleMenu{
		RoleID: uuid.New(),
		MenuID: uuid.New(),
	}

	assert.NotNil(t, roleMenu)
	assert.Equal(t, "role_menus", roleMenu.TableName())
}
