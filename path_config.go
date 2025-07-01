package dbsecretengine

import (
	"DatabasePluginVault/internal/dbengines"
	"DatabasePluginVault/storage"
	"context"
	"fmt"
	"github.com/hashicorp/vault/helper/versions"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"log"
	"net/url"
)

func pathConfigurePluginConnection(b *databaseBackend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: "config/" + framework.GenericNameRegex("name"),
			DisplayAttrs: &framework.DisplayAttributes{
				OperationPrefix: "database",
			},
			Fields: map[string]*framework.FieldSchema{
				"name": {
					Type:        framework.TypeString,
					Description: "Name of this database connection",
					Required:    true,
				},
				"plugin_name": {
					Type:        framework.TypeString,
					Description: "Name of the plugin to use.",
				},
				"plugin_version": {
					Type:        framework.TypeString,
					Description: "Version of the plugin to use.",
				},
				"verify_connection": {
					Type:        framework.TypeBool,
					Default:     true,
					Description: "If true, the connection details are verified by connecting to the database.",
				},
				"allowed_roles": {
					Type:        framework.TypeCommaStringSlice,
					Description: "Comma-separated or array of allowed role names. '*' allows all.",
				},
				"root_rotation_statements": {
					Type:        framework.TypeCommaStringSlice,
					Description: "Statements to execute for rotating root credentials.",
				},
				"password_policy": {
					Type:        framework.TypeString,
					Description: "Password policy name.",
				},
				"ci_name": {
					Type:        framework.TypeString,
					Description: "Service CI name for config.",
				},
				"emails": {
					Type:        framework.TypeCommaStringSlice,
					Description: "Emails for rotation result notifications.",
				},
			},
			ExistenceCheck: b.connectionExistenceCheck(),
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.connectionWriteHandler(),
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.connectionWriteHandler(),
				},
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.connectionReadHandler(),
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.connectionDeleteHandler(),
				},
			},
		},
	}
}

func (b *databaseBackend) connectionWriteHandler() framework.OperationFunc {
	return func(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
		name := data.Get("name").(string)
		if name == "" {
			return logical.ErrorResponse("missing name"), nil
		}
		verifyConn := data.Get("verify_connection").(bool)

		// Read existing config
		config := &storage.DatabaseConfig{}
		entry, err := req.Storage.Get(ctx, fmt.Sprintf("config/%s", name))
		if err != nil {
			return nil, err
		}
		if entry != nil {
			if err := entry.DecodeJSON(config); err != nil {
				return nil, err
			}
		}

		// Overwrite fields if in request
		if pluginNameRaw, ok := data.GetOk("plugin_name"); ok {
			config.PluginName = pluginNameRaw.(string)
		}
		if pluginVerRaw, ok := data.GetOk("plugin_version"); ok {
			config.PluginVersion = pluginVerRaw.(string)
		}
		if rolesRaw, ok := data.GetOk("allowed_roles"); ok {
			config.AllowedRoles = rolesRaw.([]string)
		}
		if emailsRaw, ok := data.GetOk("emails"); ok {
			config.Emails = emailsRaw.([]string)
		}
		if rootStmtRaw, ok := data.GetOk("root_rotation_statements"); ok {
			config.RootCredentialsRotateStatements = rootStmtRaw.([]string)
		}
		if pwdPolicyRaw, ok := data.GetOk("password_policy"); ok {
			config.PasswordPolicy = pwdPolicyRaw.(string)
		}
		if ciName, ok := data.GetOk("ci_name"); ok {
			config.CiName = ciName.(string)
		}

		// Sanitize framework data to store only custom DB fields
		delete(data.Raw, "name")
		delete(data.Raw, "plugin_name")
		delete(data.Raw, "plugin_version")
		delete(data.Raw, "allowed_roles")
		delete(data.Raw, "root_rotation_statements")
		delete(data.Raw, "password_policy")
		delete(data.Raw, "verify_connection")

		// Store remaining fields as ConnectionDetails
		if config.ConnectionDetails == nil {
			config.ConnectionDetails = make(map[string]interface{})
		}
		for k, v := range data.Raw {
			config.ConnectionDetails[k] = v
		}

		// Load typed config from map
		engine, err := dbengines.New(config.PluginName, config.ConnectionDetails)
		if err != nil {
			return logical.ErrorResponse(fmt.Sprintf("invalid plugin config: %s", err)), nil
		}
		log.Print("About to verify")
		// Test DB connection
		if verifyConn {
			if _, err := engine.Connect(ctx); err != nil {
				return logical.ErrorResponse(fmt.Sprintf("connection failed: %s", err)), nil
			}
		}

		// Save config
		if err := b.storeConfig(ctx, req.Storage, name, config); err != nil {
			return nil, err
		}

		old := b.conn.Put(name, engine)
		if old != nil {
			old.Close()
		}

		// 1.12.0 and 1.12.1 stored builtin plugins in storage, but 1.12.2 reverted
		// that, so clean up any pre-existing stored builtin versions on write.
		if versions.IsBuiltinVersion(config.PluginVersion) {
			config.PluginVersion = ""
		}

		resp := &logical.Response{}
		if parsed, ok := config.ConnectionDetails["connection_url"].(string); ok {
			if u, err := url.Parse(parsed); err == nil && u.User != nil {
				if _, ok := u.User.Password(); ok {
					resp.AddWarning("Password found in connection_url, use a templated URL to avoid password leak.")
				}
			}
		}
		if len(resp.Warnings) == 0 {
			return nil, nil
		}
		return resp, nil
	}
}

func (b *databaseBackend) connectionReadHandler() framework.OperationFunc {
	return func(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
		name := data.Get("name").(string)
		entry, err := req.Storage.Get(ctx, fmt.Sprintf("config/%s", name))
		if err != nil || entry == nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		var cfg storage.DatabaseConfig
		if err := entry.DecodeJSON(&cfg); err != nil {
			return nil, err
		}

		// redact password fields if any
		delete(cfg.ConnectionDetails, "password")
		delete(cfg.ConnectionDetails, "private_key")

		return &logical.Response{
			Data: map[string]interface{}{
				"plugin_name":              cfg.PluginName,
				"plugin_version":           cfg.PluginVersion,
				"allowed_roles":            cfg.AllowedRoles,
				"emails":                   cfg.Emails,
				"ci_name":                  cfg.CiName,
				"password_policy":          cfg.PasswordPolicy,
				"root_rotation_statements": cfg.RootCredentialsRotateStatements,
				"connection_details":       cfg.ConnectionDetails,
			},
		}, nil
	}
}

func (b *databaseBackend) connectionDeleteHandler() framework.OperationFunc {
	return func(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
		name := data.Get("name").(string)
		if err := req.Storage.Delete(ctx, fmt.Sprintf("config/%s", name)); err != nil {
			return nil, err
		}
		b.conn.ClearConnection(name)
		return nil, nil
	}
}

func (b *databaseBackend) connectionExistenceCheck() framework.ExistenceFunc {
	return func(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
		name := data.Get("name").(string)
		entry, err := req.Storage.Get(ctx, fmt.Sprintf("config/%s", name))
		if err != nil || entry == nil {
			return false, nil
		}
		return true, nil
	}
}

func (b *databaseBackend) storeConfig(ctx context.Context, s logical.Storage, name string, config *storage.DatabaseConfig) error {
	entry, err := logical.StorageEntryJSON(fmt.Sprintf("config/%s", name), config)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	return s.Put(ctx, entry)
}
