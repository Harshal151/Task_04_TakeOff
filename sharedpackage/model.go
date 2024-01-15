package sharedpackage

import "github.com/dgrijalva/jwt-go"

type Employee struct {
	FirstName string              `firestore:"firstName" json:"firstName"`
	LastName  string              `firestore:"lastName" json:"lastName"`
	Email     string              `firestore:"mailID" json:"mailID"`
	Password  string              `firestore:"password" json:"password"`
	Role      string              `firestore:"role" json:"role"`
	IAMRoles  map[string][]string `firestore:"iamRoles" json:"iamRoles"`
	TeamIDs   []string            `firestore:"teamIDs" json:"teamIDs"`
	DeptID    string              `firestore:"departmentID" json:"departmentID"`
}

type Department struct {
	DepartmentName string   `firestore:"departmentName" json:"departmentName"`
	IAMRoles       []string `firestore:"iamRoles" json:"iamRoles"`
	HeadID         string   `firestore:"headID" json:"headID"`
	CreatedTime    string   `firestore:"createdTime" json:"createdTime"`
	UpdatedTime    string   `firestore:"updatedTime" json:"updatedTime"`
}

type Team struct {
	TeamName     string   `firestore:"teamName" json:"teamName"`
	IAMRoles     []string `firestore:"iamRoles" json:"iamRoles"`
	LeadID       string   `firestore:"leadID" json:"leadID"`
	DepartmentID string   `firestore:"departmentID" json:"departmentID"`
	CreatedTime  string   `firestore:"createdTime" json:"createdTime"`
	UpdatedTime  string   `firestore:"updatedTime" json:"updatedTime"`
}

type AssignRole struct {
	TeamID   string   `json:"teamID"`
	DeptID   string   `json:"departmentID"`
	IAMRoles []string `json:"iamRoles"`
}

type CustomRole struct {
	Name      string   `json:"name"`
	RoleID    string   `json:"roleID"`
	Title     string   `json:"title"`
	Stage     string   `json:"stage"`
	Desc      string   `json:"description"`
	Perm      []string `json:"permissions"`
	ProjectID string   `json:"projectID"`
}

type UpdateCR struct {
	Title string   `json:"title"`
	Stage string   `json:"stage"`
	Desc  string   `json:"description"`
	Perm  []string `json:"permissions"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Binding struct {
	Role    string   // Role name
	Members []string // Members assigned to the role
}

type Policy struct {
	Bindings []Binding // List of role bindings
}
