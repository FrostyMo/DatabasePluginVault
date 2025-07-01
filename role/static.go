package role

//
//import (
//	"DatabasePluginVault/storage"
//	"context"
//	"encoding/json"
//	"fmt"
//)
//
//type StaticRole struct {
//	Name           string   `json:"name"`
//	DBType         string   `json:"db_type"`
//	ConnectionName string   `json:"connection_name"`
//	Username       string   `json:"username"`
//	PasswordPolicy string   `json:"password_policy,omitempty"`
//	RotationSQL    []string `json:"rotation_statements"`
//}
//
//func (r *StaticRole) Validate() error {
//	if r.Name == "" {
//		return fmt.Errorf("role name cannot be empty")
//	}
//	if r.Username == "" {
//		return fmt.Errorf("username is required")
//	}
//	if r.DBType == "" {
//		return fmt.Errorf("db_type is required")
//	}
//	if r.ConnectionName == "" {
//		return fmt.Errorf("connection_name is required")
//	}
//	return nil
//}
//
//func CreateOrUpdateStaticRole(ctx context.Context, s storage.Storage, role *StaticRole) error {
//	if err := role.Validate(); err != nil {
//		return err
//	}
//	path := fmt.Sprintf("roles/%s/%s", role.DBType, role.Name)
//	return s.SaveRaw(ctx, path, role)
//}
//
//func GetStaticRole(ctx context.Context, s storage.Storage, dbType, name string) (*StaticRole, error) {
//	path := fmt.Sprintf("roles/%s/%s", dbType, name)
//	raw, err := s.LoadRaw(ctx, path)
//	if err != nil {
//		return nil, err
//	}
//	var r StaticRole
//	if err := json.Unmarshal(raw, &r); err != nil {
//		return nil, err
//	}
//	return &r, nil
//}
//
//func DeleteStaticRole(ctx context.Context, s storage.Storage, dbType, name string) error {
//	path := fmt.Sprintf("roles/%s/%s", dbType, name)
//	return s.DeleteRaw(ctx, path)
//}
