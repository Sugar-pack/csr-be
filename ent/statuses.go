// Code generated by entc, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/statuses"
)

// Statuses is the model entity for the Statuses schema.
type Statuses struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Name holds the value of the "name" field.
	Name string `json:"name,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the StatusesQuery when eager-loading is set.
	Edges StatusesEdges `json:"edges"`
}

// StatusesEdges holds the relations/edges for other nodes in the graph.
type StatusesEdges struct {
	// Equipments holds the value of the equipments edge.
	Equipments []*Equipment `json:"equipments,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// EquipmentsOrErr returns the Equipments value or an error if the edge
// was not loaded in eager-loading.
func (e StatusesEdges) EquipmentsOrErr() ([]*Equipment, error) {
	if e.loadedTypes[0] {
		return e.Equipments, nil
	}
	return nil, &NotLoadedError{edge: "equipments"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Statuses) scanValues(columns []string) ([]interface{}, error) {
	values := make([]interface{}, len(columns))
	for i := range columns {
		switch columns[i] {
		case statuses.FieldID:
			values[i] = new(sql.NullInt64)
		case statuses.FieldName:
			values[i] = new(sql.NullString)
		default:
			return nil, fmt.Errorf("unexpected column %q for type Statuses", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Statuses fields.
func (s *Statuses) assignValues(columns []string, values []interface{}) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case statuses.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			s.ID = int(value.Int64)
		case statuses.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				s.Name = value.String
			}
		}
	}
	return nil
}

// QueryEquipments queries the "equipments" edge of the Statuses entity.
func (s *Statuses) QueryEquipments() *EquipmentQuery {
	return (&StatusesClient{config: s.config}).QueryEquipments(s)
}

// Update returns a builder for updating this Statuses.
// Note that you need to call Statuses.Unwrap() before calling this method if this Statuses
// was returned from a transaction, and the transaction was committed or rolled back.
func (s *Statuses) Update() *StatusesUpdateOne {
	return (&StatusesClient{config: s.config}).UpdateOne(s)
}

// Unwrap unwraps the Statuses entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (s *Statuses) Unwrap() *Statuses {
	tx, ok := s.config.driver.(*txDriver)
	if !ok {
		panic("ent: Statuses is not a transactional entity")
	}
	s.config.driver = tx.drv
	return s
}

// String implements the fmt.Stringer.
func (s *Statuses) String() string {
	var builder strings.Builder
	builder.WriteString("Statuses(")
	builder.WriteString(fmt.Sprintf("id=%v", s.ID))
	builder.WriteString(", name=")
	builder.WriteString(s.Name)
	builder.WriteByte(')')
	return builder.String()
}

// StatusesSlice is a parsable slice of Statuses.
type StatusesSlice []*Statuses

func (s StatusesSlice) config(cfg config) {
	for _i := range s {
		s[_i].config = cfg
	}
}
