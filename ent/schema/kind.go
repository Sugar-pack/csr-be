package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Kind holds the schema definition for the Kind entity.

type Kind struct {
	ent.Schema
}

// Fields of the Kind.
func (Kind) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Default("unknown"),
	}
}

// Edges of the Kind.
func (Kind) Edges() []ent.Edge {
	return nil
}
