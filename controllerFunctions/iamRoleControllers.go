package controllerFunctions

import (
	"Task_04/iamRole"
	"Task_04/sharedpackage"
	"context"
	"errors"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

var FirestoreClient *firestore.Client
var projectID = "ems-web-application-409305"

func removeElementsFromB(A []string, B []string) []string {
	result := make([]string, 0, len(B))

	// Create a map to store elements of slice A for faster lookup
	ASet := make(map[string]struct{})
	for _, value := range A {
		ASet[value] = struct{}{}
	}

	// Iterate through slice B and keep elements not present in A
	for _, value := range B {
		if _, exists := ASet[value]; !exists {
			result = append(result, value)
		}
	}

	return result
}

// InitializeFirestore initializes the Firestore client.
func InitializeFirestore() error {
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("Failed to create Firestore client: %v", err)
	}

	FirestoreClient = client
	log.Println("INFO: Firestore client successfully initialized.")
	return nil
}

func LoginUser(username string) (*string, error) {
	ctx := context.Background()

	// Check if Firestore client is initialized
	if FirestoreClient == nil {
		log.Println("ERROR: Firestore client not initialized")
		return nil, errors.New("Firestore client not initialized")
	}

	// Create a query to search for documents where the specified field equals the specified value
	query := FirestoreClient.Collection("employees").Where("mailID", "==", username)

	// Execute the query
	iter := query.Documents(ctx)
	defer iter.Stop()

	// Iterate through the results
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			// No more documents to iterate
			break
		}
		if err != nil {
			fmt.Println("Error getting document:", err)
			return nil, err
		}

		// Check if the "password" field exists in the document data
		// Use type assertion to get the value and check if it exists
		password, ok := doc.Data()["password"].(string)
		if !ok {
			fmt.Println("Password field not found or not a string in the document")
			return nil, errors.New("Password field not found or not a string in the document")
		}

		// If the password field is found, return it
		return &password, nil
	}

	// If no document is found with the specified username, return an error
	return nil, errors.New("ERROR: User not found")
}

// AddToDBAndAssign adds new roles to an employee document in the Firestore database.
func AssignIAMRole(deptID string, teamID string, empID string, newRoles []string, role string) (*sharedpackage.Employee, error) {
	ctx := context.Background()
	var key string
	// Check if Firestore client is initialized
	if FirestoreClient == nil {
		log.Println("ERROR: Firestore client not initialized")
		return nil, errors.New("Firestore client not initialized")
	}

	// Specify the path to the document
	docRef := FirestoreClient.Collection("employees").Doc(empID)

	// Get the document snapshot
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		log.Printf("ERROR: Error getting document: %v", err)
		return nil, fmt.Errorf("Error getting document: %v", err)
	}

	// Check if the document exists
	if !docSnapshot.Exists() {
		log.Printf("ERROR: Document with ID %s does not exist", empID)
		return nil, fmt.Errorf("Document with ID %s does not exist", empID)
	}

	// Initialize an empty Employee struct
	var employee sharedpackage.Employee

	// Convert Firestore document data to the Employee struct
	if err := docSnapshot.DataTo(&employee); err != nil {
		log.Printf("ERROR: Error converting document data: %v", err)
		return nil, fmt.Errorf("Error converting document data: %v", err)
	}

	log.Printf("INFO: Found employee with ID: %s", empID)

	if employee.DeptID == "" {
		employee.DeptID = deptID
	}
	if employee.Role == "" {
		employee.Role = role
	}
	if len(employee.TeamIDs) == 0 {
		employee.TeamIDs = append(employee.TeamIDs, teamID)
	}

	if teamID == "" {
		key = deptID
	} else {
		key = teamID
	}

	// Iterate over all keys in employee.IAMRoles
	for _, roles := range employee.IAMRoles {
		// Create a map to store unique roles from the existing slice
		existingRolesMap := make(map[string]struct{})
		for _, role := range roles {
			existingRolesMap[role] = struct{}{}
		}

		// Check if any role from newRoles exists in the existingRolesMap
		for i := 0; i < len(newRoles); {
			if _, exists := existingRolesMap[newRoles[i]]; exists {
				// Role exists, remove it from newRoles
				newRoles = append(newRoles[:i], newRoles[i+1:]...)
			} else {
				// Role doesn't exist, move to the next role
				i++
			}
		}
	}

	if roles, keyExists := employee.IAMRoles[key]; keyExists {
		// roles is the slice associated with the provided key

		// Create a map to store unique roles from the existing slice
		existingRolesMap := make(map[string]struct{})
		for _, role := range roles {
			existingRolesMap[role] = struct{}{}
		}
		log.Printf("INFO: Existing roles: %v", existingRolesMap)

		log.Printf("INFO: New roles to be added: %v", newRoles)
		// Add new roles to the existingRolesMap if they are not already present
		for _, role := range newRoles {
			if _, exists := existingRolesMap[role]; !exists {
				roles = append(roles, role)
			}
		}

		// Update the roles for the provided key in the map
		employee.IAMRoles[key] = roles
		log.Printf("INFO: Roles after update: %v", employee.IAMRoles[key])
	} else {
		log.Printf("INFO: Creating new key %v with new roles %v.", key, newRoles)
		employee.IAMRoles[key] = newRoles
		log.Printf("INFO: Created new field with key %v and vale %v.", key, newRoles)
	}

	// Update the document with the modified IAMRoles field
	// if _, err := docRef.Set(ctx, map[string]interface{}{"iamRoles": employee.IAMRoles}, firestore.MergeAll); err != nil {
	// 	log.Printf("ERROR: Error updating document: %v", err)
	// 	return nil, fmt.Errorf("Error updating document: %v", err)
	// }

	_, err = docRef.Set(ctx, employee)
	if err != nil {
		log.Printf("ERROR: Failed to add team document: %v", err)
		return nil, fmt.Errorf("Failed to add team document: %v", err)
	}

	// Call the function directly without specifying the package name
	iamRole.AssignIAM(projectID, employee.IAMRoles[key], employee.Email)
	return &employee, nil
}

