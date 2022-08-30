package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Equipment holds the schema definition for the Equipment entity.
type Equipment struct {
	ent.Schema
}

// Fields of the Equipment.
func (Equipment) Fields() []ent.Field {
	return []ent.Field{
		field.String("category").Default("unknown"),
		field.String("name").Default("unknown"),
		field.String("title").Default("unknown"),
		field.Int64("compensationCost").Optional(),
		field.String("tech_issue").Default("нет"),
		field.String("condition").Optional(),
		field.Int64("inventoryNumber").Optional(),
		field.String("supplier").Default("unknown"),
		field.String("receiptDate").Default("unknown"),
		field.Int64("maximumAmount").Optional(),
		field.Int64("maximumDays").Optional(),
		field.String("description").Default("unknown"),
	}
}

// Edges of the Equipment.
func (Equipment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("kind", Kind.Type).Ref("equipments").Unique(),
		edge.From("status", Statuses.Type).Ref("equipments").Unique(),
		edge.From("pet_size", PetSize.Type).Ref("equipments").Unique(),
		edge.From("photo", Photo.Type).Ref("equipments").Unique(),
		edge.From("petKinds", PetKind.Type).Ref("equipments"),
	}
}
