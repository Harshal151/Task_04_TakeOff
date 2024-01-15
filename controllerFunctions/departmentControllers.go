package controllerFunctions

import (
	"Task_04/sharedpackage"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// generateIncrementingDocID generates an incrementing document ID for the given collection.
func generateIncrementingDocID(collection *firestore.CollectionRef) (string, error) {
	// Query the collection to find the highest existing document ID
	iter := collection.Documents(context.Background())
	defer iter.Stop()

	existingIDs := make(map[string]bool)
	highestID := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", err
		}

		// Extract the numeric part of the document ID
		idStr := doc.Ref.ID[len("dept_"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue // Ignore non-numeric IDs
		}

		// Update the highestID if the current ID is greater
		if id > highestID {
			highestID = id
		}

		// Store existing IDs in a map
		existingIDs[doc.Ref.ID] = true
	}

	// Generate a new incrementing document ID
	for {
		newID := highestID + 1
		newIDStr := fmt.Sprintf("dept_%d", newID)

		// Check if the generated ID already exists
		if !existingIDs[newIDStr] {
			return newIDStr, nil
		}

		highestID++
	}

	// This should not be reached
}

// AddDepartment adds a new department document to the Firestore "departments" collection.
func AddDepartment(departmentName string, roles []string, headID string) (map[string]interface{}, error) {
	ctx := context.Background()
	currentTime := time.Now()
	formattedTime := currentTime.Format("Mon, 02 Jan 2006 15:04:05 MST")
	// Check if Firestore client is initialized
	if FirestoreClient == nil {
		log.Println("ERROR: Firestore client not initialized")
		return nil, fmt.Errorf("Firestore client not initialized")
	}

	// Specify the path to the "departments" collection
	departmentsCollection := FirestoreClient.Collection("departments")
	employeeCollection := FirestoreClient.Collection("employees")

	// Generate an incrementing document ID
	newDocID, err := generateIncrementingDocID(departmentsCollection)
	if err != nil {
		log.Printf("ERROR: Unable to generate a unique document ID: %v", err)
		return nil, fmt.Errorf("Unable to generate a unique document ID: %v", err)
	}

	// Create a map representing the department data
	departmentData := map[string]interface{}{
		"departmentName": departmentName,
		"iamRoles":       roles,
		"headID":         headID,
		"createdTime":    formattedTime,
		"updatedTime":    "",
	}

	// Add the department data to the "departments" collection with the generated document ID
	_, err = departmentsCollection.Doc(newDocID).Set(ctx, departmentData)
	if err != nil {
		log.Printf("ERROR: Failed to add department document: %v", err)
		return nil, fmt.Errorf("Failed to add department document: %v", err)
	}

	AssignIAMRole(newDocID, "", headID, roles, "HOD")

	// Set the document data for the employee with merge
	_, err = employeeCollection.Doc(headID).Set(ctx, map[string]interface{}{
		"role":         "HOD",
		"departmentID": newDocID,
		"teamIDs":      []string{},
	}, firestore.MergeAll)

	if err != nil {
		log.Printf("ERROR: Failed to set employee document data: %v", err)
		return nil, fmt.Errorf("Failed to set employee document data: %v", err)
	}

	log.Printf("INFO: Department document with ID %s and employee document updated successfully", newDocID)

	// Retrieve and return the department document data
	docSnapshot, err := departmentsCollection.Doc(newDocID).Get(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve department document: %v", err)
		return nil, fmt.Errorf("Failed to retrieve department document: %v", err)
	}

	// Check if the document exists
	if !docSnapshot.Exists() {
		log.Printf("ERROR: Document with ID %s does not exist", newDocID)
		return nil, fmt.Errorf("Document with ID %s does not exist", newDocID)
	}

	// Get the document data
	documentData := docSnapshot.Data()

	log.Printf("INFO: Retrieved department document with ID: %s", newDocID)
	return documentData, nil
}

