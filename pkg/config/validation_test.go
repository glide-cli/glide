package config

import (
	"strings"
	"testing"
)

func TestValidator_Required(t *testing.T) {
	type Config struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required"`
		Age   int    `json:"age"`
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  Config{Name: "John", Email: "john@example.com", Age: 30},
			wantErr: false,
		},
		{
			name:    "missing required name",
			config:  Config{Email: "john@example.com"},
			wantErr: true,
		},
		{
			name:    "missing required email",
			config:  Config{Name: "John"},
			wantErr: true,
		},
		{
			name:    "missing both required fields",
			config:  Config{Age: 30},
			wantErr: true,
		},
		{
			name:    "missing optional age is ok",
			config:  Config{Name: "John", Email: "john@example.com"},
			wantErr: false,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_Min(t *testing.T) {
	type Config struct {
		Age      int    `json:"age" validate:"min=0"`
		Score    int    `json:"score" validate:"min=1,max=100"`
		Name     string `json:"name" validate:"min=2"`
		Tags     []string `json:"tags" validate:"min=1"`
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  Config{Age: 25, Score: 85, Name: "Jo", Tags: []string{"tag1"}},
			wantErr: false,
		},
		{
			name:    "age too low",
			config:  Config{Age: -1, Score: 85, Name: "Jo", Tags: []string{"tag1"}},
			wantErr: true,
		},
		{
			name:    "score too low",
			config:  Config{Age: 25, Score: 0, Name: "Jo", Tags: []string{"tag1"}},
			wantErr: true,
		},
		{
			name:    "name too short",
			config:  Config{Age: 25, Score: 85, Name: "J", Tags: []string{"tag1"}},
			wantErr: true,
		},
		{
			name:    "empty tags array",
			config:  Config{Age: 25, Score: 85, Name: "Jo", Tags: []string{}},
			wantErr: true,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_Max(t *testing.T) {
	type Config struct {
		Age      int    `json:"age" validate:"max=120"`
		Score    int    `json:"score" validate:"min=1,max=100"`
		Name     string `json:"name" validate:"max=50"`
		Tags     []string `json:"tags" validate:"max=5"`
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  Config{Age: 30, Score: 85, Name: "John", Tags: []string{"tag1", "tag2"}},
			wantErr: false,
		},
		{
			name:    "age too high",
			config:  Config{Age: 150, Score: 85, Name: "John", Tags: []string{"tag1"}},
			wantErr: true,
		},
		{
			name:    "score too high",
			config:  Config{Age: 30, Score: 101, Name: "John", Tags: []string{"tag1"}},
			wantErr: true,
		},
		{
			name:    "name too long",
			config:  Config{Age: 30, Score: 85, Name: strings.Repeat("a", 51), Tags: []string{"tag1"}},
			wantErr: true,
		},
		{
			name:    "too many tags",
			config:  Config{Age: 30, Score: 85, Name: "John", Tags: []string{"1", "2", "3", "4", "5", "6"}},
			wantErr: true,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_Enum(t *testing.T) {
	type Config struct {
		Role   string `json:"role" validate:"enum=admin|user|guest"`
		Status int    `json:"status" validate:"enum=0|1|2"`
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid role admin",
			config:  Config{Role: "admin", Status: 1},
			wantErr: false,
		},
		{
			name:    "valid role user",
			config:  Config{Role: "user", Status: 0},
			wantErr: false,
		},
		{
			name:    "valid role guest",
			config:  Config{Role: "guest", Status: 2},
			wantErr: false,
		},
		{
			name:    "invalid role",
			config:  Config{Role: "superadmin", Status: 1},
			wantErr: true,
		},
		{
			name:    "invalid status",
			config:  Config{Role: "admin", Status: 3},
			wantErr: true,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_MultipleErrors(t *testing.T) {
	type Config struct {
		Name  string `json:"name" validate:"required,min=2"`
		Age   int    `json:"age" validate:"required,min=0,max=120"`
		Email string `json:"email" validate:"required"`
	}

	// Config with multiple validation errors
	config := Config{
		Name: "J",   // Too short (min=2)
		Age:  -1,    // Too low (min=0)
		Email: "",   // Required but missing
	}

	validator := NewValidator()
	err := validator.Validate(config)

	if err == nil {
		t.Fatal("Expected validation errors, got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("Expected ValidationErrors, got %T", err)
	}

	// Should have at least 3 errors (name too short, age too low, email missing)
	if len(verrs) < 3 {
		t.Errorf("Expected at least 3 errors, got %d: %v", len(verrs), verrs)
	}
}

func TestValidator_NestedStructs(t *testing.T) {
	type Address struct {
		Street string `json:"street" validate:"required"`
		City   string `json:"city" validate:"required"`
		Zip    string `json:"zip" validate:"min=5,max=10"`
	}

	type Person struct {
		Name    string  `json:"name" validate:"required"`
		Address Address `json:"address"`
	}

	tests := []struct {
		name    string
		config  Person
		wantErr bool
	}{
		{
			name: "valid nested config",
			config: Person{
				Name: "John",
				Address: Address{
					Street: "123 Main St",
					City:   "Springfield",
					Zip:    "12345",
				},
			},
			wantErr: false,
		},
		{
			name: "missing nested required field",
			config: Person{
				Name: "John",
				Address: Address{
					Street: "123 Main St",
					// City missing
					Zip: "12345",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid nested validation",
			config: Person{
				Name: "John",
				Address: Address{
					Street: "123 Main St",
					City:   "Springfield",
					Zip:    "123", // Too short (min=5)
				},
			},
			wantErr: true,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If we expect an error and got one, check that field names include nesting
			if tt.wantErr && err != nil {
				errStr := err.Error()
				if !strings.Contains(errStr, "Address.") {
					t.Errorf("Expected nested field name in error, got: %s", errStr)
				}
			}
		})
	}
}

func TestValidateWithDefaults(t *testing.T) {
	type Config struct {
		Name    string `json:"name" validate:"required"`
		Timeout int    `json:"timeout" validate:"min=1"`
		Enabled bool   `json:"enabled"`
	}

	defaults := Config{
		Name:    "default-name",
		Timeout: 30,
		Enabled: true,
	}

	tests := []struct {
		name    string
		config  Config
		want    Config
		wantErr bool
	}{
		{
			name:   "empty config gets defaults",
			config: Config{},
			want: Config{
				Name:    "default-name",
				Timeout: 30,
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "partial config gets missing defaults",
			config: Config{
				Name: "custom-name",
			},
			want: Config{
				Name:    "custom-name",
				Timeout: 30,
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "full config keeps values",
			config: Config{
				Name:    "my-name",
				Timeout: 60,
				Enabled: false,
			},
			want: Config{
				Name:    "my-name",
				Timeout: 60,
				Enabled: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWithDefaults(&tt.config, defaults)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWithDefaults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.config.Name != tt.want.Name {
					t.Errorf("Name = %v, want %v", tt.config.Name, tt.want.Name)
				}
				if tt.config.Timeout != tt.want.Timeout {
					t.Errorf("Timeout = %v, want %v", tt.config.Timeout, tt.want.Timeout)
				}
				if tt.config.Enabled != tt.want.Enabled {
					t.Errorf("Enabled = %v, want %v", tt.config.Enabled, tt.want.Enabled)
				}
			}
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	// Test single error
	singleErr := ValidationErrors{
		ValidationError{
			Field:   "name",
			Value:   "",
			Rule:    "required",
			Message: "field is required but has zero value",
		},
	}

	errStr := singleErr.Error()
	if !strings.Contains(errStr, "name") {
		t.Errorf("Single error should contain field name, got: %s", errStr)
	}

	// Test multiple errors
	multiErr := ValidationErrors{
		ValidationError{Field: "name", Message: "name is required"},
		ValidationError{Field: "age", Message: "age must be positive"},
		ValidationError{Field: "email", Message: "email is invalid"},
	}

	errStr = multiErr.Error()
	if !strings.Contains(errStr, "3 validation errors") {
		t.Errorf("Multiple errors should show count, got: %s", errStr)
	}
	if !strings.Contains(errStr, "name") || !strings.Contains(errStr, "age") || !strings.Contains(errStr, "email") {
		t.Errorf("Multiple errors should list all fields, got: %s", errStr)
	}
}
