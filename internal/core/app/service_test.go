package app

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

type MockCustomer struct {
	ID               int64
	FirstName        pgtype.Text
	LastName         pgtype.Text
	Location         pgtype.Text
	PreferredProduct pgtype.Text
	Phone            string
	CreatedAt        pgtype.Timestamp
}

type SimpleStruct struct {
	Name  string
	Age   int
	Email string
}

type StructWithPointers struct {
	Name  *string
	Email *string
}

func TestRenderTemplate_SimpleSubstitution(t *testing.T) {
	svc := &Service{}

	tests := []struct {
		name     string
		template string
		data     any
		expected string
	}{
		{
			name:     "single field substitution",
			template: "Hi {Name}",
			data: &SimpleStruct{
				Name: "John",
			},
			expected: "Hi John",
		},
		{
			name:     "multiple field substitutions",
			template: "Hello {Name}, you are {Age} years old",
			data: &SimpleStruct{
				Name: "Alice",
				Age:  30,
			},
			expected: "Hello Alice, you are 30 years old",
		},
		{
			name:     "all fields substituted",
			template: "{Name} - {Email}",
			data: &SimpleStruct{
				Name:  "Bob",
				Email: "bob@example.com",
			},
			expected: "Bob - bob@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.renderTemplate(tt.template, tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderTemplate_PgtypeTimestampFields(t *testing.T) {
	svc := &Service{}

	ts := time.Date(2024, 12, 4, 15, 30, 0, 0, time.UTC)
	customer := &MockCustomer{
		ID:        1,
		FirstName: pgtype.Text{String: "John", Valid: true},
		CreatedAt: pgtype.Timestamp{Time: ts, Valid: true},
	}

	result := svc.renderTemplate("User {FirstName} created at {CreatedAt}", customer)
	expectedTime := ts.Format(time.RFC3339)
	expected := "User John created at " + expectedTime

	assert.Equal(t, expected, result)
}

func TestRenderTemplate_PointerFields(t *testing.T) {
	svc := &Service{}

	name := "Charlie"
	email := "charlie@example.com"

	tests := []struct {
		name     string
		template string
		data     any
		expected string
	}{
		{
			name:     "pointer fields with valid values",
			template: "{Name} - {Email}",
			data: &StructWithPointers{
				Name:  &name,
				Email: &email,
			},
			expected: "Charlie - charlie@example.com",
		},
		{
			name:     "pointer field is nil",
			template: "Name: {Name}, Email: {Email}",
			data: &StructWithPointers{
				Name:  &name,
				Email: nil,
			},
			expected: "Name: Charlie, Email: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.renderTemplate(tt.template, tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderTemplate_FallbackBehavior(t *testing.T) {
	svc := &Service{}

	tests := []struct {
		name     string
		template string
		data     any
		expected string
	}{
		{
			name:     "field not found - placeholder unchanged",
			template: "Hi {NonExistent}",
			data: &SimpleStruct{
				Name: "John",
			},
			expected: "Hi {NonExistent}",
		},
		{
			name:     "multiple placeholders, some not found",
			template: "{Name} works at {Company} in {Location}",
			data: &SimpleStruct{
				Name: "John",
			},
			expected: "John works at {Company} in {Location}",
		},
		{
			name:     "empty template",
			template: "",
			data: &SimpleStruct{
				Name: "John",
			},
			expected: "",
		},
		{
			name:     "nil data",
			template: "Hello {Name}",
			data:     nil,
			expected: "Hello {Name}",
		},
		{
			name:     "non-struct data",
			template: "Hello {Name}",
			data:     "not a struct",
			expected: "Hello {Name}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.renderTemplate(tt.template, tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderTemplate_DirectStructValue(t *testing.T) {
	svc := &Service{}

	customer := SimpleStruct{
		Name:  "David",
		Email: "david@example.com",
	}

	result := svc.renderTemplate("Hello {Name}", customer)
	assert.Equal(t, "Hello David", result)
}

func TestRenderTemplate_PgtypeInvalidFields(t *testing.T) {
	svc := &Service{}

	customer := &MockCustomer{
		ID:        1,
		FirstName: pgtype.Text{String: "", Valid: false},
		LastName:  pgtype.Text{String: "Doe", Valid: true},
	}

	result := svc.renderTemplate("Name: {FirstName} {LastName}", customer)
	assert.Equal(t, "Name:  Doe", result)
}

func TestRenderTemplate_NoPlaceholders(t *testing.T) {
	svc := &Service{}

	template := "Just plain text with no placeholders"
	result := svc.renderTemplate(template, &SimpleStruct{Name: "John"})

	assert.Equal(t, template, result)
}

func TestRenderTemplate_ComplexScenario(t *testing.T) {
	svc := &Service{}

	customer := &MockCustomer{
		FirstName:        pgtype.Text{String: "Mohammed", Valid: true},
		LastName:         pgtype.Text{String: "Mwijaa", Valid: true},
		Location:         pgtype.Text{String: "Mombasa", Valid: true},
		PreferredProduct: pgtype.Text{String: "White Sneakers", Valid: true},
		Phone:            "+254712832088",
	}

	template := "Hi {FirstName}, thank you for choosing us. We have {PreferredProduct} in stock at our {Location} store. Call us at {Phone} or visit {NonExistent}."
	result := svc.renderTemplate(template, customer)

	expected := "Hi Mohammed, thank you for choosing us. We have White Sneakers in stock at our Mombasa store. Call us at +254712832088 or visit {NonExistent}."
	assert.Equal(t, expected, result)
}
