package handlerFunctions

import (
	"Task_04/controllerFunctions"
	"Task_04/iamRole"
	"Task_04/sharedpackage"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

var projectID = "ems-web-application-409305"
var jwtKey = []byte("secret_key")

// Login handles the login functionality.
func Login(w http.ResponseWriter, r *http.Request) {
	var credentials sharedpackage.Credentials
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	expectedPassword, err := controllerFunctions.LoginUser(credentials.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check if the password matches the expected password
	if credentials.Password != *expectedPassword {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(time.Minute * 5)

	claims := &sharedpackage.Claims{
		Username: credentials.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the JWT token as a cookie
	http.SetCookie(w,
		&http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
		})

	// Optionally, you can send a success response if needed
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login successful"))
}

// AssignIAMRoleHandler handles the HTTP request to assign IAM roles to an employee.
func AssignIAMRoleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	employeeIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		log.Println("WARN: Invalid URL in AssignIAMRoleHandler")
		return
	}

	log.Printf("INFO: Request received for employee with ID: %s", employeeIDStr)

	// Parse the request body to get the updated fields
	var request sharedpackage.AssignRole

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("ERROR: Error decoding request body: %v", err)
		return
	}

	log.Printf("INFO: AssignIAMRoleHandler - Decoded request body with updated fields: %+v", request)

	if request.DeptID == "" {
		http.Error(w, "Please provide departmentID", http.StatusBadRequest)
		log.Println("ERROR: Provide departmentID.")
		return
	}
	if request.IAMRoles == nil {
		http.Error(w, "Please provide iamRoles ", http.StatusBadRequest)
		log.Println("ERROR: Provide iamRoles .")
		return
	}
	if request.Role == "" {
		http.Error(w, "Please provide role", http.StatusBadRequest)
		log.Println("ERROR: Provide role.")
		return
	}
	if request.Role == "HOD" {
		http.Error(w, "HOD role cannot be assigned.", http.StatusBadRequest)
		log.Println("ERROR: HOD job cannot be assigned.")
		return
	}
	if request.Role == "Lead" {
		http.Error(w, "Lead role cannot be assigned.", http.StatusBadRequest)
		log.Println("ERROR: Lead job cannot be assigned.")
		return
	}
	if request.Role == "Admin" {
		http.Error(w, "Admin role cannot be assigned.", http.StatusBadRequest)
		log.Println("ERROR: Admin job cannot be assigned.")
		return
	}

	data, err := controllerFunctions.AssignIAMRole(request.DeptID, request.TeamID, employeeIDStr, request.IAMRoles, request.Role)
	if err != nil {
		http.Error(w, "Failed to assign IAM role", http.StatusInternalServerError)
		log.Printf("ERROR: Failed to assign IAM role: %v", err)
		return
	}

	// Convert struct to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("INFO: %s", jsonData)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	// Write JSON data as the response
	w.Write(jsonData)
}

// RemoveMemberHandler handles the HTTP request to remove an employee.
func RemoveMemberHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	employeeIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		log.Println("WARN: Invalid URL in RemoveMemberHandler")
		return
	}
	log.Printf("INFO: Request received for employee with ID: %s", employeeIDStr)

	err := controllerFunctions.RemoveMember(employeeIDStr)
	if err != nil {
		http.Error(w, "Failed to remove employee", http.StatusInternalServerError)
		log.Printf("ERROR: Failed to remove employee: %v", err)
		return
	}

	log.Printf("INFO: Employee with ID %s successfully removed", employeeIDStr)
	w.WriteHeader(http.StatusOK)
}

func CreateCustomRoleHandler(w http.ResponseWriter, r *http.Request) {
	var request sharedpackage.CustomRole

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("ERROR: Error decoding request body: %v", err)
		return
	}

	log.Printf("INFO: CreateCustomRoleHandler - Decoded request body fields: %+v", request)

	if request.Desc == "" || request.ProjectID == "" || request.Name == "" || request.RoleID == "" || request.Title == "" || request.Perm == nil {
		http.Error(w, "Please provide all fields", http.StatusBadRequest)
		log.Println("ERROR: Empty fields.")
		return
	}

	// Specify your projectID (replace "your-project-id" with your actual project ID)
	role, err := iamRole.CreateRole(log.Writer(), request.ProjectID, request.Name, request.Title, request.Desc, request.Stage, request.Perm)
	if err != nil {
		http.Error(w, "Error creating role", http.StatusInternalServerError)
		log.Printf("ERROR: Error creating role: %v", err)
		return
	}

	// Use the created role if needed
	log.Printf("INFO: Role created successfully: %v", role)

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Role created successfully: %v", role)
}