func deleteTeams(deptID string) error {
	ctx := context.Background()
	teamCollection := FirestoreClient.Collection("teams")

	// Step 1: Create a query to get all documents in the teams collection where deptID matches
	teamQuery := teamCollection.Where("departmentID", "==", deptID)

	// Step 2: Get all team documents
	teamDocs, err := teamQuery.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("ERROR: Failed to get team documents: %v", err)
		return fmt.Errorf("Failed to get team documents: %v", err)
	}

	// Step 3: Loop through the team documents
	for _, teamDoc := range teamDocs {
		// Step 4: Extract the leadID field from the team document
		leadID, ok := teamDoc.Data()["leadID"].(string)
		if !ok {
			log.Printf("ERROR: Field 'leadID' not found or not a string in team document with ID: %s", teamDoc.Ref.ID)
			// Continue to the next team document if leadID is not found or not a string
			continue
		}

		// Step 5: Remove function for specific logic with leadID
		err := RemoveMember(leadID)
		if err != nil {
			log.Printf("ERROR: Failed to remove employee with leadID %s: %v", leadID, err)
			// Handle the error as needed, e.g., log or return an error
			continue
		}

		// Step 6: Delete the team document
		if _, err := teamDoc.Ref.Delete(ctx); err != nil {
			log.Printf("ERROR: Failed to delete team document with ID %s: %v", teamDoc.Ref.ID, err)
			// Handle the error as needed, e.g., log or return an error
			continue
		}

		// Step 7: Print a success message for the deleted team document and removed employee
		fmt.Printf("Deleted team document with ID: %s, and removed employee with leadID: %s\n", teamDoc.Ref.ID, leadID)
	}

	// Step 8: Return nil if the function completes without errors
	return nil
}

func removeElement(slice []string, elementToRemove string) []string {
	// Step 1: Find the index of the element to remove
	indexToRemove := -1
	for i, element := range slice {
		if element == elementToRemove {
			indexToRemove = i
			break
		}
	}

	// Step 2: If the element is found, remove it from the slice
	if indexToRemove != -1 {
		// Using slicing to remove the element
		slice = append(slice[:indexToRemove], slice[indexToRemove+1:]...)
	}

	return slice
}

func removeEmployees(deptID string, headID string) error {

	ctx := context.Background()
	employeeCollection1 := FirestoreClient.Collection("employees")
	var employeeQuery firestore.Query
	// Check if deptID is a departmentID or teamID
	if strings.HasPrefix(deptID, "dept_") {
		// Step 1: Create a query to get all documents in the employees collection where deptID matches
		employeeQuery = employeeCollection1.Where("departmentID", "==", deptID)
	} else if strings.HasPrefix(deptID, "team_") {
		// Step 1: Create a query to get all documents in the employees collection where deptID matches
		employeeQuery = employeeCollection1.Where("teamIDs", "array-contains", deptID)
	} else {
		// Handle invalid deptID
		return fmt.Errorf("Invalid deptID: %s", deptID)
	}

	// Step 2: Remove function for specific logic with headID
	if err := RemoveMember(headID); err != nil {
		log.Printf("ERROR: Failed to remove employee with leadID %s: %v", headID, err)
		// Handle the error as needed, e.g., log or return an error
		return fmt.Errorf("Failed to remove employee with leadID %s: %v", headID, err)
	}
	log.Printf("INFO: Removed employee successfully with leadID: %s", headID)

	// Step 3: Get all employee documents
	employeeDocs, err := employeeQuery.Documents(ctx).GetAll()
	if err != nil {
		log.Fatalf("ERROR: Failed to get employee documents: %v", err)
		return fmt.Errorf("Failed to get employee documents: %v", err)
	}

	// Step 4: Loop through the employee documents
	for _, employeeDoc := range employeeDocs {
		var employee sharedpackage.Employee

		// Convert Firestore document data to the Employee struct
		if err := employeeDoc.DataTo(&employee); err != nil {
			log.Printf("ERROR: Error converting document data: %v", err)
			return fmt.Errorf("Error converting document data: %v", err)
		}

		if strings.HasPrefix(deptID, "dept_") {
			// Step 5: Create a map with the modified fields
			updateData := map[string]interface{}{
				"departmentID": "",         // Set departmentID to an empty string
				"teamIDs":      []string{}, // Set teamID to an empty slice
				"role":         "",
			}
			// Step 6: Update the employee document with the modified fields
			if _, err := employeeDoc.Ref.Set(ctx, updateData, firestore.MergeAll); err != nil {
				log.Printf("ERROR: Failed to update employee document with ID %s: %v", employeeDoc.Ref.ID, err)
				// Handle the error as needed
				continue
			}
			// Step 7: Print a success message for the updated employee document
			log.Printf("INFO: Updated employee document with ID: %s", employeeDoc.Ref.ID)
			return nil
		} else {
			for i := 0; i < len(employee.TeamIDs); i++ {
				if employee.TeamIDs[i] == deptID {
					employee.TeamIDs = removeElement(employee.TeamIDs, deptID)
				}
			}
			if len(employee.TeamIDs) == 0 {
				employee.DeptID = ""
			}
			// Step 5: Create a map with the modified fields
			updateData := map[string]interface{}{
				"departmentID": employee.DeptID,
				"teamIDs":      employee.TeamIDs, // Set teamID to an empty slice
				"role":         "",
			}
			// Step 6: Update the employee document with the modified fields
			if _, err := employeeDoc.Ref.Set(ctx, updateData, firestore.MergeAll); err != nil {
				log.Printf("ERROR: Failed to update employee document with ID %s: %v", employeeDoc.Ref.ID, err)
				// Handle the error as needed
				continue
			}
			return nil
		}

	}

	// Step 8: Return nil if the function completes without errors
	return nil
}

