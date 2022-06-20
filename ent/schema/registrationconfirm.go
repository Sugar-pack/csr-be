package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// RegistrationConfirm holds the schema definition for the RegistrationConfirm entity.
type RegistrationConfirm struct {
	ent.Schema
}

// Fields of the RegistrationConfirm.
func (RegistrationConfirm) Fields() []ent.Field {
	return []ent.Field{
		field.Time("ttl").
			Default(time.Now()),
		field.String("token").Unique(),
	}
}

// Edges of the RegistrationConfirm.
func (RegistrationConfirm) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", User.Type).Ref("registration_confirm").Unique(),
	}
}