// DeleteCustomRoleHandler will delete custom role creates in roles.
func DeleteCustomRoleHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve query parameters from the request URL
	name := r.URL.Query().Get("name")
	projectID := r.URL.Query().Get("projectID")

	// Check if name or projectID is missing
	if name == "" || projectID == "" {
		http.Error(w, "Both 'name' and 'projectID' are required query parameters", http.StatusBadRequest)
		log.Println("ERROR: Missing 'name' or 'projectID' in the request.")
		return
	}

	err := iamRole.DeleteRole(log.Writer(), projectID, name)
	if err != nil {
		http.Error(w, "Error deleting role", http.StatusInternalServerError)
		log.Printf("ERROR: Error deleting role: %v", err)
		return
	}

	// Use the created role if needed
	log.Printf("INFO: Role deleted successfully: %v", name)

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Role deleted successfully: %v", name)
}

// UndeleteCustomerRoleHandler undeletes custom role which is previously deleted.
func UndeleteCustomRoleHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve query parameters from the request URL
	name := r.URL.Query().Get("name")
	projectID := r.URL.Query().Get("projectID")

	// Check if name or projectID is missing
	if name == "" || projectID == "" {
		http.Error(w, "Both 'name' and 'projectID' are required query parameters", http.StatusBadRequest)
		log.Println("ERROR: Missing 'name' or 'projectID' in the request.")
		return
	}

	err := iamRole.UndeleteRole(log.Writer(), projectID, name)
	if err != nil {
		http.Error(w, "Error undeleting role", http.StatusInternalServerError)
		log.Printf("ERROR: Error undeleting role: %v", err)
		return
	}

	log.Printf("INFO: Role undeleted successfully: %v", name)

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Role undeleted successfully: %v", name)
}

func ListCustomRolesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, ok := vars["projectID"]
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		log.Println("WARN: Invalid URL in ListRolesHandler")
		return
	}

	roleNames, err := iamRole.ListCustomRoles(log.Writer(), projectID)
	if err != nil {
		http.Error(w, "Error listing roles", http.StatusInternalServerError)
		log.Printf("ERROR: Error listing roles: %v", err)
		return
	}

	// Use the role names data if needed
	log.Printf("INFO: Roles listed successfully for project %s: %+v", projectID, roleNames)

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Roles listed successfully for project %s: %+v", projectID, roleNames)
}

func UpdateCustomRolesHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve query parameters from the request URL
	name := r.URL.Query().Get("name")
	projectID := r.URL.Query().Get("projectID")

	// Check if name or projectID is missing
	if name == "" || projectID == "" {
		http.Error(w, "Both 'name' and 'projectID' are required query parameters", http.StatusBadRequest)
		log.Println("ERROR: Missing 'name' or 'projectID' in the request.")
		return
	}

	var request sharedpackage.UpdateCR

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("ERROR: Error decoding request body: %v", err)
		return
	}

	log.Printf("INFO: UpdateCustomRolesHandler - Decoded request body fields: %+v", request)

	// Specify your projectID (replace "your-project-id" with your actual project ID)
	role, err := iamRole.UpdateCustomRole(log.Writer(), projectID, name, request.Title, request.Desc, request.Stage, request.Perm)
	if err != nil {
		http.Error(w, "Error creating role", http.StatusInternalServerError)
		log.Printf("ERROR: Error creating role: %v", err)
		return
	}

	// Convert role to JSON format
	roleJSON, err := json.Marshal(role)
	if err != nil {
		http.Error(w, "Error encoding role to JSON", http.StatusInternalServerError)
		log.Printf("ERROR: Error encoding role to JSON: %v", err)
		return
	}

	// Respond with JSON file
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(roleJSON)

	// Log success
	log.Printf("INFO: UpdateCustomRolesHandler - Role created successfully: %+v", role)
}

func RemoveIAMRolesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	employeeID, ok := vars["empID"]
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		log.Println("WARN: Invalid URL in RemoveIAMRolesHandler")
		return
	}

	log.Printf("INFO: Request received for employee with ID: %s", employeeID)

	var request sharedpackage.RemoveRoles

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("ERROR: Error decoding request body: %v", err)
		return
	}

	log.Printf("INFO: RemoveIAMRolesHandler - Decoded request body fields: %+v", request)

	// Specify your projectID (replace "your-project-id" with your actual project ID)
	data, err := controllerFunctions.RemoveIAMRoles(employeeID, request)
	if err != nil {
		http.Error(w, "Error creating role", http.StatusInternalServerError)
		log.Printf("ERROR: Error creating role: %v", err)
		return
	}
	// Convert role to JSON format
	roleJSON, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Error encoding document to JSON", http.StatusInternalServerError)
		log.Printf("ERROR: Error encoding document to JSON: %v", err)
		return
	}

	// Respond with JSON file
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(roleJSON)

	// Log success
	log.Printf("INFO: RemoveIAMRolesHandler - Role updated successfully: %+v", data)
}
