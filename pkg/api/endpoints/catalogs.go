package endpoints

import (
	"errors"
	"fmt"
	"maps"

	"github.com/gofiber/fiber/v3"
	"github.com/peeklapp/peekl/pkg/api/responses"
	"github.com/peeklapp/peekl/pkg/config"
	"github.com/peeklapp/peekl/pkg/inventory"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/roles"
)

func GetCatalog(ctx fiber.Ctx) error {
	// Get node name
	nodeName := ctx.RequestCtx().TLSConnectionState().PeerCertificates[0].Subject.CommonName

	// Get configuration
	conf, _ := ctx.Locals("config").(*config.ServerConfig)

	// Response
	var resp responses.GetCatalog

	// Declare resources slices
	var resources []models.Resource

	// Tags
	var tags []string

	// Variables
	nodeVars := map[string]any{}
	groupsVars := map[string]any{}
	rolesVars := map[string]any{}

	// Roles to load
	var rolesToLoad []string

	// Get node from inventory
	node, err := inventory.LoadNodeFromInventory(conf.Code.Directory, nodeName)
	if err != nil {
		if errors.As(err, &inventory.NodeNotFoundError{}) {
			ctx.Status(404).JSON(responses.ErrorResponse{
				Error:   "Node not found in inventory",
				Details: fmt.Sprintf("The node %s could not be found in the inventory", nodeName),
			})
			return nil
		} else {
			ctx.Status(500).JSON(responses.ErrorResponse{
				Error:   "Internal Server Error",
				Details: err.Error(),
			})
			return nil
		}
	}

	// Add node resources to resources of catalog
	for _, nodeRes := range node.Resources {
		resources = append(resources, *nodeRes)
	}

	// Add node tags to tags of catalog
	tags = append(tags, node.Tags...)

	// Add node variables to variables
	nodeVars = node.Variables

	// Get list of roles to load
	rolesToLoad = append(rolesToLoad, node.Roles...)

	// Get all groups the node is a member of
	for _, group := range node.Groups {
		groupInv, err := inventory.LoadGroupFromInventory(conf.Code.Directory, group)
		if err != nil {
			if errors.As(err, &inventory.GroupNotFoundError{}) {
				ctx.Status(404).JSON(responses.ErrorResponse{
					Error:   "Group not found in inventory",
					Details: fmt.Sprintf("The group %s could not be found in the inventory"),
				})
				return nil
			} else {
				ctx.Status(500).JSON(responses.ErrorResponse{
					Error:   "Internal Server Error",
					Details: err.Error(),
				})
				return nil
			}
		}
		// Add group resources to resources of catalog
		for _, groupRes := range groupInv.Resources {
			resources = append(resources, *groupRes)
		}

		// Add group tags to tags of catalog
		tags = append(tags, groupInv.Tags...)

		// Add group variables to variables
		maps.Copy(groupsVars, groupInv.Variables)

		// Get list of roles to load
		rolesToLoad = append(rolesToLoad, groupInv.Roles...)
	}

	// Load all roles
	var loadedRoles []models.Role
	for _, role := range rolesToLoad {
		loadedRole, err := roles.LoadRoleFromCode(conf.Code.Directory, role)
		if err != nil {
			if errors.As(err, &roles.RoleNotFoundError{}) {
				ctx.Status(404).JSON(responses.ErrorResponse{
					Error:   "Role could not be found",
					Details: err.Error(),
				})
				return nil
			} else {
				ctx.Status(500).JSON(responses.ErrorResponse{
					Error:   "Internal Server Error",
					Details: err.Error(),
				})
				return nil
			}
		}

		// Append loaded roles to roles
		loadedRoles = append(loadedRoles, *loadedRole)

		// Append variables of roles to variables
		maps.Copy(rolesVars, loadedRole.Variables)
	}

	// Generate response variables with precedence
	// - Roles
	// - Overwritten by groups vars
	// - Overwritten by node vars
	respVars := map[string]any{}
	maps.Copy(respVars, rolesVars)
	maps.Copy(respVars, groupsVars)
	maps.Copy(respVars, nodeVars)

	// Assign data to response
	resp.GlobalResource = resources
	resp.Roles = loadedRoles
	resp.Tags = tags
	resp.Variables = respVars

	// Send resources to node
	ctx.Status(200).JSON(resp)
	return nil
}
