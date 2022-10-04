package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// EquipmentStatus holds the schema definition for the EquipmentStatus entity.
type EquipmentStatus struct {
	ent.Schema
}

// Fields of the EquipmentStatus.
func (EquipmentStatus) Fields() []ent.Field {
	return []ent.Field{
		field.String("comment").Default(""),
		field.Time("created_at").Default(time.Now()),
		field.Time("updated_at").Default(time.Now()),
		field.Time("start_date"),
		field.Time("end_date"),
		field.Time("closed_at").Optional(),
	}
}

// Edges of the OrderStatus.
func (EquipmentStatus) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("equipments", Equipment.Type).Ref("equipment_status").Unique(),
		edge.From("equipment_status_name", EquipmentStatusName.Type).Ref("equipment_status").Unique(),
		edge.From("order", Order.Type).Ref("equipment_status").Unique(),
	}
}
