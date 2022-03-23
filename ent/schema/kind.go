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
		field.Int64("max_reservation_time").Default(120000000),
		field.Int64("max_reservation_units").Default(10),
	}
}

// Edges of the Kind.
func (Kind) Edges() []ent.Edge {
	return nil
}
