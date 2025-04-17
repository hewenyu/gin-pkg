// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/hewenyu/gin-pkg/internal/ent/predicate"
	"github.com/hewenyu/gin-pkg/internal/ent/user"
)

// UserDelete is the builder for deleting a User entity.
type UserDelete struct {
	config
	hooks    []Hook
	mutation *UserMutation
}

// Where appends a list predicates to the UserDelete builder.
func (ud *UserDelete) Where(ps ...predicate.User) *UserDelete {
	ud.mutation.Where(ps...)
	return ud
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (ud *UserDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, ud.sqlExec, ud.mutation, ud.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (ud *UserDelete) ExecX(ctx context.Context) int {
	n, err := ud.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (ud *UserDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(user.Table, sqlgraph.NewFieldSpec(user.FieldID, field.TypeString))
	if ps := ud.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, ud.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	ud.mutation.done = true
	return affected, err
}

// UserDeleteOne is the builder for deleting a single User entity.
type UserDeleteOne struct {
	ud *UserDelete
}

// Where appends a list predicates to the UserDelete builder.
func (udo *UserDeleteOne) Where(ps ...predicate.User) *UserDeleteOne {
	udo.ud.mutation.Where(ps...)
	return udo
}

// Exec executes the deletion query.
func (udo *UserDeleteOne) Exec(ctx context.Context) error {
	n, err := udo.ud.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{user.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (udo *UserDeleteOne) ExecX(ctx context.Context) {
	if err := udo.Exec(ctx); err != nil {
		panic(err)
	}
}
