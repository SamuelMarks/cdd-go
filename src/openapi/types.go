// Package openapi provides structures and utilities for parsing and emitting OpenAPI descriptions.
// It complies with the OpenAPI Specification Version 3.2.0.
package openapi

import "encoding/json"

// OpenAPI is the root object of the OpenAPI Description.
type OpenAPI struct {
	// OpenAPI is the version number of the OpenAPI Specification that the OpenAPI document uses.
	OpenAPI string `json:"openapi"`

	// Self provides the self-assigned URI of this document.
	Self string `json:"$self,omitempty"`

	// Info provides metadata about the API.
	Info Info `json:"info"`

	// JSONSchemaDialect is the default value for the $schema keyword within Schema Objects.
	JSONSchemaDialect string `json:"jsonSchemaDialect,omitempty"`

	// Servers is an array of Server Objects.
	Servers []Server `json:"servers,omitempty"`

	// Paths is the available paths and operations for the API.
	Paths Paths `json:"paths,omitempty"`

	// Webhooks are incoming webhooks that MAY be received.
	Webhooks map[string]PathItem `json:"webhooks,omitempty"`

	// Components holds various reusable Objects for the OpenAPI Description.
	Components *Components `json:"components,omitempty"`

	// Security is a declaration of which security mechanisms can be used across the API.
	Security []SecurityRequirement `json:"security,omitempty"`

	// Tags is a list of tags used by the OpenAPI Description with additional metadata.
	Tags []Tag `json:"tags,omitempty"`

	// ExternalDocs provides additional external documentation.
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
}

// Info provides metadata about the API.
type Info struct {
	// Title of the API.
	Title string `json:"title"`

	// Summary is a short summary of the API.
	Summary string `json:"summary,omitempty"`

	// Description is a description of the API.
	Description string `json:"description,omitempty"`

	// TermsOfService is a URI for the Terms of Service for the API.
	TermsOfService string `json:"termsOfService,omitempty"`

	// Contact is the contact information for the exposed API.
	Contact *Contact `json:"contact,omitempty"`

	// License is the license information for the exposed API.
	License *License `json:"license,omitempty"`

	// Version is the version of the OpenAPI document.
	Version string `json:"version"`
}

// Contact contains contact information for the exposed API.
type Contact struct {
	// Name is the identifying name of the contact person/organization.
	Name string `json:"name,omitempty"`

	// URL is the URI for the contact information.
	URL string `json:"url,omitempty"`

	// Email is the email address of the contact person/organization.
	Email string `json:"email,omitempty"`
}

// License contains license information for the exposed API.
type License struct {
	// Name is the license name used for the API.
	Name string `json:"name"`

	// Identifier is an SPDX license expression for the API.
	Identifier string `json:"identifier,omitempty"`

	// URL is a URI for the license used for the API.
	URL string `json:"url,omitempty"`
}

