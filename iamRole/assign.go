package iamRole

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"google.golang.org/api/cloudresourcemanager/v1"
	iam "google.golang.org/api/iam/v1"
)

//PREDIFINED ROLE PART WITH ROLE ASSIGNMENT

// RemoveIAM removes the specified user from the IAM roles in the project.
func RemoveIAM(proID string, userMail string) {
	projectID := proID
	member := "user:" + userMail

	// Initializes the Cloud Resource Manager service
	ctx := context.Background()
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		log.Fatalf("cloudresourcemanager.NewService: %v", err)
	}

	// Remove the specified member from IAM roles
	removeMember(crmService, projectID, member)
	log.Printf("INFO: Removed IAM roles for user %s in project %s", userMail, projectID)
}

// AssignIAM assigns the specified roles to the user in the project.
func AssignIAM(proID string, roles []string, userMail string) {
	projectID := proID
	member := "user:" + userMail

	// Initializes the Cloud Resource Manager service
	ctx := context.Background()
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		log.Fatalf("cloudresourcemanager.NewService: %v", err)
	}

	// Assign each specified role to the user
	for _, role := range roles {
		addBinding(crmService, projectID, member, role)
		log.Printf("INFO: Assigned IAM role %s to user %s in project %s", role, userMail, projectID)
	}
}

// addBinding adds the specified member to the project's IAM policy with the given role.
func addBinding(crmService *cloudresourcemanager.Service, projectID, member, role string) {
	policy := getPolicy(crmService, projectID)

	// Find or create a binding for the specified role
	binding := findOrCreateBinding(policy, role)
	binding.Members = append(binding.Members, member)

	setPolicy(crmService, projectID, policy)
}

// removeMember removes the specified user from all roles in the project's IAM policy.
func removeMember(crmService *cloudresourcemanager.Service, projectID, member string) {
	policy := getPolicy(crmService, projectID)

	// Remove the specified member from all roles in the IAM policy
	updatedBindings := removeMemberFromBindings(policy.Bindings, member)
	policy.Bindings = updatedBindings

	setPolicy(crmService, projectID, policy)
}

// getPolicy gets the IAM policy for the specified project.
func getPolicy(crmService *cloudresourcemanager.Service, projectID string) *cloudresourcemanager.Policy {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	request := new(cloudresourcemanager.GetIamPolicyRequest)
	policy, err := crmService.Projects.GetIamPolicy(projectID, request).Do()
	if err != nil {
		log.Fatalf("Projects.GetIamPolicy: %v", err)
	}

	return policy
}

// setPolicy sets the IAM policy for the specified project.
func setPolicy(crmService *cloudresourcemanager.Service, projectID string, policy *cloudresourcemanager.Policy) {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	request := new(cloudresourcemanager.SetIamPolicyRequest)
	request.Policy = policy

	policy, err := crmService.Projects.SetIamPolicy(projectID, request).Do()
	if err != nil {
		log.Fatalf("Projects.SetIamPolicy: %v", err)
	}
}

// findOrCreateBinding finds or creates a binding for the specified role in the policy.
func findOrCreateBinding(policy *cloudresourcemanager.Policy, role string) *cloudresourcemanager.Binding {
	for _, b := range policy.Bindings {
		if b.Role == role {
			return b
		}
	}

	// If the binding does not exist, create a new one
	newBinding := &cloudresourcemanager.Binding{
		Role:    role,
		Members: []string{},
	}
	policy.Bindings = append(policy.Bindings, newBinding)

	return newBinding
}

// removeMemberFromBindings removes the specified member from all bindings.
func removeMemberFromBindings(bindings []*cloudresourcemanager.Binding, member string) []*cloudresourcemanager.Binding {
	var updatedBindings []*cloudresourcemanager.Binding

	for _, binding := range bindings {
		updatedMembers := removeMemberFromSlice(binding.Members, member)

		// If the binding still has members after removal, update the binding
		if len(updatedMembers) > 0 {
			updatedBinding := &cloudresourcemanager.Binding{
				Role:    binding.Role,
				Members: updatedMembers,
			}
			updatedBindings = append(updatedBindings, updatedBinding)
		}
	}

	return updatedBindings
}

// removeMemberFromSlice removes the specified member from the string slice.
func removeMemberFromSlice(slice []string, member string) []string {
	var updatedSlice []string

	for _, item := range slice {
		if item != member {
			updatedSlice = append(updatedSlice, item)
		}
	}

	return updatedSlice
}

