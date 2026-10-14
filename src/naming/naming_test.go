package naming

import "testing"

func TestToPascalCase(t *testing.T) {
	if ToPascalCase("") != "" {
		t.Errorf("expected empty string")
	}
	if ToPascalCase("user_id") != "UserId" {
		t.Errorf("expected UserId, got %s", ToPascalCase("user_id"))
	}
	if ToPascalCase("_user-id_") != "UserId" {
		t.Errorf("expected UserId, got %s", ToPascalCase("_user-id_"))
	}
	if ToPascalCase("Test") != "Test" {
		t.Errorf("expected Test")
	}
}

func TestToCamelCasePath(t *testing.T) {
	if ToCamelCasePath("/") != "Root" {
		t.Errorf("expected Root")
	}
	if ToCamelCasePath("/users/{id}") != "UsersId" {
		t.Errorf("expected UsersId")
	}
}

func TestToSnakeCase(t *testing.T) {
	if ToSnakeCase("") != "" {
		t.Errorf("expected empty string")
	}
	if ToSnakeCase("GetUserById") != "get_user_by_id" {
		t.Errorf("expected get_user_by_id, got %s", ToSnakeCase("GetUserById"))
	}
}

func TestExportedName(t *testing.T) {
	if ExportedName("") != "" {
		t.Errorf("expected empty string")
	}
	if ExportedName("id") != "Id" {
		t.Errorf("expected Id, got %s", ExportedName("id"))
	}
	if ExportedName("username") != "Username" {
		t.Errorf("expected Username, got %s", ExportedName("username"))
	}
}

func TestSanitizePathToFilename(t *testing.T) {
	if SanitizePathToFilename("/") != "root" {
		t.Errorf("expected root, got %s", SanitizePathToFilename("/"))
	}
	if SanitizePathToFilename("/users/{id}") != "users_id" {
		t.Errorf("expected users_id, got %s", SanitizePathToFilename("/users/{id}"))
	}
}
