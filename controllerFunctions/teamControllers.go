package controllerFunctions

import (
	"Task_04/sharedpackage"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// generateIncrementingTeamID generates an incrementing document ID for the given collection.
func generateIncrementingTeamID(collection *firestore.CollectionRef) (string, error) {
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
		idStr := doc.Ref.ID[len("team_"):]
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
		newIDStr := fmt.Sprintf("team_%d", newID)

		// Check if the generated ID already exists
		if !existingIDs[newIDStr] {
			return newIDStr, nil
		}

		highestID++
	}

	// This should not be reached
}

func CreateTeam(team sharedpackage.Team) (*sharedpackage.Team, error) {
	ctx := context.Background()
	currentTime := time.Now()
	formattedTime := currentTime.Format("Mon, 02 Jan 2006 15:04:05 MST")

	// Specify the path to the "departments" collection
	teamCollection := FirestoreClient.Collection("teams")
	employeeCollection := FirestoreClient.Collection("employees")

	// Generate an incrementing document ID
	newDocID, err := generateIncrementingTeamID(teamCollection)
	if err != nil {
		log.Printf("ERROR: Unable to generate a unique document ID: %v", err)
		return nil, fmt.Errorf("Unable to generate a unique document ID: %v", err)
	}

	if team.DepartmentID != "" {
		docRef := FirestoreClient.Collection("departments").Doc(team.DepartmentID)
		_, err := docRef.Get(ctx)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// Document not found
				log.Printf("CreateTeam ERROR: Document ID %v you entered is not present in departments collection: %v", team.DepartmentID, err)
				return nil, fmt.Errorf("Document ID %v you entered is not present in departments collection: %v", team.DepartmentID, err)
			}

			// Handle other errors
			log.Printf("CreateTeam ERROR: Error checking document existence: %v", err)
			return nil, fmt.Errorf("Error checking document existence: %v", err)
		}

		log.Printf("CreateTeam INFO: Document ID found!!")
	}

	if team.LeadID != "" {
		// Get the existing document data
		existingData, err := employeeCollection.Doc(team.LeadID).Get(ctx)
		if err != nil {
			// Handle the error
			log.Printf("ERROR: Failed to get employee document: %v", err)
			return nil, err
		}

		// Check if the document exists
		if existingData.Exists() {
			// Extract existing departmentID and teamIDs if available
			existingDepartmentIDInterface, err := existingData.DataAt("departmentID")
			if err != nil {
				// Handle the error
				log.Printf("ERROR: Failed to get departmentID: %v", err)
				return nil, err
			}

			existingTeamIDsInterface, err := existingData.DataAt("teamIDs")
			if err != nil {
				// Handle the error
				log.Printf("ERROR: Failed to get teamIDs: %v", err)
				return nil, err
			}

			// Type assert to get the actual types
			existingDepartmentID, ok := existingDepartmentIDInterface.(string)
			if !ok {
				// Handle the type assertion failure
				log.Printf("ERROR: Failed to assert departmentID as string")
				return nil, fmt.Errorf("Failed to assert departmentID as string")
			}

			existingTeamIDs, ok := existingTeamIDsInterface.([]interface{})
			if !ok {
				// Handle the type assertion failure
				log.Printf("ERROR: Failed to assert teamIDs as []interface{}")
				return nil, fmt.Errorf("Failed to assert teamIDs as []interface{}")
			}

			// Check whether departmentID and teamIDs are empty
			if existingDepartmentID == "" && len(existingTeamIDs) == 0 {
				// Both departmentID and teamIDs are empty
				log.Println("INFO: DepartmentID and TeamIDs are empty.")
				team.CreatedTime = formattedTime
				// Set the document data for the employee with merge
				_, err = employeeCollection.Doc(team.LeadID).Set(ctx, map[string]interface{}{
					"role":         "Lead",
					"departmentID": team.DepartmentID,
					"teamIDs":      firestore.ArrayUnion(newDocID),
				}, firestore.MergeAll)
			} else if existingDepartmentID != "" && len(existingTeamIDs) == 0 {
				// Either departmentID or teamIDs is not empty
				log.Printf("ERROR: Employee %v id HOD of department %v.", team.LeadID, existingDepartmentID)
				return nil, fmt.Errorf("Cannot assign lead role to HOD.")
			} else {
				log.Printf("ERROR: Employee %v is already in a different department and team.", team.LeadID)
				return nil, fmt.Errorf("Employee is already in a different department and team.")
			}
		} else {
			// Document does not exist
			log.Println("INFO: Employee document does not exist.")
		}
	}

	log.Printf("INFO: Department document with ID %s and employee document updated successfully", newDocID)
	AssignIAMRole(team.DepartmentID, newDocID, team.LeadID, team.IAMRoles, "Lead")
	// Add the team data to the "teams" collection with the generated document ID
	_, err = teamCollection.Doc(newDocID).Set(ctx, team)
	if err != nil {
		log.Printf("ERROR: Failed to add team document: %v", err)
		return nil, fmt.Errorf("Failed to add team document: %v", err)
	}

	return &team, nil
}

