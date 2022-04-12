package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

type Order struct {
	ent.Schema
}

// Fields of the Role.
func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.Text("description"),
		field.Int("quantity").Min(0),
		field.Time("rent_start"),
		field.Time("rent_end"),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the Role.
func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type),
		edge.To("equipments", Equipment.Type),
	}
}
