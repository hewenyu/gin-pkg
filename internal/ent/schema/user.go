package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			Unique().
			NotEmpty().
			DefaultFunc(func() string {
				return uuid.New().String()
			}).Comment("主键"),
		field.String("email").
			Unique().
			NotEmpty().
			Comment("邮箱"),
		field.String("username").
			Unique().
			NotEmpty().
			Comment("用户名"),
		field.String("password_hash").
			NotEmpty().
			Sensitive().
			Comment("密码"),
		field.String("role").
			Default("user").
			Comment("角色"),
		field.Bool("active").
			Default(true).
			Comment("是否激活"),
		field.String("avatar_url").
			Optional().
			Comment("头像"),
		field.Time("last_login").
			Optional().
			Nillable().
			Comment("最后登录时间"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}

// Mixin of the User schema.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
		index.Fields("username"),
	}
}
