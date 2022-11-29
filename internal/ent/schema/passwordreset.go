package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// PasswordReset holds the schema definition for the PasswordReset entity.
type PasswordReset struct {
	ent.Schema
}

// Fields of the PasswordReset.
func (PasswordReset) Fields() []ent.Field {
	return []ent.Field{
		field.Time("ttl").
			Default(time.Now()),
		field.String("token").Unique(),
	}
}

// Edges of the PasswordReset.
func (PasswordReset) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", User.Type).Ref("password_reset").Unique(),
	}
}