// Remove deletes an employee document from the Firestore database.
func RemoveMember(empID string) error {
	// Specify the path to the document
	ctx := context.Background()
	docRef := FirestoreClient.Collection("employees").Doc(empID)

	// Get the document snapshot
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		log.Printf("ERROR: Error getting document: %v", err)
		return fmt.Errorf("Error getting document: %v", err)
	}

	// Check if the document exists
	if !docSnapshot.Exists() {
		log.Printf("ERROR: Document with ID %s does not exist", empID)
		return fmt.Errorf("Document with ID %s does not exist", empID)
	}

	// Initialize an empty Employee struct
	var employee sharedpackage.Employee

	// Convert Firestore document data to the Employee struct
	if err := docSnapshot.DataTo(&employee); err != nil {
		log.Printf("ERROR: Error converting document data: %v", err)
		return fmt.Errorf("Error converting document data: %v", err)
	}

	iamRole.RemoveIAM(projectID, employee.Email)

	// Delete all keys from the map
	for key := range employee.IAMRoles {
		if key == "0" {
			continue
		}
		delete(employee.IAMRoles, key)
	}

	employee.Role = ""
	employee.TeamIDs = make([]string, 0)
	employee.DeptID = ""

	// Update the document with the modified field
	if _, err := docRef.Set(ctx, employee); err != nil {
		log.Printf("ERROR: Error updating document: %v", err)
		return fmt.Errorf("Error updating document: %v", err)
	}

	log.Printf("INFO: Document with ID %s successfully deleted", empID)

	return nil
}

func RemoveIAMRoles(empID string, info sharedpackage.RemoveRoles) (*sharedpackage.Employee, error) {
	ctx := context.Background()
	docRef := FirestoreClient.Collection("employees").Doc(empID)

	// Get the document snapshot
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		log.Printf("ERROR: Error getting document: %v", err)
		return nil, fmt.Errorf("Error getting document: %v", err)
	}

	// Check if the document exists
	if !docSnapshot.Exists() {
		log.Printf("ERROR: Document with ID %s does not exist", empID)
		return nil, fmt.Errorf("Document with ID %s does not exist", empID)
	}

	// Initialize an empty Employee struct
	var employee sharedpackage.Employee

	// Convert Firestore document data to the Employee struct
	if err := docSnapshot.DataTo(&employee); err != nil {
		log.Printf("ERROR: Error converting document data: %v", err)
		return nil, fmt.Errorf("Error converting document data: %v", err)
	}

	// Assuming employee.IAMRoles is a map[string][]string

	// Loop through the map using range
	for key, value := range employee.IAMRoles {
		// Check if the key matches the GroupID
		if key == info.GroupID {
			// Remove elements from value based on info.IAMRoles
			value = removeElementsFromB(info.IAMRoles, value)
			// Update the IAMRoles for the specified key
			employee.IAMRoles[key] = value
		}
	}

	// Remove IAM roles using your iamRole.RemoveIAM function
	iamRole.RemoveIAM(projectID, employee.Email)

	// Loop through the map using range
	for _, roles := range employee.IAMRoles {
		// Assign IAM roles using your iamRole.AssignIAM function
		iamRole.AssignIAM(projectID, roles, employee.Email)
	}

	// Update the document with the modified field
	if _, err := docRef.Set(ctx, employee); err != nil {
		log.Printf("ERROR: Error updating document: %v", err)
		return nil, fmt.Errorf("Error updating document: %v", err)
	}

	return &employee, nil
}
