package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// OrderStatusName holds the schema definition for the OrderStatusName entity.
type OrderStatusName struct {
	ent.Schema
}

// Fields of the OrderStatusName.
func (OrderStatusName) Fields() []ent.Field {
	return []ent.Field{
		field.String("status").Unique(),
	}
}

// Edges of the OrderStatusName.
func (OrderStatusName) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("orders", Order.Type),
		edge.To("order_status", OrderStatus.Type),
	}
}
