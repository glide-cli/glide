package sdk

// ConfigSchema defines the configuration schema for a plugin
type ConfigSchema struct {
	// Name is the unique identifier for this config section
	// This will be the key under which plugin config appears in .glide.yml
	Name string

	// Fields defines the configuration fields expected by this plugin
	Fields []FieldSchema

	// Description provides documentation for this config section
	Description string

	// Required indicates if this config section must be present
	Required bool
}

// FieldSchema defines a single configuration field
type FieldSchema struct {
	// Name is the field name as it appears in config
	Name string

	// Type describes the expected data type (string, bool, int, array, object, etc.)
	Type string

	// Description provides documentation for this field
	Description string

	// Required indicates if this field must be present
	Required bool

	// Default provides a default value if the field is not specified
	Default interface{}

	// Validation provides validation rules (e.g., "must be positive", "valid path")
	Validation string

	// Nested fields for complex types like objects
	Nested []FieldSchema
}

// ConfigProvider is the interface plugins implement to provide configuration schema
type ConfigProvider interface {
	// ProvideConfigSchema returns the configuration schema for this plugin
	// Returns nil if the plugin does not require configuration
	ProvideConfigSchema() *ConfigSchema
}

// ValidateConfig validates configuration data against a schema
func ValidateConfig(schema *ConfigSchema, data map[string]interface{}) []ValidationError {
	var errors []ValidationError

	// Check required schema
	if schema.Required && data == nil {
		errors = append(errors, ValidationError{
			Field:   schema.Name,
			Message: "required configuration section is missing",
		})
		return errors
	}

	if data == nil {
		return errors
	}

	// Validate each field
	for _, field := range schema.Fields {
		value, exists := data[field.Name]

		// Check required fields
		if field.Required && !exists {
			errors = append(errors, ValidationError{
				Field:   field.Name,
				Message: "required field is missing",
			})
			continue
		}

		// Skip validation if field doesn't exist and isn't required
		if !exists {
			continue
		}

		// Type validation
		if !validateType(field.Type, value) {
			errors = append(errors, ValidationError{
				Field:   field.Name,
				Message: "invalid type: expected " + field.Type,
			})
		}

		// Validate nested fields for objects
		if field.Type == "object" && len(field.Nested) > 0 {
			if objValue, ok := value.(map[string]interface{}); ok {
				nestedSchema := &ConfigSchema{
					Name:   field.Name,
					Fields: field.Nested,
				}
				nestedErrors := ValidateConfig(nestedSchema, objValue)
				for _, err := range nestedErrors {
					err.Field = field.Name + "." + err.Field
					errors = append(errors, err)
				}
			}
		}
	}

	return errors
}

// validateType checks if a value matches the expected type
func validateType(expectedType string, value interface{}) bool {
	if value == nil {
		return true
	}

	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "bool":
		_, ok := value.(bool)
		return ok
	case "int":
		_, ok := value.(int)
		if ok {
			return true
		}
		// Also accept float64 that represents an integer (JSON unmarshaling)
		if f, ok := value.(float64); ok {
			return f == float64(int(f))
		}
		return false
	case "float":
		_, ok := value.(float64)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	default:
		// Unknown type, allow it
		return true
	}
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ApplyDefaults applies default values from schema to config data
func ApplyDefaults(schema *ConfigSchema, data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = make(map[string]interface{})
	}

	for _, field := range schema.Fields {
		if _, exists := data[field.Name]; !exists && field.Default != nil {
			data[field.Name] = field.Default
		}

		// Apply defaults to nested objects
		if field.Type == "object" && len(field.Nested) > 0 {
			if objValue, ok := data[field.Name].(map[string]interface{}); ok {
				nestedSchema := &ConfigSchema{
					Name:   field.Name,
					Fields: field.Nested,
				}
				data[field.Name] = ApplyDefaults(nestedSchema, objValue)
			}
		}
	}

	return data
}
