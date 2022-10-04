package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// EquipmentStatusName holds the schema definition for the EquipmentStatusName entity.
type EquipmentStatusName struct {
	ent.Schema
}

// Fields of the EquipmentStatusName.
func (EquipmentStatusName) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique(),
	}
}

// Edges of the EquipmentStatusName.
func (EquipmentStatusName) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("equipments", Equipment.Type),
		edge.To("equipment_status", EquipmentStatus.Type),
	}
}
