package controllerFunctions

import (
	"Task_04/iamRole"
	"Task_04/sharedpackage"
	"context"
	"fmt"
	"strconv"

	// "errors"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mergeMaps(mapA, mapB map[string][]string) map[string][]string {
	mergedMap := make(map[string][]string)

	// Iterate over mapA
	for key, values := range mapA {
		// Add key-value pair to the merged map
		mergedMap[key] = append(mergedMap[key], values...)
	}

	// Iterate over mapB
	for key, values := range mapB {
		// If key already exists, append values; otherwise, add new key-value pair
		if existingValues, ok := mergedMap[key]; ok {
			mergedMap[key] = append(existingValues, values...)
		} else {
			mergedMap[key] = append([]string{}, values...)
		}
	}

	return mergedMap
}

func mergeSlices(sliceA, sliceB []string) []string {
	// Create a map to store unique values from sliceB
	uniqueValues := make(map[string]bool)

	// Add values from sliceB to the map
	for _, value := range sliceB {
		uniqueValues[value] = true
	}

	// Iterate through sliceA and add non-duplicate values to the result slice
	var result []string
	for _, value := range sliceA {
		if _, exists := uniqueValues[value]; !exists {
			result = append(result, value)
		}
	}

	// Append remaining values from sliceB to the result slice
	for value := range uniqueValues {
		result = append(result, value)
	}

	return result
}

// generateIncrementingEmployeeID generates an incrementing document ID for the given collection.
func generateIncrementingEmployeeID(collection *firestore.CollectionRef) (string, error) {
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
		idStr := doc.Ref.ID[len("emp_"):]
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
		newIDStr := fmt.Sprintf("emp_%d", newID)

		// Check if the generated ID already exists
		if !existingIDs[newIDStr] {
			return newIDStr, nil
		}

		highestID++
	}

	// This should not be reached
}

func CreateEmployee(employee sharedpackage.Employee) (*sharedpackage.Employee, error) {
	ctx := context.Background()
	log.Printf("CreateEmployee INFO: Employee details received successfully.")

	// Reference to the "departments" collection
	departmentCollection := FirestoreClient.Collection("departments").Doc(employee.DeptID)
	// Reference to the "employees" collection
	employeeCollection := FirestoreClient.Collection("employees")

	// Check if the department document exists
	docSnapshot, err := departmentCollection.Get(ctx)
	if err != nil {
		if grpcStatus, ok := status.FromError(err); ok && grpcStatus.Code() == codes.NotFound {
			log.Printf("CreateEmployee ERROR: Department with ID %s not found: %v", employee.DeptID, err)
			return nil, fmt.Errorf("Department with ID %s not found", employee.DeptID)
		}
		log.Printf("CreateEmployee ERROR: Unable to retrieve department with ID %s: %v", employee.DeptID, err)
		return nil, err
	}
	log.Printf("CreateEmployee INFO: Department with ID %s found: %v", employee.DeptID, docSnapshot)

	// Check if the team documents exist
	for i := 0; i < len(employee.TeamIDs); i++ {
		teamsCollection := FirestoreClient.Collection("teams").Doc(employee.TeamIDs[i])
		docSnapshot1, err := teamsCollection.Get(ctx)
		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus.Code() == codes.NotFound {
				log.Printf("CreateEmployee ERROR: Team with ID %s not found: %v", employee.TeamIDs[i], err)
				return nil, fmt.Errorf("Team with ID %s not found", employee.TeamIDs[i])
			}
			log.Printf("CreateEmployee ERROR: Unable to retrieve team with ID %s: %v", employee.TeamIDs[i], err)
			return nil, err
		}
		log.Printf("CreateEmployee INFO: Team with ID %s found: %v", employee.TeamIDs[i], docSnapshot1)
	}

	// Generate an incrementing document ID
	newDocID, err := generateIncrementingEmployeeID(employeeCollection)
	if err != nil {
		log.Printf("ERROR: Unable to generate a unique document ID: %v", err)
		return nil, fmt.Errorf("Unable to generate a unique document ID: %v", err)
	}

	// Add the employee data to the "employees" collection with the generated document ID
	_, err = employeeCollection.Doc(newDocID).Set(ctx, employee)
	if err != nil {
		log.Printf("ERROR: Failed to add employee document: %v", err)
		return nil, fmt.Errorf("Failed to add employee document: %v", err)
	}

	// Assign IAM roles based on teams (replace with your actual implementation)
	for i := 0; i < len(employee.TeamIDs); i++ {
		AssignIAMRole(employee.DeptID, employee.TeamIDs[i], newDocID, employee.IAMRoles[employee.TeamIDs[i]], employee.Role)
	}

	log.Printf("CreateEmployee INFO: Employee added to Firestore: %+v", employee)
	// Return the employee or any other relevant information
	return &employee, nil
}

func DeleteEmployee(empID string) (*sharedpackage.Employee, error) {
	ctx := context.Background()
	docRef := FirestoreClient.Collection("employees").Doc(empID)

	// Check if the document exists before attempting to delete
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
			log.Printf("Document with ID %s not found: %v", empID, err)
			return nil, fmt.Errorf("Document with ID %s not found", empID)
		}
		log.Printf("Error retrieving document with ID %s: %v", empID, err)
		return nil, fmt.Errorf("Error retrieving document: %v", err)
	}

	var employee sharedpackage.Employee
	if err := docSnapshot.DataTo(&employee); err != nil {
		log.Printf("Error decoding document: %v", err)
		return nil, fmt.Errorf("Error decoding document: %v", err)
	}

	err = RemoveMember(empID)
	if err != nil {
		log.Printf("ERROR: Failed to remove employee with leadID %s: %v", empID, err)
		// Handle the error as needed, e.g., log or return an error
		return nil, fmt.Errorf("Failed to remove employee with leadID %s: %v", empID, err)
	}
	log.Printf("INFO: Removed employee successfully with leadID: %s", empID)

	// Document exists, proceed with deletion
	_, err = docRef.Delete(ctx)
	if err != nil {
		log.Printf("Error deleting document with ID %s: %v", empID, err)
		return nil, fmt.Errorf("Error deleting document: %v", err)
	}

	log.Printf("Document with ID %s deleted successfully", empID)
	return &employee, nil
}

