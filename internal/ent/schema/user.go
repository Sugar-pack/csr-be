package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("login").Unique(),
		field.String("email").NotEmpty(),
		field.String("password").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("surname").Optional().Nillable(),
		field.String("patronymic").Optional().Nillable(),
		field.String("passport_series").Optional().Nillable(),
		field.String("passport_number").Optional().Nillable(),
		field.String("passport_authority").Optional().Nillable(),
		field.Time("passport_issue_date").
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamp",
			}),
		field.String("phone").Optional().Nillable(),
		field.Bool("is_readonly").Default(false),
		field.Enum("type").Values("person", "organization").Default("person"),
		field.String("org_name").Optional().Nillable(),
		field.String("website").Optional().Nillable(),
		field.String("vk").Optional().Nillable(),
		field.Bool("is_confirmed").Default(false),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("groups", Group.Type).Ref("users"),
		edge.From("role", Role.Type).Ref("users").Unique(),
		edge.To("tokens", Token.Type),
		edge.From("active_areas", ActiveArea.Type).Ref("users"),
		edge.To("order", Order.Type),
		edge.To("order_status", OrderStatus.Type),
		edge.To("password_reset", PasswordReset.Type),
		edge.To("registration_confirm", RegistrationConfirm.Type),
	}
}
