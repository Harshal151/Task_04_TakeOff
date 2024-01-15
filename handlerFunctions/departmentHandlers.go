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

func UpadateDepartmentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	departmentID, ok := vars["dept_id"]
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		log.Println("UpadateDepartmentHandler WARN: Invalid URL")
		return
	}
	log.Printf("UpadateDepartmentHandler INFO: Request received to delete employee with ID: %s", departmentID)

	var updateDept sharedpackage.Department
	err := json.NewDecoder(r.Body).Decode(&updateDept)
	if err != nil {
		log.Printf("UpadateDepartmentHandler ERROR: Error parsing request body: %v", err)
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: UpadateDepartmentHandler - Decoded request body fields: %+v", updateDept)

	// Add the department and get the data
	data, err := controllerFunctions.UpdateDepartment(departmentID, updateDept)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update employee: %v", err), http.StatusInternalServerError)
		log.Printf("UpadateDepartmentHandler ERROR: Failed to update employee: %v", err)
		return
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal updated employe data to JSON: %v", err), http.StatusInternalServerError)
		log.Printf("UpadateDepartmentHandler ERROR: Failed to marshal updated employee data to JSON: %v", err)
		return
	}

	// Send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

}

func ListDepartmentsHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve all departments
	departments, err := controllerFunctions.ListDepartments()
	if err != nil {
		log.Printf("ERROR: Failed to get departments: %v", err)
		http.Error(w, "Failed to retrieve departments", http.StatusInternalServerError)
		return
	}
	log.Printf("INFO: Retrieved %d departments", len(*departments))

	// Convert departments to JSON format
	departmentsJSON, err := json.Marshal(departments)
	if err != nil {
		log.Printf("ERROR: Failed to marshal departments to JSON: %v", err)
		http.Error(w, "Failed to convert departments to JSON", http.StatusInternalServerError)
		return
	}
	log.Printf("INFO: Successfully marshaled departments to JSON")

	w.Header().Set("Content-Type", "application/json")

	// Write JSON response
	if _, err := w.Write(departmentsJSON); err != nil {
		log.Printf("ERROR: Failed to send departments JSON response: %v", err)
		http.Error(w, "Failed to send departments JSON response", http.StatusInternalServerError)
		return
	}
	log.Println("INFO: Sent departments JSON response")
}
