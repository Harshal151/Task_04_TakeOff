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

func CreateTeamHandler(w http.ResponseWriter, r *http.Request) {
	var team sharedpackage.Team
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&team); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("ERROR: Error decoding request body: %v", err)
		return
	}

	log.Printf("INFO: CreateCustomRoleHandler - Decoded request body fields: %+v", team)
	if team.TeamName == "" {
		http.Error(w, "Please provide team name (teamName)", http.StatusBadRequest)
		log.Println("ERROR: Provide team name (teamName)")
		return
	}
	if team.LeadID == "" {
		http.Error(w, "Please provide team head ID (headID)", http.StatusBadRequest)
		log.Println("ERROR: Please provide team head ID (headID)")
		return
	}
	if len(team.IAMRoles) == 0 {
		http.Error(w, "Please assign roles (roles)", http.StatusBadRequest)
		log.Println("ERROR: Please assign roles (roles)")
		return
	}

	// Add the team and get the data
	data, err := controllerFunctions.CreateTeam(team)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add team: %v", err), http.StatusInternalServerError)
		log.Printf("ERROR: Failed to add team: %v", err)
		return
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal team data to JSON: %v", err), http.StatusInternalServerError)
		log.Printf("ERROR: Failed to marshal team data to JSON: %v", err)
		return
	}

	// Send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func DeleteTeamHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, ok := vars["teamID"]
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		log.Println("WARN: Invalid URL in DeleteTeamHandler")
		return
	}
	log.Printf("INFO: Request received to delete team with ID: %s", teamID)

	data, err := controllerFunctions.DeleteTeam(teamID)
	if err != nil {
		http.Error(w, "Failed to delete team", http.StatusInternalServerError)
		log.Printf("ERROR: Failed to delete team with ID %s: %v", teamID, err)
		return
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal team data to JSON: %v", err), http.StatusInternalServerError)
		log.Printf("ERROR: Failed to marshal team data to JSON: %v", err)
		return
	}

	// Send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func UpdateTeamHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, ok := vars["teamID"]
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		log.Println("UpdateTeamHandler WARN: Invalid URL")
		return
	}
	log.Printf("UpdateTeamHandler INFO: Request received to update employee with ID: %s", teamID)

	var updateTeam sharedpackage.Team
	err := json.NewDecoder(r.Body).Decode(&updateTeam)
	if err != nil {
		log.Printf("UpdateTeamHandler ERROR: Error parsing request body: %v", err)
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: UpdateTeamHandler - Decoded request body fields: %+v", updateTeam)

	if updateTeam.DepartmentID != "" {
		log.Printf("UpdateTeamHandler ERROR: Cannot change department ID: %v", err)
		http.Error(w, "Cannot change department ID", http.StatusBadRequest)
		return
	}
	if updateTeam.CreatedTime != "" || updateTeam.UpdatedTime != "" {
		log.Printf("UpdateTeamHandler ERROR: cannot change created time and updated time: %v", err)
		http.Error(w, "cannot change created time and updated time", http.StatusBadRequest)
		return
	}

	// Add the department and get the data
	data, err := controllerFunctions.UpdateTeam(teamID, updateTeam)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update employee: %v", err), http.StatusInternalServerError)
		log.Printf("UpdateTeamHandler ERROR: Failed to update employee: %v", err)
		return
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal updated team data to JSON: %v", err), http.StatusInternalServerError)
		log.Printf("UpdateTeamHandler ERROR: Failed to marshal updated team data to JSON: %v", err)
		return
	}

	// Send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

}

func ListTeamHandler(w http.ResponseWriter, r *http.Request) {
    // Retrieve all teams
    teams, err := controllerFunctions.ListTeams()
    if err != nil {
        log.Printf("ERROR: Failed to get teams: %v", err)
        http.Error(w, "Failed to retrieve teams", http.StatusInternalServerError)
        return
    }
    log.Printf("INFO: Retrieved %d teams", len(*teams))

    // Convert teams to JSON format
    teamsJSON, err := json.Marshal(teams)
    if err != nil {
        log.Printf("ERROR: Failed to marshal teams to JSON: %v", err)
        http.Error(w, "Failed to convert teams to JSON", http.StatusInternalServerError)
        return
    }
    log.Printf("INFO: Successfully marshaled teams to JSON")

    w.Header().Set("Content-Type", "application/json")

    // Write JSON response
    if _, err := w.Write(teamsJSON); err != nil {
        log.Printf("ERROR: Failed to send teams JSON response: %v", err)
        http.Error(w, "Failed to send teams JSON response", http.StatusInternalServerError)
        return
    }
    log.Println("INFO: Sent teams JSON response")
}
