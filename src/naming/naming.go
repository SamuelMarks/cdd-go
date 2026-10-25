// Package naming provides shared identifier, casing, and path sanitization utilities across AST emitters and parsers.
package naming

import (
	"strings"
)

// isASCIIAlnum reports whether r is an ASCII letter or decimal digit.
// Parameter r is the rune to inspect, and it returns true if alphanumeric.
func isASCIIAlnum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// isValidIdentName reports whether s starts with an ASCII letter and contains only ASCII letters, digits, or underscores.
// Parameter s is the candidate identifier string, and it returns true if valid.
func isValidIdentName(s string) bool {
	if s == "" {
		return false
	}
	first := s[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		return false
	}
	for _, r := range s {
		if !isASCIIAlnum(r) && r != '_' {
			return false
		}
	}
	return true
}

// ToPascalCase converts a snake_case, kebab-case, space-separated, or lowerCamelCase identifier into PascalCase.
// Parameter s is the input string, and it returns the PascalCase representation.
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return !isASCIIAlnum(r)
	})
	var res string
	for _, p := range parts {
		res += strings.ToUpper(p[:1]) + p[1:]
	}
	return res
}

// ToCamelCasePath converts an OpenAPI URL path (such as "/users/{id}" or "/2.0/repositories/{username}") into a PascalCase identifier suffix (such as "UsersId", "2_0RepositoriesUsername", or "Root").
// Parameter s is the OpenAPI URL path, and it returns the sanitized PascalCase identifier.
func ToCamelCasePath(s string) string {
	parts := strings.Split(s, "/")
	var res string
	for _, p := range parts {
		p = strings.ReplaceAll(p, "{", "")
		p = strings.ReplaceAll(p, "}", "")
		p = strings.Trim(strings.ReplaceAll(p, ".", "_"), "_")
		tokens := strings.FieldsFunc(p, func(r rune) bool {
			return !isASCIIAlnum(r) && r != '_'
		})
		for _, tok := range tokens {
			res += strings.ToUpper(tok[:1]) + tok[1:]
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
	if isValidIdentName(name) {
		return strings.ToUpper(name[:1]) + name[1:]
	}
	pascal := ToPascalCase(name)
	if pascal == "" || (pascal[0] >= '0' && pascal[0] <= '9') {
		return "Field" + pascal
	}
	return pascal
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
