package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// OrderStatus holds the schema definition for the OrderStatus entity.
type OrderStatus struct {
	ent.Schema
}

// Fields of the OrderStatus.
func (OrderStatus) Fields() []ent.Field {
	return []ent.Field{
		field.String("comment").Unique(),
		field.Time("current_date"),
	}
}

// Edges of the OrderStatus.
func (OrderStatus) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).Ref("order_status").Unique(),
		edge.From("status_name", StatusName.Type).Ref("order_status").Unique(),
		edge.From("users", User.Type).Ref("order_status").Unique(),
	}
}
