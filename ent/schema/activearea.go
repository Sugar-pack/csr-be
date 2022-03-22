package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ActiveArea holds the schema definition for the ActiveArea entity.
type ActiveArea struct {
	ent.Schema
}

// Fields of the ActiveArea.
func (ActiveArea) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
	}
}

// Edges of the ActiveArea.
func (ActiveArea) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type),
	}
}
