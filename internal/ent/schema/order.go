package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Order struct {
	ent.Schema
}

// Fields of the Order.
func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.Text("description"),
		field.Int("quantity").Min(0),
		field.Time("rent_start"),
		field.Time("rent_end"),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the Order.
func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", User.Type).Ref("order").Unique(),
		edge.From("equipments", Equipment.Type).Ref("order"),
		edge.To("order_status", OrderStatus.Type),
		edge.To("equipment_status", EquipmentStatus.Type),
	}
}
