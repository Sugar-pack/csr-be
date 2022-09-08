package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Subcategory holds the schema definition for the Subcategory entity.
type Subcategory struct {
	ent.Schema
}

// Fields of the Subcategory.
func (Subcategory) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Default("unknown"),
	}
}

// Edges of the Subcategory.
func (Subcategory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("equipments", Equipment.Type),
		edge.From("category", Category.Type).Ref("subcategories").Unique(),
	}
}
