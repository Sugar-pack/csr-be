package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type PetSize struct {
	ent.Schema
}

func (PetSize) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Default("unknown"),
		field.String("size").Default("unknown"),
		field.Bool("is_universal").Default(false),
	}
}

func (PetSize) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("equipments", Equipment.Type),
	}
}
