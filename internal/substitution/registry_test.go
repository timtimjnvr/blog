package substitution

import (
	"testing"

	"github.com/timtimjnvr/blog/internal/context"
)

// mockSubstituter is a test double for Substituter
type mockSubstituter struct {
	placeholder string
	resolveFunc func(ctx *context.PageContext) string
}

func (m *mockSubstituter) Placeholder() string {
	return m.placeholder
}

func (m *mockSubstituter) Resolve(ctx *context.PageContext) string {
	return m.resolveFunc(ctx)
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry[*context.PageContext]()
	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if len(registry.substitutions) != 0 {
		t.Errorf("new registry should have no substitutions, got %d", len(registry.substitutions))
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry[*context.PageContext]()

	sub := &mockSubstituter{
		placeholder: "{{test}}",
		resolveFunc: func(ctx *context.PageContext) string { return "value" },
	}

	registry.Register(sub)

	if len(registry.substitutions) != 1 {
		t.Errorf("expected 1 substitution, got %d", len(registry.substitutions))
	}
}

func TestRegistry_RegisterMultiple(t *testing.T) {
	registry := NewRegistry[*context.PageContext]()

	sub1 := &mockSubstituter{placeholder: "{{one}}", resolveFunc: func(ctx *context.PageContext) string { return "1" }}
	sub2 := &mockSubstituter{placeholder: "{{two}}", resolveFunc: func(ctx *context.PageContext) string { return "2" }}

	registry.Register(sub1)
	registry.Register(sub2)

	if len(registry.substitutions) != 2 {
		t.Errorf("expected 2 substitutions, got %d", len(registry.substitutions))
	}
}

func TestRegistry_Apply(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		substituters []*mockSubstituter
		ctx          *context.PageContext
		expected     string
	}{
		{
			name:         "no substitutions",
			template:     "Hello World",
			substituters: nil,
			ctx:          &context.PageContext{},
			expected:     "Hello World",
		},
		{
			name:     "single substitution",
			template: "Hello {{name}}!",
			substituters: []*mockSubstituter{
				{placeholder: "{{name}}", resolveFunc: func(ctx *context.PageContext) string { return "Alice" }},
			},
			ctx:      &context.PageContext{},
			expected: "Hello Alice!",
		},
		{
			name:     "multiple substitutions",
			template: "{{greeting}} {{name}}!",
			substituters: []*mockSubstituter{
				{placeholder: "{{greeting}}", resolveFunc: func(ctx *context.PageContext) string { return "Hello" }},
				{placeholder: "{{name}}", resolveFunc: func(ctx *context.PageContext) string { return "Bob" }},
			},
			ctx:      &context.PageContext{},
			expected: "Hello Bob!",
		},
		{
			name:     "substitution uses context",
			template: "Title: {{title}}",
			substituters: []*mockSubstituter{
				{placeholder: "{{title}}", resolveFunc: func(ctx *context.PageContext) string {
					return string(ctx.Source[:5])
				}},
			},
			ctx:      &context.PageContext{Source: []byte("Hello World")},
			expected: "Title: Hello",
		},
		{
			name:     "multiple occurrences of same placeholder",
			template: "{{x}} and {{x}}",
			substituters: []*mockSubstituter{
				{placeholder: "{{x}}", resolveFunc: func(ctx *context.PageContext) string { return "Y" }},
			},
			ctx:      &context.PageContext{},
			expected: "Y and Y",
		},
		{
			name:     "placeholder not in template",
			template: "No placeholders here",
			substituters: []*mockSubstituter{
				{placeholder: "{{missing}}", resolveFunc: func(ctx *context.PageContext) string { return "value" }},
			},
			ctx:      &context.PageContext{},
			expected: "No placeholders here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry[*context.PageContext]()
			for _, sub := range tt.substituters {
				registry.Register(sub)
			}

			result := registry.Apply(tt.template, tt.ctx)
			if result != tt.expected {
				t.Errorf("Apply() = %q, want %q", result, tt.expected)
			}
		})
	}
}
