package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// PetKind holds the schema definition for the PetKind entity.

type PetKind struct {
	ent.Schema
}

// Fields of the PetKind.
func (PetKind) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique(),
	}
}

// Edges of the PetKind.
func (PetKind) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("equipments", Equipment.Type),
	}
}
