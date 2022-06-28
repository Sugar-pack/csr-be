package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Kind holds the schema definition for the Kind entity.

type PetKind struct {
	ent.Schema
}

// Fields of the Kind.
func (PetKind) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique(),
	}
}

// Edges of the Kind.
func (PetKind) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("equipments", Equipment.Type),
	}
}
