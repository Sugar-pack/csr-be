package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// StatusName holds the schema definition for the StatusName entity.
type StatusName struct {
	ent.Schema
}

// Fields of the StatusName.
func (StatusName) Fields() []ent.Field {
	return []ent.Field{
		field.String("status").Unique(),
	}
}

// Edges of the StatusName.
func (StatusName) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("order_status", OrderStatus.Type),
	}
}
