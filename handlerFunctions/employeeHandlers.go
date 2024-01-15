package handlerFunctions

import (
	"Task_04/controllerFunctions"
	"Task_04/sharedpackage"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func CreateEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	var newEmployee sharedpackage.Employee
	err := json.NewDecoder(r.Body).Decode(&newEmployee)
	if err != nil {
		log.Printf("CreateEmployeeHandler ERROR: Error parsing request body: %v", err)
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	if newEmployee.FirstName == "" {
		http.Error(w, "Please provide Firstname of employee  (firstName)", http.StatusBadRequest)
		log.Println("CreateEmployeeHandler ERROR: Provide firstname (firstName)")
		return
	}
	if newEmployee.LastName == "" {
		http.Error(w, "Please provide Lastname of employee  (lastName)", http.StatusBadRequest)
		log.Println("CreateEmployeeHandler ERROR: Provide lastname (lastName)")
		return
	}
	if newEmployee.Email == "" {
		http.Error(w, "Please provide Email of employee  (email)", http.StatusBadRequest)
		log.Println("CreateEmployeeHandler ERROR: Provide Email (email)")
		return
	}
	if newEmployee.Password == "" {
		http.Error(w, "Please provide password for employee  (password)", http.StatusBadRequest)
		log.Println("CreateEmployeeHandler ERROR: Provide password (passowrd)")
		return
	}
	if newEmployee.Role == "HOD" {
		http.Error(w, "Cannot assign HOD role.", http.StatusBadRequest)
		log.Println("CreateEmployee ERROR: Cannot assign HOD role.")
		return
	}
	if newEmployee.Role == "admin" {
		http.Error(w, "Cannot assign admin role.", http.StatusBadRequest)
		log.Println("CreateEmployee ERROR: Cannot assign admin role.")
		return
	}
	if newEmployee.Role == "Lead" {
		http.Error(w, "Cannot assign Lead role.", http.StatusBadRequest)
		log.Println("CreateEmployee ERROR: Cannot assign Lead role.")
		return
	}

	log.Printf("CreateEmployeeHandler INFO: New employee data received: %+v", newEmployee)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newEmployee.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("CreateEmployeeHandler ERROR: error while hashing password", err)
	}

	newEmployee.Password = string(hashedPassword)

	// Call AddEmployee function to store the new employee in Firestore
	data, err := controllerFunctions.CreateEmployee(newEmployee)
	if err != nil {
		log.Printf("CreateEmployeeHandler ERROR: Error adding employee to Firestore: %v", err)
		http.Error(w, "Error adding employee to Firestore", http.StatusInternalServerError)
		return
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal employee data to JSON: %v", err), http.StatusInternalServerError)
		log.Printf("CreateEmployeeHandler ERROR: Failed to marshal employee data to JSON: %v", err)
		return
	}

	// Send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

}

func DeleteEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	employeeID, ok := vars["empID"]
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		log.Println("DeleteEmployeeHandler WARN: Invalid URL")
		return
	}
	log.Printf("DeleteEmployeeHandler INFO: Request received to delete employee with ID: %s", employeeID)

	data, err := controllerFunctions.DeleteEmployee(employeeID)
	if err != nil {
		http.Error(w, "Failed to delete employee", http.StatusInternalServerError)
		log.Printf("DeleteEmployeeHandler ERROR: Failed to delete employee with ID %s: %v", employeeID, err)
		return
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal employee data to JSON: %v", err), http.StatusInternalServerError)
		log.Printf("DeleteEmployeeHandler ERROR: Failed to marshal employee data to JSON: %v", err)
		return
	}

	// Send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func UpdateEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	employeeID, ok := vars["empID"]
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		log.Println("UpdateEmployeeHandler WARN: Invalid URL")
		return
	}
	log.Printf("UpdateEmployeeHandler INFO: Request received to delete employee with ID: %s", employeeID)

	var updateEmp sharedpackage.Employee
	err := json.NewDecoder(r.Body).Decode(&updateEmp)
	if err != nil {
		log.Printf("CreateEmployeeHandler ERROR: Error parsing request body: %v", err)
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: UpdateEmployeeHandler - Decoded request body fields: %+v", updateEmp)

	if updateEmp.Role == "HOD" {
		http.Error(w, "Cannot assign HOD role.", http.StatusBadRequest)
		log.Println("UpdateEmployeeHandler ERROR: Cannot assign HOD role.")
		return
	}
	if updateEmp.Role == "admin" {
		http.Error(w, "Cannot assign admin role.", http.StatusBadRequest)
		log.Println("UpdateEmployeeHandler ERROR: Cannot assign admin role.")
		return
	}
	if updateEmp.Role == "Lead" {
		http.Error(w, "Cannot assign Lead role.", http.StatusBadRequest)
		log.Println("UpdateEmployeeHandler ERROR: Cannot assign Lead role.")
		return
	}
	if updateEmp.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updateEmp.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("UpdateEmployeeHandler ERROR: error while hashing password", err)
		}

		updateEmp.Password = string(hashedPassword)
	}

	// Add the department and get the data
	data, err := controllerFunctions.UpdateEmployee(employeeID, updateEmp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update employee: %v", err), http.StatusInternalServerError)
		log.Printf("UpdateEmployeeHandler ERROR: Failed to update employee: %v", err)
		return
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal updated employe data to JSON: %v", err), http.StatusInternalServerError)
		log.Printf("UpdateEmployeeHandler ERROR: Failed to marshal updated employee data to JSON: %v", err)
		return
	}

	// Send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

}

func ListEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve all employees
	employees, err := controllerFunctions.ListEmployee()
	if err != nil {
		log.Printf("ERROR: Failed to get employees: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Printf("INFO: Retrieved %d employees", len(employees))

	// Convert employees to JSON format
	employeesJSON, err := json.Marshal(employees)
	if err != nil {
		log.Printf("ERROR: Failed to marshal employees to JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Printf("INFO: Successfully marshaled employees to JSON")

	w.Header().Set("Content-Type", "application/json")

	// Write JSON response
	if _, err := w.Write(employeesJSON); err != nil {
		log.Printf("ERROR: Failed to send employees JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Println("INFO: Sent employees JSON response")
}