func DeleteTeam(teamID string) (*sharedpackage.Team, error) {
	var deletedTeam sharedpackage.Team
	ctx := context.Background()

	// Step 1: Get the departments collection
	teamsCollection := FirestoreClient.Collection("teams")

	// Step 2: Delete teams associated with the department
	if err := deleteTeams(teamID); err != nil {
		log.Printf("ERROR: Failed to delete teams: %v", err)
		// You can choose to return the error or handle it based on your requirements
		return nil, fmt.Errorf("Failed to delete teams: %v", err)
	}

	// Step 3: Specify the path to the document
	docRef := teamsCollection.Doc(teamID)

	// Step 4: Check if the document exists
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		log.Printf("Error getting document: %v", err)
		return nil, fmt.Errorf("Error getting document: %v", err)
	}

	if !docSnapshot.Exists() {
		log.Printf("Document with ID %s does not exist", teamID)
		return nil, fmt.Errorf("Document with ID %s does not exist", teamID)
	}

	// Step 5: Extract the value of the "headID" field
	leadID, ok := docSnapshot.Data()["leadID"].(string)
	if !ok {
		log.Printf("Field 'headID' not found or not a string")
		return nil, fmt.Errorf("Field 'headID' not found or not a string")
	}

	log.Printf("INFO: Extracted 'leadID' from document: %s", leadID)
	// Step 6: Handle the error from removeAccess
	if err := removeEmployees(teamID, leadID); err != nil {
		log.Printf("ERROR: Failed to remove access: %v", err)
		// You can choose to return the error or handle it based on your requirements
		return nil, fmt.Errorf("Failed to remove access: %v", err)
	}

	// Step 7: Delete the document
	_, err = docRef.Delete(ctx)
	if err != nil {
		log.Printf("ERROR: Error deleting document: %v", err)
		return nil, fmt.Errorf("Error deleting document: %v", err)
	}

	return &deletedTeam, nil
}

