package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Photo struct {
	ent.Schema
}

func (Photo) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.String("fileName").Default("unknown.jpg"),
	}
}

func (Photo) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("equipments", Equipment.Type),
	}
}
