package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Category holds the schema definition for the Category entity.
type Category struct {
	ent.Schema
}

// Fields of the Category.
func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Default("unknown"),
		field.Bool("has_subcategory").Default(false),
		field.Int64("max_reservation_time").Default(120000000),
		field.Int64("max_reservation_units").Default(10),
	}
}

// Edges of the Category.
func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("equipments", Equipment.Type),
		edge.To("subcategories", Subcategory.Type),
	}
}