// Server represents a Server.
type Server struct {
	// URL to the target host.
	URL string `json:"url"`

	// Description is an optional string describing the host.
	Description string `json:"description,omitempty"`

	// Name is an optional unique string to refer to the host.
	Name string `json:"name,omitempty"`

	// Variables is a map between a variable name and its value.
	Variables map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable represents a Server Variable for server URL template substitution.
type ServerVariable struct {
	// Enum is an enumeration of string values to be used if the substitution options are from a limited set.
	Enum []string `json:"enum,omitempty"`

	// Default is the default value to use for substitution.
	Default string `json:"default"`

	// Description is an optional description for the server variable.
	Description string `json:"description,omitempty"`
}

// Components holds a set of reusable objects for different aspects of the OAS.
type Components struct {
	// Schemas holds reusable Schema Objects.
	Schemas map[string]Schema `json:"schemas,omitempty"`

	// Responses holds reusable Response Objects.
	Responses map[string]Response `json:"responses,omitempty"`

	// Parameters holds reusable Parameter Objects.
	Parameters map[string]Parameter `json:"parameters,omitempty"`

	// Examples holds reusable Example Objects.
	Examples map[string]Example `json:"examples,omitempty"`

	// RequestBodies holds reusable Request Body Objects.
	RequestBodies map[string]RequestBody `json:"requestBodies,omitempty"`

	// Headers holds reusable Header Objects.
	Headers map[string]Header `json:"headers,omitempty"`

	// SecuritySchemes holds reusable Security Scheme Objects.
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`

	// Links holds reusable Link Objects.
	Links map[string]Link `json:"links,omitempty"`

	// Callbacks holds reusable Callback Objects.
	Callbacks map[string]Callback `json:"callbacks,omitempty"`

	// PathItems holds reusable Path Item Objects.
	PathItems map[string]PathItem `json:"pathItems,omitempty"`

	// MediaTypes holds reusable Media Type Objects.
	MediaTypes map[string]MediaType `json:"mediaTypes,omitempty"`
}

// Paths holds the relative paths to the individual endpoints and their operations.
type Paths map[string]PathItem

// PathItem describes the operations available on a single path.
type PathItem struct {
	// Ref allows for a referenced definition of this path item.
	Ref string `json:"$ref,omitempty"`

	// Summary is an optional string summary, intended to apply to all operations in this path.
	Summary string `json:"summary,omitempty"`

	// Description is an optional string description, intended to apply to all operations in this path.
	Description string `json:"description,omitempty"`

	// Get is a definition of a GET operation on this path.
	Get *Operation `json:"get,omitempty"`

	// Put is a definition of a PUT operation on this path.
	Put *Operation `json:"put,omitempty"`

	// Post is a definition of a POST operation on this path.
	Post *Operation `json:"post,omitempty"`

	// Delete is a definition of a DELETE operation on this path.
	Delete *Operation `json:"delete,omitempty"`

	// Options is a definition of a OPTIONS operation on this path.
	Options *Operation `json:"options,omitempty"`

	// Head is a definition of a HEAD operation on this path.
	Head *Operation `json:"head,omitempty"`

	// Patch is a definition of a PATCH operation on this path.
	Patch *Operation `json:"patch,omitempty"`

	// Trace is a definition of a TRACE operation on this path.
	Trace *Operation `json:"trace,omitempty"`

	// Query is a definition of a QUERY operation on this path.
	Query *Operation `json:"query,omitempty"`

	// AdditionalOperations is a map of additional operations on this path.
	AdditionalOperations map[string]Operation `json:"additionalOperations,omitempty"`

	// Servers is an alternative servers array to service all operations in this path.
	Servers []Server `json:"servers,omitempty"`

	// Parameters is a list of parameters that are applicable for all the operations described under this path.
	Parameters []Parameter `json:"parameters,omitempty"`
}

// Operation describes a single API operation on a path.
type Operation struct {
	// Tags is a list of tags for API documentation control.
	Tags []string `json:"tags,omitempty"`

	// Summary is a short summary of what the operation does.
	Summary string `json:"summary,omitempty"`

	// Description is a verbose explanation of the operation behavior.
	Description string `json:"description,omitempty"`

	// ExternalDocs holds additional external documentation for this operation.
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`

	// OperationID is a unique string used to identify the operation.
	OperationID string `json:"operationId,omitempty"`

	// Parameters is a list of parameters that are applicable for this operation.
	Parameters []Parameter `json:"parameters,omitempty"`

	// RequestBody is the request body applicable for this operation.
	RequestBody *RequestBody `json:"requestBody,omitempty"`

	// Responses is the list of possible responses as they are returned from executing this operation.
	Responses Responses `json:"responses"`

	// Callbacks is a map of possible out-of band callbacks related to the parent operation.
	Callbacks map[string]Callback `json:"callbacks,omitempty"`

	// Deprecated declares this operation to be deprecated.
	Deprecated bool `json:"deprecated,omitempty"`

	// Security is a declaration of which security mechanisms can be used for this operation.
	Security []SecurityRequirement `json:"security,omitempty"`

	// Servers is an alternative servers array to service this operation.
	Servers []Server `json:"servers,omitempty"`
}

// ExternalDocs allows referencing an external resource for extended documentation.
type ExternalDocs struct {
	// Description is a description of the target documentation.
	Description string `json:"description,omitempty"`

	// URL is the URI for the target documentation.
	URL string `json:"url"`
}

// Parameter describes a single operation parameter.
type Parameter struct {
	// Ref allows for a referenced definition of this parameter.
	Ref string `json:"$ref,omitempty"`

	// Name is the name of the parameter.
	Name string `json:"name,omitempty"`

	// In is the location of the parameter.
	In string `json:"in,omitempty"`

	// Description is a brief description of the parameter.
	Description string `json:"description,omitempty"`

	// Required determines whether this parameter is mandatory.
	Required bool `json:"required,omitempty"`

	// Deprecated specifies that a parameter is deprecated and SHOULD be transitioned out of usage.
	Deprecated bool `json:"deprecated,omitempty"`

	// AllowEmptyValue specifies if clients MAY pass a zero-length string value.
	AllowEmptyValue bool `json:"allowEmptyValue,omitempty"`

	// Example is an example of the parameter's potential value.
	Example json.RawMessage `json:"example,omitempty"`

	// Examples are examples of the parameter's potential value.
	Examples map[string]Example `json:"examples,omitempty"`

	// Style describes how the parameter value will be serialized.
	Style string `json:"style,omitempty"`

	// Explode indicates whether parameter values of type array or object generate separate parameters.
	Explode bool `json:"explode,omitempty"`

	// AllowReserved indicates whether parameter values are serialized using reserved expansion.
	AllowReserved bool `json:"allowReserved,omitempty"`

	// Schema is the schema defining the type used for the parameter.
	Schema *Schema `json:"schema,omitempty"`

	// Content is a map containing the representations for the parameter.
	Content map[string]MediaType `json:"content,omitempty"`
}

// RequestBody describes a single request body.
type RequestBody struct {
	// Ref allows for a referenced definition of this request body.
	Ref string `json:"$ref,omitempty"`

	// Description is a brief description of the request body.
	Description string `json:"description,omitempty"`

	// Content is the content of the request body.
	Content map[string]MediaType `json:"content"`

	// Required determines if the request body is required in the request.
	Required bool `json:"required,omitempty"`
}

// MediaType describes content structured in accordance with the media type identified by its key.
type MediaType struct {
	// Ref allows for a referenced definition.
	Ref string `json:"$ref,omitempty"`

	// Schema is a schema describing the complete content.
	Schema *Schema `json:"schema,omitempty"`

	// ItemSchema is a schema describing each item within a sequential media type.
	ItemSchema *Schema `json:"itemSchema,omitempty"`

	// Example is an example of the media type.
	Example json.RawMessage `json:"example,omitempty"`

	// Examples are examples of the media type.
	Examples map[string]Example `json:"examples,omitempty"`

	// Encoding is a map between a property name and its encoding information.
	Encoding map[string]Encoding `json:"encoding,omitempty"`

	// PrefixEncoding is an array of positional encoding information.
	PrefixEncoding []Encoding `json:"prefixEncoding,omitempty"`

	// ItemEncoding is a single Encoding Object that provides encoding information for multiple array items.
	ItemEncoding *Encoding `json:"itemEncoding,omitempty"`
}

// Encoding represents a single encoding definition applied to a single value.
type Encoding struct {
	// ContentType is the Content-Type for encoding a specific property.
	ContentType string `json:"contentType,omitempty"`

	// Headers is a map allowing additional information to be provided as headers.
	Headers map[string]Header `json:"headers,omitempty"`

	// Encoding applies nested Encoding Objects in the same manner as the MediaType Object's encoding field.
	Encoding map[string]Encoding `json:"encoding,omitempty"`

	// PrefixEncoding applies nested Encoding Objects in the same manner as the MediaType Object's prefixEncoding field.
	PrefixEncoding []Encoding `json:"prefixEncoding,omitempty"`

	// ItemEncoding applies nested Encoding Objects in the same manner as the MediaType Object's itemEncoding field.
	ItemEncoding *Encoding `json:"itemEncoding,omitempty"`

	// Style describes how a specific property value will be serialized.
	Style string `json:"style,omitempty"`

	// Explode indicates whether property values of type array or object generate separate parameters.
	Explode bool `json:"explode,omitempty"`

	// AllowReserved indicates whether parameter values are serialized using reserved expansion.
	AllowReserved bool `json:"allowReserved,omitempty"`
}

// Schema represents the schema object, a generic implementation here.
type Schema struct {
	// Ref allows for a referenced definition.
	Ref string `json:"$ref,omitempty"`
	// Type specifies the data type.
	Type string `json:"type,omitempty"`
	// Properties define the structure of the object.
	Properties map[string]Schema `json:"properties,omitempty"`
	// Items define the item schema for array types.
	Items *Schema `json:"items,omitempty"`
	// AdditionalProperties dictate validation rules for extraneous fields.
	AdditionalProperties *Schema `json:"additionalProperties,omitempty"`
	// Description provides context for the schema.
	Description string `json:"description,omitempty"`
	// Format suggests data format rules (e.g., date-time).
	Format string `json:"format,omitempty"`
	// Required specifies which object properties must be present.
	Required []string `json:"required,omitempty"`
	// Discriminator adds support for polymorphism.
	Discriminator *Discriminator `json:"discriminator,omitempty"`
	// XML adds XML specific properties.
	XML *XML `json:"xml,omitempty"`
	// Default specifies the default value.
	Default json.RawMessage `json:"default,omitempty"`
	// Example provides an example of the schema.
	Example json.RawMessage `json:"example,omitempty"`
	// Examples provides multiple examples.
	Examples []json.RawMessage `json:"examples,omitempty"`
	// Title is a short summary of the schema.
	Title string `json:"title,omitempty"`
	// MultipleOf specifies a multiple of which the value must be.
	MultipleOf *float64 `json:"multipleOf,omitempty"`
	// Maximum specifies the maximum value.
	Maximum *float64 `json:"maximum,omitempty"`
	// ExclusiveMaximum specifies if the maximum is exclusive.
	ExclusiveMaximum *bool `json:"exclusiveMaximum,omitempty"`
	// Minimum specifies the minimum value.
	Minimum *float64 `json:"minimum,omitempty"`
	// ExclusiveMinimum specifies if the minimum is exclusive.
	ExclusiveMinimum *bool `json:"exclusiveMinimum,omitempty"`
	// MaxLength specifies the maximum length of a string.
	MaxLength *int `json:"maxLength,omitempty"`
	// MinLength specifies the minimum length of a string.
	MinLength *int `json:"minLength,omitempty"`
	// Pattern specifies a regular expression pattern the string must match.
	Pattern string `json:"pattern,omitempty"`
	// MaxItems specifies the maximum number of items in an array.
	MaxItems *int `json:"maxItems,omitempty"`
	// MinItems specifies the minimum number of items in an array.
	MinItems *int `json:"minItems,omitempty"`
	// UniqueItems specifies if items in an array must be unique.
	UniqueItems *bool `json:"uniqueItems,omitempty"`
	// MaxProperties specifies the maximum number of properties in an object.
	MaxProperties *int `json:"maxProperties,omitempty"`
	// MinProperties specifies the minimum number of properties in an object.
	MinProperties *int `json:"minProperties,omitempty"`
	// Enum specifies the possible values.
	Enum []json.RawMessage `json:"enum,omitempty"`
	// AllOf requires all subschemas to be valid.
	AllOf []Schema `json:"allOf,omitempty"`
	// OneOf requires exactly one subschema to be valid.
	OneOf []Schema `json:"oneOf,omitempty"`
	// AnyOf requires at least one subschema to be valid.
	AnyOf []Schema `json:"anyOf,omitempty"`
	// Not requires the subschema to be invalid.
	Not *Schema `json:"not,omitempty"`
	// Nullable specifies if the value can be null.
	Nullable *bool `json:"nullable,omitempty"`
	// ReadOnly specifies if the property is read-only.
	ReadOnly *bool `json:"readOnly,omitempty"`
	// WriteOnly specifies if the property is write-only.
	WriteOnly *bool `json:"writeOnly,omitempty"`
	// Deprecated specifies if the schema is deprecated.
	Deprecated *bool `json:"deprecated,omitempty"`
}

// Discriminator Object for polymorphism support
type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty"`
}

