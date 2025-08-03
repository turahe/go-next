package models

import "github.com/google/uuid"

type Organization struct {
	BaseModelWithOrdering
	Name        string           `json:"name" gorm:"not null;size:100" validate:"required,min=1,max=100"`
	Slug        string           `json:"slug" gorm:"uniqueIndex;not null;size:100" validate:"required,min=1,max=100"`
	Description string           `json:"description" gorm:"size:500"`
	Code        string           `json:"code" gorm:"uniqueIndex;not null;size:100" validate:"required,min=1,max=100"`
	Type        OrganizationType `json:"type" gorm:"not null;size:100" validate:"required,min=1,max=100"`
	ParentID    uuid.UUID        `json:"parent_id" gorm:"type:uuid;index"`
	Children    []Organization   `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Users       []User           `json:"users,omitempty" gorm:"many2many:organization_users;constraint:OnDelete:CASCADE"`
	Media       []Media          `json:"media,omitempty" gorm:"many2many:mediables;constraint:OnDelete:CASCADE"`
}

func (Organization) TableName() string {
	return "organizations"
}

// OrganizationType represents the allowed types for Organization.Type
type OrganizationType string

const (
	OrganizationTypeCompany           OrganizationType = "COMPANY"
	OrganizationTypeCompanyHolding    OrganizationType = "COMPANY_HOLDING"
	OrganizationTypeCompanySubsidiary OrganizationType = "COMPANY_SUBSIDIARY"
	OrganizationTypeOutlet            OrganizationType = "OUTLET"
	OrganizationTypeStore             OrganizationType = "STORE"
	OrganizationTypeDepartment        OrganizationType = "DEPARTMENT"
	OrganizationTypeSubDepartment     OrganizationType = "SUB_DEPARTMENT"
	OrganizationTypeDivision          OrganizationType = "DIVISION"
	OrganizationTypeSubDivision       OrganizationType = "SUB_DIVISION"
	OrganizationTypeDesignation       OrganizationType = "DESIGNATION"
	OrganizationTypeInstitution       OrganizationType = "INSTITUTION"
	OrganizationTypeCommunity         OrganizationType = "COMMUNITY"
	OrganizationTypeOrganization      OrganizationType = "ORGANIZATION"
	OrganizationTypeFoundation        OrganizationType = "FOUNDATION"
	OrganizationTypeBranchOffice      OrganizationType = "BRANCH_OFFICE"
	OrganizationTypeBranchOutlet      OrganizationType = "BRANCH_OUTLET"
	OrganizationTypeBranchStore       OrganizationType = "BRANCH_STORE"
	OrganizationTypeRegional          OrganizationType = "REGIONAL"
	OrganizationTypeFranchisee        OrganizationType = "FRANCHISEE"
	OrganizationTypePartner           OrganizationType = "PARTNER"
)