func DeleteDepartment(docID string) error {
	ctx := context.Background()

	// Step 1: Get the departments collection
	departmentsCollection := FirestoreClient.Collection("departments")

	// Step 2: Delete teams associated with the department
	if err := deleteTeams(docID); err != nil {
		log.Printf("ERROR: Failed to delete teams: %v", err)
		// You can choose to return the error or handle it based on your requirements
		return fmt.Errorf("Failed to delete teams: %v", err)
	}

	// Step 3: Specify the path to the document
	docRef := departmentsCollection.Doc(docID)

	// Step 4: Check if the document exists
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		log.Printf("Error getting document: %v", err)
		return fmt.Errorf("Error getting document: %v", err)
	}

	if !docSnapshot.Exists() {
		log.Printf("Document with ID %s does not exist", docID)
		return fmt.Errorf("Document with ID %s does not exist", docID)
	}

	// Step 5: Extract the value of the "headID" field
	headID, ok := docSnapshot.Data()["headID"].(string)
	if !ok {
		log.Printf("Field 'headID' not found or not a string")
		return fmt.Errorf("Field 'headID' not found or not a string")
	}

	log.Printf("INFO: Extracted 'headID' from document: %s", headID)

	// Step 6: Handle the error from removeAccess
	if err := removeEmployees(docID, headID); err != nil {
		log.Printf("ERROR: Failed to remove access: %v", err)
		// You can choose to return the error or handle it based on your requirements
		return fmt.Errorf("Failed to remove access: %v", err)
	}

	// Step 7: Delete the document
	_, err = docRef.Delete(ctx)
	if err != nil {
		log.Printf("ERROR: Error deleting document: %v", err)
		return fmt.Errorf("Error deleting document: %v", err)
	}

	log.Printf("INFO: Deleted document with ID %s successfully", docID)

	// Step 8: Return nil if the function completes without errors
	return nil
}

