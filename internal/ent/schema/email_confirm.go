package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// EmailConfirm holds the schema definition for the EmailConfirm entity.
type EmailConfirm struct {
	ent.Schema
}

// Fields of the EmailConfirm.
func (EmailConfirm) Fields() []ent.Field {
	return []ent.Field{
		field.Time("ttl").
			Default(time.Now()),
		field.String("token").Unique(),
		field.String("email").Unique(),
	}
}

// Edges of the EmailConfirm.
func (EmailConfirm) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", User.Type).Ref("email_confirm").Unique(),
	}
}