func UpdateEmployee(empID string, updatedEmp sharedpackage.Employee) (*sharedpackage.Employee, error) {
	ctx := context.Background()
	var employee sharedpackage.Employee
	docRef := FirestoreClient.Collection("employees").Doc(empID)

	// Check if the document exists before attempting to update
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
			log.Printf("Employee with ID %s not found: %v", empID, err)
			return nil, fmt.Errorf("Employee with ID %s not found", empID)
		}
		log.Printf("Error retrieving document with ID %s: %v", empID, err)
		return nil, fmt.Errorf("Error retrieving document: %v", err)
	}

	// Decode Firestore document data to Employee struct
	if err := docSnapshot.DataTo(&employee); err != nil {
		log.Printf("Error decoding document: %v", err)
		return nil, fmt.Errorf("Error decoding document: %v", err)
	}

	// Update TeamIDs and IAMRoles based on conditions
	if updatedEmp.DeptID != "" {
		docRef1 := FirestoreClient.Collection("departments").Doc(updatedEmp.DeptID)
		docSnapshot1, err := docRef1.Get(ctx)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// Document not found
				return nil, fmt.Errorf("UpdateEmployee ERROR: Document ID %v you entered is not present in document collection: %v", updatedEmp.DeptID, err)
			}

			// Handle other errors
			log.Printf("UpdateEmployee ERROR: Error checking document existence: %v", err)
			return nil, fmt.Errorf("Error checking document existence: %v", err)
		}

		log.Printf("UpdateEmployee INFO: Document ID found!!: %v", docSnapshot1)

		if len(updatedEmp.TeamIDs) == 0 {
			log.Printf("UpdateEmployee ERROR: Please enter teamIDs")
			return nil, fmt.Errorf("Please enter teamIDs")
		}
		employee.TeamIDs = updatedEmp.TeamIDs

		if len(updatedEmp.IAMRoles) == 0 {
			log.Printf("UpdateEmployee ERROR: Please provide IAMroles")
			return nil, fmt.Errorf("Please provide IAMroles")
		} else if len(updatedEmp.IAMRoles) < len(updatedEmp.TeamIDs) {
			log.Printf("UpdateEmployee ERROR: Few teamIDs are missing IAMRoles")
			return nil, fmt.Errorf("Few teamIDs are missing IAMRoles")
		}

		employee.IAMRoles = updatedEmp.IAMRoles

		// Remove and Assign IAM roles
		RemoveMember(empID)
		for key := range employee.IAMRoles {
			err := iamRole.AssignIAM(projectID, employee.IAMRoles[key], employee.Email)
			if err != nil {
				log.Printf("UpdateEmployee ERROR: Failed to assign IAM role: %v", err)
				return nil, fmt.Errorf("Failed to assign IAM role: %v", err)
			}
		}
	} else {
		if len(updatedEmp.TeamIDs) != 0 {
			employee.TeamIDs = mergeSlices(updatedEmp.TeamIDs, employee.TeamIDs)

			// Check if IAMRoles map is empty
			if len(updatedEmp.IAMRoles) == 0 {
				log.Printf("UpdateEmployee ERROR: Enter the roles for newly added teams")
				return nil, fmt.Errorf("Enter roles for newly added teams")
			}

			// Merge IAMRoles maps
			employee.IAMRoles = mergeMaps(updatedEmp.IAMRoles, employee.IAMRoles)
			// Remove and Assign IAM roles
			RemoveMember(empID)
			for key := range employee.IAMRoles {
				err := iamRole.AssignIAM(projectID, employee.IAMRoles[key], employee.Email)
				if err != nil {
					log.Printf("UpdateEmployee ERROR: Failed to assign IAM role: %v", err)
					return nil, fmt.Errorf("Failed to assign IAM role: %v", err)
				}
			}
		}
	}

	if updatedEmp.FirstName != "" {
		employee.FirstName = updatedEmp.FirstName
	}
	if updatedEmp.LastName != "" {
		employee.LastName = updatedEmp.LastName
	}
	if updatedEmp.Email != "" {
		employee.Email = updatedEmp.Email
	}
	if updatedEmp.Password != "" {
		employee.Password = updatedEmp.Password
	}
	if updatedEmp.Role != "" {
		employee.Role = updatedEmp.Role
	}

	// Update the Firestore document with merged data
	_, err = docRef.Set(ctx, employee)
	if err != nil {
		log.Printf("Error updating data in document with ID %s: %v", empID, err)
		return nil, fmt.Errorf("Error updating data in document: %v", err)
	}

	log.Printf("Employee with ID %s updated successfully", empID)
	return &employee, nil
}

func ListEmployee() ([]sharedpackage.Employee, error) {
	ctx := context.Background()

	iter := FirestoreClient.Collection("employees").Documents(ctx)

	var employees []sharedpackage.Employee

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("ERROR: Error iterating over documents: %v", err)
			return nil, err
		}

		var employeeData sharedpackage.Employee
		if err := doc.DataTo(&employeeData); err != nil {
			log.Printf("ERROR: Error converting document data: %v", err)
			return nil, err
		}

		employees = append(employees, employeeData)
		log.Printf("INFO: Processed employee.")
	}

	return employees, nil
}