func UpdateDepartment(deptID string, dept sharedpackage.Department) (*sharedpackage.Department, error) {
	ctx := context.Background()
	departmentCollection := FirestoreClient.Collection("departments")
	employeeCollection := FirestoreClient.Collection("employees")

	docDepartment := departmentCollection.Doc(deptID)
	// Step 4: Check if the document exists
	docSnapshotTeam, err := docDepartment.Get(ctx)
	if err != nil {
		log.Printf("Error getting document: %v", err)
		return nil, fmt.Errorf("Error getting document: %v", err)
	}

	if !docSnapshotTeam.Exists() {
		log.Printf("Document with ID %s does not exist", deptID)
		return nil, fmt.Errorf("Document with ID %s does not exist", deptID)
	}

	var departmentData sharedpackage.Department
	// Convert Firestore document data to the Employee struct
	if err := docSnapshotTeam.DataTo(&departmentData); err != nil {
		log.Printf("ERROR: Error converting document data: %v", err)
		return nil, fmt.Errorf("Error converting document data: %v", err)
	}

	// Step 5: Extract the value of the "headID" field
	headID := departmentData.HeadID

	docEmployee := employeeCollection.Doc(headID)
	// Step 4: Check if the document exists
	docSnapshotEmployee, err := docEmployee.Get(ctx)
	if err != nil {
		log.Printf("Error getting document: %v", err)
		return nil, fmt.Errorf("Error getting document: %v", err)
	}

	if !docSnapshotEmployee.Exists() {
		log.Printf("Document with ID %s does not exist", deptID)
		return nil, fmt.Errorf("Document with ID %s does not exist", deptID)
	}
	var employee sharedpackage.Employee
	// Convert Firestore document data to the Employee struct
	if err := docSnapshotEmployee.DataTo(&employee); err != nil {
		log.Printf("ERROR: Error converting document data: %v", err)
		return nil, fmt.Errorf("Error converting document data: %v", err)
	}

	if dept.HeadID != "" {
		docEmployee2 := employeeCollection.Doc(dept.HeadID)

		docSnapshotEmployee2, err := docEmployee2.Get(ctx)
		if err != nil {
			log.Printf("Error getting document: %v", err)
			return nil, fmt.Errorf("Error getting document: %v", err)
		}

		if !docSnapshotEmployee2.Exists() {
			log.Printf("Document with ID %s does not exist", dept.HeadID)
			return nil, fmt.Errorf("Document with ID %s does not exist", dept.HeadID)
		}
		var employee1 sharedpackage.Employee
		// Convert Firestore document data to the Employee struct
		if err := docSnapshotEmployee2.DataTo(&employee1); err != nil {
			log.Printf("ERROR: Error converting document data: %v", err)
			return nil, fmt.Errorf("Error converting document data: %v", err)
		}
		employee1.DeptID = employee.DeptID
		employee1.Role = employee.Role
		employee1.IAMRoles = employee.IAMRoles
		data, err := AssignIAMRole(employee1.DeptID, "", dept.HeadID, employee1.IAMRoles[deptID], employee1.Role)
		if err != nil {
			log.Printf("ERROR: Error while assigning IAM role: %v", err)
			return nil, fmt.Errorf("Error assigning IAM role: %v", err)
		}
		log.Printf("INFO: Assigned IAM roles to %v: %v", dept.HeadID, data)
		RemoveMember(headID)
		employee.DeptID = ""
		// Update the document with the modified field
		if _, err := docEmployee2.Set(ctx, employee1); err != nil {
			log.Printf("ERROR: Error updating document: %v", err)
			return nil, fmt.Errorf("Error updating document: %v", err)
		}
	} else {
		dept.HeadID = departmentData.HeadID
	}
	if len(dept.IAMRoles) != 0 {
		employee.IAMRoles[deptID] = mergeSlices(dept.IAMRoles, employee.IAMRoles[deptID])
		role := employee.Role
		deptID := employee.DeptID
		RemoveMember(headID)
		data, err := AssignIAMRole(deptID, deptID, dept.HeadID, employee.IAMRoles[deptID], role)
		if err != nil {
			log.Printf("ERROR: Error while assigning IAM role: %v", err)
			return nil, fmt.Errorf("Error assigning IAM role: %v", err)
		}
		log.Printf("INFO: Assigned IAM roles to %v: %v", dept.HeadID, data)
		employee.Role = role
		employee.DeptID = deptID
		// Update the document with the modified field
		if _, err := docEmployee.Set(ctx, employee); err != nil {
			log.Printf("ERROR: Error updating document: %v", err)
			return nil, fmt.Errorf("Error updating document: %v", err)
		}
		dept.IAMRoles = employee.IAMRoles[deptID]
	} else {
		dept.IAMRoles = departmentData.IAMRoles
	}
	if dept.DepartmentName == "" {
		dept.DepartmentName = departmentData.DepartmentName
	}

	currentTime := time.Now()
	updatedTime := currentTime.Format("Mon, 02 Jan 2006 15:04:05 MST")
	dept.UpdatedTime = updatedTime
	dept.CreatedTime = departmentData.CreatedTime

	// Update the Firestore document with merged data
	_, err = docDepartment.Set(ctx, dept)
	if err != nil {
		log.Printf("Error updating data in document with ID %s: %v", deptID, err)
		return nil, fmt.Errorf("Error updating data in document: %v", err)
	}

	log.Printf("Employee with ID %s updated successfully", deptID)
	return &dept, nil
}

func ListDepartments() (*[]sharedpackage.Department, error) {
    ctx := context.Background()

    iter := FirestoreClient.Collection("departments").Documents(ctx)

    var departments []sharedpackage.Department

    for {
        doc, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            log.Printf("ERROR: Error iterating over documents: %v", err)
            return nil, err
        }

        var departmentData sharedpackage.Department
        if err := doc.DataTo(&departmentData); err != nil {
            log.Printf("ERROR: Error converting document data: %v", err)
            return nil, err
        }

        departments = append(departments, departmentData)
        log.Printf("INFO: Processed department.")
    }

    return &departments, nil
}
