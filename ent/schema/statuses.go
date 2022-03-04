package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Statuses holds the schema definition for the Statuses entity.
type Statuses struct {
	ent.Schema
}

// Fields of the Statuses.
func (Statuses) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique(),
	}
}

// Edges of the Statuses.
func (Statuses) Edges() []ent.Edge {
	return nil
}
