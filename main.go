package main

import (
	"Task_04/controllerFunctions"
	"Task_04/handlerFunctions"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	controllerFunctions.InitializeFirestore()
	r := mux.NewRouter()
	r.HandleFunc("/assignRole/{id}", handlerFunctions.AssignIAMRoleHandler).Methods("POST")
	r.HandleFunc("/deleteMember/{id}", handlerFunctions.RemoveMemberHandler).Methods("DELETE")
	r.HandleFunc("/createCustomRole", handlerFunctions.CreateCustomRoleHandler).Methods("POST")
	r.HandleFunc("/deleteCustomRole", handlerFunctions.DeleteCustomRoleHandler).Methods("DELETE")
	r.HandleFunc("/undeleteCustomRole", handlerFunctions.UndeleteCustomRoleHandler).Methods("PUT")
	r.HandleFunc("/listCustomRoles/{projectID}", handlerFunctions.ListCustomRolesHandler).Methods("GET")
	r.HandleFunc("/updateCustomRole", handlerFunctions.UpdateCustomRolesHandler).Methods("PATCH")
	r.HandleFunc("/iamRoles/{empID}/removeRoles", handlerFunctions.RemoveIAMRolesHandler).Methods("PATCH")

	//Department Level
	r.HandleFunc("/departments/create", handlerFunctions.CreateDepartmentHandler).Methods("POST")
	r.HandleFunc("/departments/{dept_id}/delete", handlerFunctions.DeleteDepartmentHandler).Methods("DELETE")
	r.HandleFunc("/departments/{dept_id}/update", handlerFunctions.UpadateDepartmentHandler).Methods("PATCH")
	r.HandleFunc("/departments",handlerFunctions.ListDepartmentsHandler).Methods("GET")

	//Employee Level
	r.HandleFunc("/employees/create", handlerFunctions.CreateEmployeeHandler).Methods("POST")
	r.HandleFunc("/employees/{empID}/add",handlerFunctions.DeleteEmployeeHandler).Methods("POST")
	r.HandleFunc("/employees/{empID}/delete",handlerFunctions.DeleteEmployeeHandler).Methods("DELETE")
	r.HandleFunc("/employees/{empID}/update",handlerFunctions.UpdateEmployeeHandler).Methods("PATCH")
	r.HandleFunc("/employees",handlerFunctions.ListEmployeeHandler).Methods("GET")

	//Team Level
	r.HandleFunc("/teams/create",handlerFunctions.CreateTeamHandler).Methods("POST")
	r.HandleFunc("/teams/{teamID}/delete",handlerFunctions.DeleteTeamHandler).Methods("DELETE")
	r.HandleFunc("/teams/{teamID}/update",handlerFunctions.UpdateTeamHandler).Methods("PATCH")
	r.HandleFunc("/teams",handlerFunctions.ListTeamHandler).Methods("GET")

	// Start the HTTP server using the Gorilla Mux router
	http.Handle("/", r)

	// Start the server on port 8080
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
