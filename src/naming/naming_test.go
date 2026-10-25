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
	if ToPascalCase("find pet by id") != "FindPetById" {
		t.Errorf("expected FindPetById, got %s", ToPascalCase("find pet by id"))
	}
	if ToPascalCase("list-searchable-fields") != "ListSearchableFields" {
		t.Errorf("expected ListSearchableFields, got %s", ToPascalCase("list-searchable-fields"))
	}
}

func TestToCamelCasePath(t *testing.T) {
	if ToCamelCasePath("/") != "Root" {
		t.Errorf("expected Root")
	}
	if ToCamelCasePath("/users/{id}") != "UsersId" {
		t.Errorf("expected UsersId")
	}
	if ToCamelCasePath("/2.0/repositories/{username}") != "2_0RepositoriesUsername" {
		t.Errorf("expected 2_0RepositoriesUsername, got %s", ToCamelCasePath("/2.0/repositories/{username}"))
	}
	if ToCamelCasePath("/user-items/{item-id}") != "UserItemsItemId" {
		t.Errorf("expected UserItemsItemId, got %s", ToCamelCasePath("/user-items/{item-id}"))
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
	if ExportedName("Field1_ok") != "Field1_ok" {
		t.Errorf("expected Field1_ok, got %s", ExportedName("Field1_ok"))
	}
	if ExportedName("user-name") != "UserName" {
		t.Errorf("expected UserName, got %s", ExportedName("user-name"))
	}
	if ExportedName("2.0") != "Field20" {
		t.Errorf("expected Field20, got %s", ExportedName("2.0"))
	}
	if ExportedName("---") != "Field" {
		t.Errorf("expected Field, got %s", ExportedName("---"))
	}
	if isValidIdentName("") {
		t.Errorf("expected false for empty ident")
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
