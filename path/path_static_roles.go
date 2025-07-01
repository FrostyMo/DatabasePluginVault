package path

import (
	"DatabasePluginVault/role"
	"DatabasePluginVault/storage"
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// PathStaticRoles returns the Vault path definitions for static roles.
func PathStaticRoles() *framework.Path {
	return &framework.Path{
		Pattern: "static-roles/" + framework.GenericNameRegex("db_type") + "/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Name of the static role.",
				Required:    true,
			},
			"db_type": {
				Type:        framework.TypeString,
				Description: "Type of the database (e.g. mysql, snowflake).",
				Required:    true,
			},
			"connection_name": {
				Type:        framework.TypeString,
				Description: "Name of the DB config to use.",
				Required:    true,
			},
			"username": {
				Type:        framework.TypeString,
				Description: "Database username to manage.",
				Required:    true,
			},
			"password_policy": {
				Type:        framework.TypeString,
				Description: "Vault password policy to use during rotation.",
				Required:    false,
			},
			"rotation_statements": {
				Type:        framework.TypeStringSlice,
				Description: "SQL statements to use during password rotation.",
				Required:    false,
			},
		},
		ExistenceCheck: nil,
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback:                    handleStaticRoleWrite,
				ForwardPerformanceSecondary: true,
				ForwardPerformanceStandby:   true,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback:                    handleStaticRoleWrite,
				ForwardPerformanceSecondary: true,
				ForwardPerformanceStandby:   true,
			},
			logical.ReadOperation: &framework.PathOperation{Callback: handleStaticRoleRead},
			logical.DeleteOperation: &framework.PathOperation{
				Callback:                    handleStaticRoleDelete,
				ForwardPerformanceSecondary: true,
				ForwardPerformanceStandby:   true,
			},
		},
		HelpSynopsis:    "Manage static database roles.",
		HelpDescription: "This path lets you configure static roles for database credential management.",
	}
}

func handleStaticRoleWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	st := storage.NewBackendStorage(req.Storage)

	roleObj := &role.StaticRole{
		Name:           d.Get("name").(string),
		DBType:         d.Get("db_type").(string),
		ConnectionName: d.Get("connection_name").(string),
		Username:       d.Get("username").(string),
		PasswordPolicy: d.Get("password_policy").(string),
	}

	if v, ok := d.GetOk("rotation_statements"); ok {
		roleObj.RotationSQL = v.([]string)
	}

	if err := role.CreateOrUpdateStaticRole(ctx, st, roleObj); err != nil {
		return logical.ErrorResponse(fmt.Sprintf("failed to save role: %v", err)), nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"message": "Static role saved successfully",
		},
	}, nil
}

func handleStaticRoleRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	st := storage.NewBackendStorage(req.Storage)

	dbType := d.Get("db_type").(string)
	name := d.Get("name").(string)

	roleObj, err := role.GetStaticRole(ctx, st, dbType, name)
	if err != nil {
		return nil, err
	}
	if roleObj == nil {
		return nil, nil
	}

	resp := map[string]interface{}{
		"name":                roleObj.Name,
		"db_type":             roleObj.DBType,
		"connection_name":     roleObj.ConnectionName,
		"username":            roleObj.Username,
		"password_policy":     roleObj.PasswordPolicy,
		"rotation_statements": roleObj.RotationSQL,
	}

	return &logical.Response{Data: resp}, nil
}

func handleStaticRoleDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	st := storage.NewBackendStorage(req.Storage)

	dbType := d.Get("db_type").(string)
	name := d.Get("name").(string)

	if err := role.DeleteStaticRole(ctx, st, dbType, name); err != nil {
		return logical.ErrorResponse(fmt.Sprintf("failed to delete role: %v", err)), nil
	}
	return &logical.Response{}, nil
}

// a LIST operation for static roles under a db type
func PathStaticRoleList() *framework.Path {
	return &framework.Path{
		Pattern: "static-roles/" + framework.GenericNameRegex("db_type") + "$",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{Callback: handleStaticRoleList},
		},
		HelpSynopsis:    "List static roles for a given DB type.",
		HelpDescription: "Returns a list of static role names configured under the given DB type.",
	}
}

func handleStaticRoleList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	dbType := d.Get("db_type").(string)
	prefix := fmt.Sprintf("roles/%s/", dbType)

	keys, err := req.Storage.List(ctx, prefix)
	if err != nil {
		return nil, err
	}
	return logical.ListResponse(keys), nil
}
