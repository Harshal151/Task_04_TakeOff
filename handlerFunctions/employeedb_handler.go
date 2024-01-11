package handlerFunctions

import (
	"Task_04/controllerFunctions"
	"Task_04/sharedpackage"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func CreateDepartmentHandler(w http.ResponseWriter, r *http.Request) {
	var department sharedpackage.Department
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&department); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("ERROR: Error decoding request body: %v", err)
		return
	}

	log.Printf("INFO: CreateCustomRoleHandler - Decoded request body fields: %+v", department)
	if department.DepartmentName == "" {
		http.Error(w, "Please provide Department name (departmentName)", http.StatusBadRequest)
		log.Println("ERROR: Provide Department name (departmentName)")
		return
	}
	if department.HeadID == "" {
		http.Error(w, "Please provide department head ID (headID)", http.StatusBadRequest)
		log.Println("ERROR: Please provide department head ID (headID)")
		return
	}
	if department.IAMRoles == nil {
		http.Error(w, "Please assign roles (roles)", http.StatusBadRequest)
		log.Println("ERROR: Please assign roles (roles)")
		return
	}

	// Add the department and get the data
	data, err := controllerFunctions.AddDepartment(department.DepartmentName, department.IAMRoles, department.HeadID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add department: %v", err), http.StatusInternalServerError)
		log.Printf("ERROR: Failed to add department: %v", err)
		return
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal department data to JSON: %v", err), http.StatusInternalServerError)
		log.Printf("ERROR: Failed to marshal department data to JSON: %v", err)
		return
	}

	// Send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}


func DeleteDepartmentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	departmentID, ok := vars["dept_id"]
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		log.Println("WARN: Invalid URL in DeleteDepartmentHandler")
		return
	}
	log.Printf("INFO: Request received to delete department with ID: %s", departmentID)

	err := controllerFunctions.DeleteDepartment(departmentID)
	if err != nil {
		http.Error(w, "Failed to delete department", http.StatusInternalServerError)
		log.Printf("ERROR: Failed to delete department with ID %s: %v", departmentID, err)
		return
	}

	// Respond with success status
	w.WriteHeader(http.StatusNoContent)
	
	log.Printf("INFO: Department with ID %s deleted successfully", departmentID)
}