// XML Object for XML representations
type XML struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Attribute bool   `json:"attribute,omitempty"`
	Wrapped   bool   `json:"wrapped,omitempty"`
}

// Response represents a single response from an API Operation.
type Response struct {
	// Ref allows for a referenced definition.
	Ref string `json:"$ref,omitempty"`
	// Description of the response.
	Description string `json:"description,omitempty"`
	// Content contains the possible payload media types.
	Content map[string]MediaType `json:"content,omitempty"`
	// Headers expected for this response.
	Headers map[string]Header `json:"headers,omitempty"`
}

// Responses holds the possible responses for an Operation.
type Responses map[string]Response

// Example represents an example of a parameter, property, or response payload.
type Example struct {
	// Ref allows for a referenced definition.
	Ref string `json:"$ref,omitempty"`
	// Summary is a short description of the example.
	Summary string `json:"summary,omitempty"`
	// Description provides more details about the example.
	Description string `json:"description,omitempty"`
	// Value holds the literal example value.
	Value json.RawMessage `json:"value,omitempty"`
	// ExternalValue points to a URI that contains the literal example value.
	ExternalValue string `json:"externalValue,omitempty"`
}

// Header describes a single HTTP header.
type Header struct {
	// Ref allows for a referenced definition.
	Ref string `json:"$ref,omitempty"`
	// Description briefly describes the header.
	Description string `json:"description,omitempty"`
	// Required indicates if this header must be sent.
	Required bool `json:"required,omitempty"`
	// Deprecated specifies that a header is deprecated and SHOULD be transitioned out of usage.
	Deprecated bool `json:"deprecated,omitempty"`
	// AllowEmptyValue specifies if clients MAY pass a zero-length string value.
	AllowEmptyValue bool `json:"allowEmptyValue,omitempty"`
	// Style describes how the header value will be serialized.
	Style string `json:"style,omitempty"`
	// Explode indicates whether header values of type array or object generate separate parameters.
	Explode bool `json:"explode,omitempty"`
	// AllowReserved indicates whether header values are serialized using reserved expansion.
	AllowReserved bool `json:"allowReserved,omitempty"`
	// Schema is the schema defining the type used for the header.
	Schema *Schema `json:"schema,omitempty"`
	// Content is a map containing the representations for the header.
	Content map[string]MediaType `json:"content,omitempty"`
	// Example is an example of the header's potential value.
	Example json.RawMessage `json:"example,omitempty"`
	// Examples are examples of the header's potential value.
	Examples map[string]Example `json:"examples,omitempty"`
}

