// Package naming provides shared identifier, casing, and path sanitization utilities across AST emitters and parsers.
package naming

import (
	"strings"
)

// ToPascalCase converts a snake_case, kebab-case, or lowerCamelCase identifier into PascalCase.
// Parameter s is the input string, and it returns the PascalCase representation.
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}
	normalized := strings.ReplaceAll(s, "-", "_")
	parts := strings.Split(normalized, "_")
	var res string
	for _, p := range parts {
		if p != "" {
			res += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return res
}

// ToCamelCasePath converts an OpenAPI URL path (such as "/users/{id}") into a PascalCase identifier suffix (such as "UsersId" or "Root").
// Parameter s is the OpenAPI URL path, and it returns the sanitized PascalCase identifier.
func ToCamelCasePath(s string) string {
	parts := strings.Split(s, "/")
	var res string
	for _, p := range parts {
		p = strings.ReplaceAll(p, "{", "")
		p = strings.ReplaceAll(p, "}", "")
		if p != "" {
			res += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	if res == "" {
		res = "Root"
	}
	return res
}

// ToSnakeCase converts a PascalCase or camelCase string into snake_case.
// Parameter s is the input string, and it returns the snake_case representation.
func ToSnakeCase(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && c >= 'A' && c <= 'Z' {
			b.WriteRune('_')
		}
		b.WriteRune(c)
	}
	return strings.ToLower(b.String())
}

// ExportedName converts a JSON property name into an exported Go struct field name.
// Parameter name is the JSON field name, and it returns the exported Go identifier.
func ExportedName(name string) string {
	if name == "" {
		return ""
	}
	if strings.EqualFold(name, "id") {
		return "Id"
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// SanitizePathToFilename converts an OpenAPI URL path into a safe filename stem (such as "users_id" or "root").
// Parameter path is the OpenAPI URL path, and it returns the sanitized filename prefix.
func SanitizePathToFilename(path string) string {
	fileName := strings.ReplaceAll(path, "/", "_")
	fileName = strings.ReplaceAll(fileName, "{", "")
	fileName = strings.ReplaceAll(fileName, "}", "")
	fileName = strings.TrimPrefix(fileName, "_")
	if fileName == "" {
		fileName = "root"
	}
	return fileName
}
