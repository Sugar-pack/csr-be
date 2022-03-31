// Code generated by entc, DO NOT EDIT.

package kind

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldID), id))
	})
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldID), id))
	})
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldID), id))
	})
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(ids) == 0 {
			s.Where(sql.False())
			return
		}
		v := make([]interface{}, len(ids))
		for i := range v {
			v[i] = ids[i]
		}
		s.Where(sql.In(s.C(FieldID), v...))
	})
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(ids) == 0 {
			s.Where(sql.False())
			return
		}
		v := make([]interface{}, len(ids))
		for i := range v {
			v[i] = ids[i]
		}
		s.Where(sql.NotIn(s.C(FieldID), v...))
	})
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldID), id))
	})
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldID), id))
	})
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldID), id))
	})
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldID), id))
	})
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldName), v))
	})
}

// MaxReservationTime applies equality check predicate on the "max_reservation_time" field. It's identical to MaxReservationTimeEQ.
func MaxReservationTime(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldMaxReservationTime), v))
	})
}

// MaxReservationUnits applies equality check predicate on the "max_reservation_units" field. It's identical to MaxReservationUnitsEQ.
func MaxReservationUnits(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldMaxReservationUnits), v))
	})
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldName), v))
	})
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldName), v))
	})
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.Kind {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Kind(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldName), v...))
	})
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.Kind {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Kind(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldName), v...))
	})
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldName), v))
	})
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldName), v))
	})
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldName), v))
	})
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldName), v))
	})
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldName), v))
	})
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldName), v))
	})
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldName), v))
	})
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldName), v))
	})
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldName), v))
	})
}

// MaxReservationTimeEQ applies the EQ predicate on the "max_reservation_time" field.
func MaxReservationTimeEQ(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldMaxReservationTime), v))
	})
}

// MaxReservationTimeNEQ applies the NEQ predicate on the "max_reservation_time" field.
func MaxReservationTimeNEQ(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldMaxReservationTime), v))
	})
}

// MaxReservationTimeIn applies the In predicate on the "max_reservation_time" field.
func MaxReservationTimeIn(vs ...int64) predicate.Kind {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Kind(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldMaxReservationTime), v...))
	})
}

// MaxReservationTimeNotIn applies the NotIn predicate on the "max_reservation_time" field.
func MaxReservationTimeNotIn(vs ...int64) predicate.Kind {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Kind(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldMaxReservationTime), v...))
	})
}

// MaxReservationTimeGT applies the GT predicate on the "max_reservation_time" field.
func MaxReservationTimeGT(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldMaxReservationTime), v))
	})
}

// MaxReservationTimeGTE applies the GTE predicate on the "max_reservation_time" field.
func MaxReservationTimeGTE(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldMaxReservationTime), v))
	})
}

// MaxReservationTimeLT applies the LT predicate on the "max_reservation_time" field.
func MaxReservationTimeLT(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldMaxReservationTime), v))
	})
}

// MaxReservationTimeLTE applies the LTE predicate on the "max_reservation_time" field.
func MaxReservationTimeLTE(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldMaxReservationTime), v))
	})
}

// MaxReservationUnitsEQ applies the EQ predicate on the "max_reservation_units" field.
func MaxReservationUnitsEQ(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldMaxReservationUnits), v))
	})
}

// MaxReservationUnitsNEQ applies the NEQ predicate on the "max_reservation_units" field.
func MaxReservationUnitsNEQ(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldMaxReservationUnits), v))
	})
}

// MaxReservationUnitsIn applies the In predicate on the "max_reservation_units" field.
func MaxReservationUnitsIn(vs ...int64) predicate.Kind {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Kind(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldMaxReservationUnits), v...))
	})
}

// MaxReservationUnitsNotIn applies the NotIn predicate on the "max_reservation_units" field.
func MaxReservationUnitsNotIn(vs ...int64) predicate.Kind {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Kind(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldMaxReservationUnits), v...))
	})
}

// MaxReservationUnitsGT applies the GT predicate on the "max_reservation_units" field.
func MaxReservationUnitsGT(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldMaxReservationUnits), v))
	})
}

// MaxReservationUnitsGTE applies the GTE predicate on the "max_reservation_units" field.
func MaxReservationUnitsGTE(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldMaxReservationUnits), v))
	})
}

// MaxReservationUnitsLT applies the LT predicate on the "max_reservation_units" field.
func MaxReservationUnitsLT(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldMaxReservationUnits), v))
	})
}

// MaxReservationUnitsLTE applies the LTE predicate on the "max_reservation_units" field.
func MaxReservationUnitsLTE(v int64) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldMaxReservationUnits), v))
	})
}

// HasEquipments applies the HasEdge predicate on the "equipments" edge.
func HasEquipments() predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(EquipmentsTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, EquipmentsTable, EquipmentsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasEquipmentsWith applies the HasEdge predicate on the "equipments" edge with a given conditions (other predicates).
func HasEquipmentsWith(preds ...predicate.Equipment) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(EquipmentsInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, EquipmentsTable, EquipmentsColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Kind) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for _, p := range predicates {
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Kind) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for i, p := range predicates {
			if i > 0 {
				s1.Or()
			}
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Kind) predicate.Kind {
	return predicate.Kind(func(s *sql.Selector) {
		p(s.Not())
	})
}