//CUSTOM ROLE PART

// createRole creates a custom role.
func CreateRole(w io.Writer, projectID, name, title, description, stage string, permissions []string) (*iam.Role, error) {
	ctx := context.Background()
	service, err := iam.NewService(ctx)
	if err != nil {
		log.Printf("ERROR: Error creating IAM service: %v", err)
		return nil, fmt.Errorf("iam.NewService: %w", err)
	}

	request := &iam.CreateRoleRequest{
		Role: &iam.Role{
			Title:               title,
			Description:         description,
			IncludedPermissions: permissions,
			Stage:               stage,
		},
		RoleId: name,
	}

	// Log request details
	log.Printf("INFO: Creating custom role: %v", request)

	role, err := service.Projects.Roles.Create("projects/"+projectID, request).Do()
	if err != nil {
		log.Printf("ERROR: Error creating custom role: %v", err)
		return nil, fmt.Errorf("Projects.Roles.Create: %w", err)
	}

	// Log successful creation
	log.Printf("INFO: Custom role created: %v", role.Name)

	// Write success message to the provided writer
	fmt.Fprintf(w, "Created role: %v", role.Name)

	return role, nil
}

// deleteRole deletes a custom role.
func DeleteRole(w io.Writer, projectID, name string) error {
	ctx := context.Background()
	service, err := iam.NewService(ctx)
	if err != nil {
		return fmt.Errorf("iam.NewService: %w", err)
	}

	_, err = service.Projects.Roles.Delete("projects/" + projectID + "/roles/" + name).Do()
	if err != nil {
		return fmt.Errorf("Projects.Roles.Delete: %w", err)
	}
	fmt.Fprintf(w, "Deleted role: %v", name)
	return nil
}

// UndeleteRole restores a deleted custom role.
func UndeleteRole(w io.Writer, projectID, name string) error {
	ctx := context.Background()
	service, err := iam.NewService(ctx)
	if err != nil {
		return fmt.Errorf("iam.NewService: %w", err)
	}

	resource := "projects/" + projectID + "/roles/" + name
	request := &iam.UndeleteRoleRequest{}
	role, err := service.Projects.Roles.Undelete(resource, request).Do()
	if err != nil {
		return fmt.Errorf("Projects.Roles.Undelete: %w", err)
	}

	// Log the success message
	fmt.Fprintf(w, "Undeleted role: %v", role.Name)
	return nil
}

// listRoles lists a project's roles.
func ListCustomRoles(w io.Writer, projectID string) ([]string, error) {
	ctx := context.Background()
	service, err := iam.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("iam.NewService: %w", err)
	}

	response, err := service.Projects.Roles.List("projects/" + projectID).Do()
	if err != nil {
		return nil, fmt.Errorf("Projects.Roles.List: %w", err)
	}

	var roleNames []string
	num := 1
	log.Printf("INFO: Listing Custom Roles: ")
	for _, role := range response.Roles {
		// Extract the last part of the role name
		roleNameParts := strings.Split(role.Name, "/")
		if len(roleNameParts) > 0 {

			lastPart := roleNameParts[len(roleNameParts)-1]
			roleNames = append(roleNames, lastPart)
			fmt.Fprintf(w, "%d] %v\n", num, lastPart)
		}
		num++
	}
	return roleNames, nil
}

// UpdateCustomRole modifies a custom role.
func UpdateCustomRole(w io.Writer, projectID, name, newTitle, newDescription, newStage string, newPermissions []string) (*iam.Role, error) {
	ctx := context.Background()
	service, err := iam.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("iam.NewService: %w", err)
	}

	resource := "projects/" + projectID + "/roles/" + name

	// Retrieve the existing role
	role, err := service.Projects.Roles.Get(resource).Do()
	if err != nil {
		return nil, fmt.Errorf("Projects.Roles.Get: %w", err)
	}

	// Update role fields if new values are provided
	if newTitle != "" {
		role.Title = newTitle
	}

	if newDescription != "" {
		role.Description = newDescription
	}

	if len(newPermissions) > 0 {
		role.IncludedPermissions = newPermissions
	}

	if newStage != "" {
		role.Stage = newStage
	}

	// Patch the role with updated information
	role, err = service.Projects.Roles.Patch(resource, role).Do()
	if err != nil {
		return nil, fmt.Errorf("Projects.Roles.Patch: %w", err)
	}

	// Log and return the updated role
	fmt.Fprintf(w, "Updated role: %v", role.Name)
	return role, nil
}