// SecurityScheme defines a security scheme that can be used by the operations.
type SecurityScheme struct {
	// Ref allows for a referenced definition.
	Ref string `json:"$ref,omitempty"`
	// Type of the security scheme.
	Type string `json:"type,omitempty"`
	// Description of the security scheme.
	Description string `json:"description,omitempty"`
	// Name of the header, query or cookie parameter to be used.
	Name string `json:"name,omitempty"`
	// In is the location of the API key.
	In string `json:"in,omitempty"`
	// Scheme is the name of the HTTP Authorization scheme to be used.
	Scheme string `json:"scheme,omitempty"`
	// BearerFormat hints at the format of the bearer token.
	BearerFormat string `json:"bearerFormat,omitempty"`
	// Flows contains configuration information for the flow types supported.
	Flows *OAuthFlows `json:"flows,omitempty"`
	// OpenIDConnectURL is the URL to discover OAuth2 configuration values.
	OpenIDConnectURL string `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows contains configuration information for the supported flow types.
type OAuthFlows struct {
	// Implicit configuration for the OAuth Implicit flow.
	Implicit *OAuthFlow `json:"implicit,omitempty"`
	// Password configuration for the OAuth Resource Owner Password flow.
	Password *OAuthFlow `json:"password,omitempty"`
	// ClientCredentials configuration for the OAuth Client Credentials flow.
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	// AuthorizationCode configuration for the OAuth Authorization Code flow.
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow contains configuration information for a single flow.
type OAuthFlow struct {
	// AuthorizationURL is the authorization URL to be used for this flow.
	AuthorizationURL string `json:"authorizationUrl,omitempty"`
	// TokenURL is the token URL to be used for this flow.
	TokenURL string `json:"tokenUrl,omitempty"`
	// RefreshURL is the URL to be used for obtaining refresh tokens.
	RefreshURL string `json:"refreshUrl,omitempty"`
	// Scopes provides the available scopes for the OAuth2 security scheme.
	Scopes map[string]string `json:"scopes,omitempty"`
}

// Link represents a possible design-time link for a response.
type Link struct {
	// Ref allows for a referenced definition.
	Ref string `json:"$ref,omitempty"`
	// OperationRef is a relative or absolute URI reference to an OAS operation.
	OperationRef string `json:"operationRef,omitempty"`
	// OperationID is the name of an existing, resolvable OAS operation.
	OperationID string `json:"operationId,omitempty"`
	// Parameters is a map representing parameters to pass to an operation as specified with operationId or identified via operationRef.
	Parameters map[string]string `json:"parameters,omitempty"`
	// RequestBody is a literal value or expression to use as a request body when calling the target operation.
	RequestBody interface{} `json:"requestBody,omitempty"`
	// Description describes the link purpose.
	Description string `json:"description,omitempty"`
	// Server is a server object to be used by the target operation.
	Server *Server `json:"server,omitempty"`
}

// Callback is a map of possible out-of band callbacks related to the parent operation.
type Callback map[string]PathItem

// SecurityRequirement is a declaration of which security mechanisms can be used across the API.
type SecurityRequirement map[string][]string

// Tag adds metadata to a single tag that is used by the Operation Object.
type Tag struct {
	// Name of the tag.
	Name string `json:"name"`
	// Description of the tag.
	Description string `json:"description,omitempty"`
	// ExternalDocs adds additional external documentation for this tag.
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
}