func UpdateTeam(teamID string, team sharedpackage.Team) (*sharedpackage.Team, error) {
	ctx := context.Background()
	teamCollection := FirestoreClient.Collection("teams")
	employeeCollection := FirestoreClient.Collection("employees")

	docTeam := teamCollection.Doc(teamID)
	// Step 4: Check if the document exists
	docSnapshotTeam, err := docTeam.Get(ctx)
	if err != nil {
		log.Printf("Error getting document: %v", err)
		return nil, fmt.Errorf("Error getting document: %v", err)
	}

	if !docSnapshotTeam.Exists() {
		log.Printf("Document with ID %s does not exist", teamID)
		return nil, fmt.Errorf("Document with ID %s does not exist", teamID)
	}

	var teamData sharedpackage.Team
	// Convert Firestore document data to the Employee struct
	if err := docSnapshotTeam.DataTo(&teamData); err != nil {
		log.Printf("ERROR: Error converting document data: %v", err)
		return nil, fmt.Errorf("Error converting document data: %v", err)
	}

	// Step 5: Extract the value of the "headID" field
	leadID := teamData.LeadID

	docEmployee := employeeCollection.Doc(leadID)
	// Step 4: Check if the document exists
	docSnapshotEmployee, err := docEmployee.Get(ctx)
	if err != nil {
		log.Printf("Error getting document: %v", err)
		return nil, fmt.Errorf("Error getting document: %v", err)
	}

	if !docSnapshotEmployee.Exists() {
		log.Printf("Document with ID %s does not exist", teamID)
		return nil, fmt.Errorf("Document with ID %s does not exist", teamID)
	}
	var employee sharedpackage.Employee
	// Convert Firestore document data to the Employee struct
	if err := docSnapshotEmployee.DataTo(&employee); err != nil {
		log.Printf("ERROR: Error converting document data: %v", err)
		return nil, fmt.Errorf("Error converting document data: %v", err)
	}

	if team.LeadID != "" {
		docEmployee2 := employeeCollection.Doc(team.LeadID)

		docSnapshotEmployee2, err := docEmployee2.Get(ctx)
		if err != nil {
			log.Printf("Error getting document: %v", err)
			return nil, fmt.Errorf("Error getting document: %v", err)
		}

		if !docSnapshotEmployee2.Exists() {
			log.Printf("Document with ID %s does not exist", team.LeadID)
			return nil, fmt.Errorf("Document with ID %s does not exist", team.LeadID)
		}
		var employee1 sharedpackage.Employee
		// Convert Firestore document data to the Employee struct
		if err := docSnapshotEmployee2.DataTo(&employee1); err != nil {
			log.Printf("ERROR: Error converting document data: %v", err)
			return nil, fmt.Errorf("Error converting document data: %v", err)
		}
		employee1.DeptID = employee.DeptID
		employee1.Role = employee.Role
		employee1.TeamIDs = employee.TeamIDs
		employee1.IAMRoles =employee.IAMRoles
		data, err := AssignIAMRole(employee1.DeptID, employee1.TeamIDs[0], team.LeadID, employee1.IAMRoles[employee1.TeamIDs[0]], employee1.Role)
		if err != nil {
			log.Printf("ERROR: Error while assigning IAM role: %v", err)
			return nil, fmt.Errorf("Error assigning IAM role: %v", err)
		}
		log.Printf("INFO: Assigned IAM roles to %v: %v", team.LeadID, data)
		RemoveMember(leadID)
		employee.DeptID = ""
		// Update the document with the modified field
		if _, err := docEmployee2.Set(ctx, employee1); err != nil {
			log.Printf("ERROR: Error updating document: %v", err)
			return nil, fmt.Errorf("Error updating document: %v", err)
		}
	} else {
		team.LeadID = teamData.LeadID
	}
	if len(team.IAMRoles) != 0 {
		employee.IAMRoles[teamID] = mergeSlices(team.IAMRoles, employee.IAMRoles[teamID])
		role := employee.Role
		teamIDs := employee.TeamIDs
		deptID := employee.DeptID
		RemoveMember(leadID)
		data, err := AssignIAMRole(deptID, teamID, team.LeadID, employee.IAMRoles[teamID], role)
		if err != nil {
			log.Printf("ERROR: Error while assigning IAM role: %v", err)
			return nil, fmt.Errorf("Error assigning IAM role: %v", err)
		}
		log.Printf("INFO: Assigned IAM roles to %v: %v", team.LeadID, data)
		employee.Role = role
		employee.TeamIDs = teamIDs
		employee.DeptID = deptID
		// Update the document with the modified field
		if _, err := docEmployee.Set(ctx, employee); err != nil {
			log.Printf("ERROR: Error updating document: %v", err)
			return nil, fmt.Errorf("Error updating document: %v", err)
		}
		team.IAMRoles = employee.IAMRoles[teamID]
	} else {
		team.IAMRoles = teamData.IAMRoles
	}
	if team.TeamName == "" {
		team.TeamName = teamData.TeamName
	}

	currentTime := time.Now()
	updatedTime := currentTime.Format("Mon, 02 Jan 2006 15:04:05 MST")
	team.UpdatedTime = updatedTime
	team.CreatedTime = teamData.CreatedTime
	team.DepartmentID = teamData.DepartmentID

	// Update the Firestore document with merged data
	_, err = docTeam.Set(ctx, team)
	if err != nil {
		log.Printf("Error updating data in document with ID %s: %v", teamID, err)
		return nil, fmt.Errorf("Error updating data in document: %v", err)
	}

	log.Printf("Employee with ID %s updated successfully", teamID)

	return &team, nil
}

func ListTeams() (*[]sharedpackage.Team, error) {
	ctx := context.Background()

	iter := FirestoreClient.Collection("teams").Documents(ctx)

	var teams []sharedpackage.Team

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("ERROR: Error iterating over documents: %v", err)
			return nil, err
		}

		var teamData sharedpackage.Team
		if err := doc.DataTo(&teamData); err != nil {
			log.Printf("ERROR: Error converting document data: %v", err)
			return nil, err
		}

		teams = append(teams, teamData)
		log.Printf("INFO: Processed team.")
	}

	return &teams, nil
}
