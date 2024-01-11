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
func AddToDBAndAssign(id string, newRoles []string) (*string, error) {
	ctx := context.Background()
	// Check if Firestore client is initialized
	if FirestoreClient == nil {
		log.Println("ERROR: Firestore client not initialized")
		return nil, errors.New("Firestore client not initialized")
	}

	// Specify the path to the document
	docRef := FirestoreClient.Collection("employees").Doc(id)

	// Get the document snapshot
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		log.Printf("ERROR: Error getting document: %v", err)
		return nil, fmt.Errorf("Error getting document: %v", err)
	}

	// Check if the document exists
	if !docSnapshot.Exists() {
		log.Printf("ERROR: Document with ID %s does not exist", id)
		return nil, fmt.Errorf("Document with ID %s does not exist", id)
	}

	// Initialize an empty Employee struct
	var employee sharedpackage.Employee

	// Convert Firestore document data to the Employee struct
	if err := docSnapshot.DataTo(&employee); err != nil {
		log.Printf("ERROR: Error converting document data: %v", err)
		return nil, fmt.Errorf("Error converting document data: %v", err)
	}

	log.Printf("INFO: Found employee with ID: %s", id)

	// Create a set to store unique roles
	uniqueRoles := make(map[string]struct{})

	// Add existing roles to the set
	for _, role := range employee.IAMRoles {
		uniqueRoles[role] = struct{}{}
	}

	// Add new roles to the set
	for _, role := range newRoles {
		uniqueRoles[role] = struct{}{}
	}

	// Convert the set back to a slice
	var roles []string
	for role := range uniqueRoles {
		roles = append(roles, role)
	}

	// Log the roles before update
	log.Printf("INFO: Existing roles: %v", employee.IAMRoles)
	log.Printf("INFO: New roles to be added: %v", newRoles)
	log.Printf("INFO: Roles after update: %v", roles)

	// Check if the "roles" field is empty before updating
	if len(roles) > 0 {
		// Update the "role" field in the document
		updateData := map[string]interface{}{
			"iamRoles": roles,
		}

		// Update the document with the modified field
		if _, err := docRef.Set(ctx, updateData, firestore.MergeAll); err != nil {
			log.Printf("ERROR: Error updating document: %v", err)
			return nil, fmt.Errorf("Error updating document: %v", err)
		}
	} else {
		log.Printf("INFO: No new roles to add. Skipping update.")
	}

	// Call the function directly without specifying the package name
	iamRole.AssignIAM(projectID, roles, employee.Email)
	returnMessage := "Role assigned to principal " + employee.Email + "."
	log.Printf("INFO: Role successfully assigned to userID %s.", id)
	return &returnMessage, nil
}

// Remove deletes an employee document from the Firestore database.
func Remove(id string) error {
	// Specify the path to the document
	ctx := context.Background()
	docRef := FirestoreClient.Collection("employees").Doc(id)

	// Get the document snapshot
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		log.Printf("ERROR: Error getting document: %v", err)
		return fmt.Errorf("Error getting document: %v", err)
	}

	// Check if the document exists
	if !docSnapshot.Exists() {
		log.Printf("ERROR: Document with ID %s does not exist", id)
		return fmt.Errorf("Document with ID %s does not exist", id)
	}

	// Initialize an empty Employee struct
	var employee sharedpackage.Employee

	// Convert Firestore document data to the Employee struct
	if err := docSnapshot.DataTo(&employee); err != nil {
		log.Printf("ERROR: Error converting document data: %v", err)
		return fmt.Errorf("Error converting document data: %v", err)
	}

	iamRole.RemoveIAM(projectID, employee.Email)

	// Clear the Role field
	employee.IAMRoles = nil

	// Update the document with the modified field
	if _, err := docRef.Set(ctx, employee); err != nil {
		log.Printf("ERROR: Error updating document: %v", err)
		return fmt.Errorf("Error updating document: %v", err)
	}

	log.Printf("INFO: Document with ID %s successfully deleted", id)

	return nil
}
