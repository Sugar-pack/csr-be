package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type equipmentStatusTestSuite struct {
	suite.Suite
	ctx           context.Context
	client        *ent.Client
	statusNameMap map[int]string
	repository    domain.EquipmentStatusRepository
	equipment     *ent.Equipment
	order         *ent.Order
	eqStatus      *ent.EquipmentStatus
}

func TestEquipmentStatusSuite(t *testing.T) {
	s := new(equipmentStatusTestSuite)
	suite.Run(t, s)
}

func (s *equipmentStatusTestSuite) SetupTest() {
	t := s.T()
	s.ctx = context.Background()
	client := enttest.Open(t, "sqlite3", "file:equipmentstatus?mode=memory&cache=shared&_fk=1")
	s.client = client

	s.statusNameMap = map[int]string{
		1: domain.EquipmentStatusAvailable,
		2: domain.EquipmentStatusBooked,
		3: domain.EquipmentStatusInUse,
		4: domain.EquipmentStatusNotAvailable,
	}

	_, err := s.client.EquipmentStatusName.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.client.User.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.client.Order.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.client.Equipment.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.client.EquipmentStatus.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	var notAvailableEquipment *ent.EquipmentStatusName
	for _, statusName := range s.statusNameMap { // create statuses
		eqStatus, errCreation := s.client.EquipmentStatusName.Create().SetName(statusName).Save(s.ctx)
		if errCreation != nil {
			t.Fatal(errCreation)
		}
		if statusName == domain.EquipmentStatusNotAvailable {
			notAvailableEquipment = eqStatus
		}
	}

	order, err := s.client.Order.Create().
		SetDescription("test order").SetQuantity(1).
		SetRentStart(time.Now().AddDate(0, 0, 10)).
		SetRentEnd(time.Now().AddDate(0, 0, 20)).
		Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	s.order = order

	eqCategory, err := s.client.Category.Create().SetName("test category").Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	eq, err := s.client.Equipment.Create().SetName("test equipment").SetCategory(eqCategory).Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	s.equipment = eq

	eqStatus, err := s.client.EquipmentStatus.Create().
		SetEquipments(eq).
		SetOrder(order).
		SetEquipmentStatusName(notAvailableEquipment).
		SetStartDate(order.RentStart).
		SetEndDate(order.RentEnd.AddDate(0, 0, 1)).
		Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	s.eqStatus = eqStatus

	s.repository = NewEquipmentStatusRepository()
}

func (s *equipmentStatusTestSuite) TearDownSuite() {
	s.client.Close()
}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_Create_OrderNotExists() {
	t := s.T()
	comment := "test comment"
	endDate := strfmt.DateTime(s.order.RentEnd.AddDate(0, 0, 1))
	startDate := strfmt.DateTime(s.order.RentStart)
	orderID := int64(s.order.ID + 10)
	equipmentID := int64(s.equipment.ID)
	status := s.statusNameMap[1]
	data := &models.NewEquipmentStatus{
		Comment:     comment,
		EndDate:     &endDate,
		EquipmentID: &equipmentID,
		OrderID:     orderID,
		StartDate:   &startDate,
		StatusName:  &status,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	eqStatus, err := s.repository.Create(ctx, data)
	require.Error(t, err)
	require.Nil(t, eqStatus)
	require.NoError(t, tx.Rollback())
}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_GetUnavailableEquipmentStatusByEquipmentID_OK() {
	t := s.T()
	comment := "test comment"
	endDate := strfmt.DateTime(s.order.RentEnd.AddDate(0, 0, 1))
	startDate := strfmt.DateTime(s.order.RentStart)
	orderID := int64(s.order.ID)
	equipmentID := int64(s.equipment.ID)
	status := s.statusNameMap[1]
	data := &models.NewEquipmentStatus{
		Comment:     comment,
		EndDate:     &endDate,
		EquipmentID: &equipmentID,
		OrderID:     orderID,
		StartDate:   &startDate,
		StatusName:  &status,
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)

	eqStatus, err := s.repository.Create(ctx, data)
	require.NoError(t, err)
	require.NotNil(t, eqStatus)

	result, err := s.repository.GetUnavailableEquipmentStatusByEquipmentID(ctx, int(equipmentID))
	require.NoError(t, err)

	assert.Equal(t, 1, len(result))

	status = s.statusNameMap[1]
	data.StatusName = &status
	eqStatus, err = s.repository.Create(ctx, data)
	require.NoError(t, err)
	require.NotNil(t, eqStatus)

	result, err = s.repository.GetUnavailableEquipmentStatusByEquipmentID(ctx, int(equipmentID))
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, 1, len(result))

	status = s.statusNameMap[2]
	data.StatusName = &status
	eqStatus, err = s.repository.Create(ctx, data)
	require.NoError(t, err)
	require.NotNil(t, eqStatus)

	result, err = s.repository.GetUnavailableEquipmentStatusByEquipmentID(ctx, int(equipmentID))
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, 2, len(result))

	status = s.statusNameMap[3]
	data.StatusName = &status
	eqStatus, err = s.repository.Create(ctx, data)
	require.NoError(t, err)
	require.NotNil(t, eqStatus)

	result, err = s.repository.GetUnavailableEquipmentStatusByEquipmentID(ctx, int(equipmentID))
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, 3, len(result))

	status = s.statusNameMap[4]
	data.StatusName = &status
	eqStatus, err = s.repository.Create(ctx, data)
	require.NoError(t, err)
	require.NotNil(t, eqStatus)

	result, err = s.repository.GetUnavailableEquipmentStatusByEquipmentID(ctx, int(equipmentID))
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, 4, len(result))

	require.NoError(t, tx.Rollback())
}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_Create_StatusNameNotExists() {
	t := s.T()
	comment := "test comment"
	endDate := strfmt.DateTime(s.order.RentEnd.AddDate(0, 0, 1))
	startDate := strfmt.DateTime(s.order.RentStart)
	orderID := int64(s.order.ID)
	equipmentID := int64(s.equipment.ID)
	status := "test status"
	data := &models.NewEquipmentStatus{
		Comment:     comment,
		EndDate:     &endDate,
		EquipmentID: &equipmentID,
		OrderID:     orderID,
		StartDate:   &startDate,
		StatusName:  &status,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	eqStatus, err := s.repository.Create(ctx, data)
	require.Error(t, err)
	require.Nil(t, eqStatus)
	require.NoError(t, tx.Rollback())
}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_Create_EquipmentNotExists() {
	t := s.T()
	comment := "test comment"
	endDate := strfmt.DateTime(s.order.RentEnd.AddDate(0, 0, 1))
	startDate := strfmt.DateTime(s.order.RentStart)
	orderID := int64(s.order.ID)
	equipmentID := int64(s.equipment.ID + 10)
	status := s.statusNameMap[1]
	data := &models.NewEquipmentStatus{
		Comment:     comment,
		EndDate:     &endDate,
		EquipmentID: &equipmentID,
		OrderID:     orderID,
		StartDate:   &startDate,
		StatusName:  &status,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	eqStatus, err := s.repository.Create(ctx, data)
	require.Error(t, err)
	require.Nil(t, eqStatus)
	require.NoError(t, tx.Rollback())
}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_Create_LongerThanMaxReservationTime() {
	t := s.T()
	comment := "test comment"
	startDate := strfmt.DateTime(s.order.RentStart.AddDate(10, 0, 0))
	endDate := strfmt.DateTime(s.order.RentEnd)
	orderID := int64(s.order.ID)
	equipmentID := int64(s.equipment.ID + 10)
	status := s.statusNameMap[1]
	data := &models.NewEquipmentStatus{
		Comment:     comment,
		EndDate:     &endDate,
		EquipmentID: &equipmentID,
		OrderID:     orderID,
		StartDate:   &startDate,
		StatusName:  &status,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	eqStatus, err := s.repository.Create(ctx, data)
	require.Error(t, err)
	require.Nil(t, eqStatus)
	require.NoError(t, tx.Rollback())
}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_Create_OK() {
	t := s.T()
	comment := "test comment"
	endDate := strfmt.DateTime(s.order.RentEnd.AddDate(0, 0, 1))
	startDate := strfmt.DateTime(s.order.RentStart)
	orderID := int64(s.order.ID)
	equipmentID := int64(s.equipment.ID)
	status := s.statusNameMap[1]
	data := &models.NewEquipmentStatus{
		Comment:     comment,
		EndDate:     &endDate,
		EquipmentID: &equipmentID,
		OrderID:     orderID,
		StartDate:   &startDate,
		StatusName:  &status,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	eqStatus, err := s.repository.Create(ctx, data)
	require.NoError(t, err)
	require.NotNil(t, eqStatus)
	require.NoError(t, tx.Rollback())
}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_Update_StatusNameNotExists() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	eqStatusID := int64(s.eqStatus.ID)
	statusName := "test status"
	eqStatus, err := s.repository.Update(ctx, &models.EquipmentStatus{
		ID:         &eqStatusID,
		StatusName: &statusName,
	})
	require.Error(t, err)
	require.Nil(t, eqStatus)
	require.NoError(t, tx.Rollback())

}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_Update_UpdStatusOK() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	eqStatusID := int64(s.eqStatus.ID)
	statusName := s.statusNameMap[2]
	eqStatus, err := s.repository.Update(ctx, &models.EquipmentStatus{
		ID:         &eqStatusID,
		StatusName: &statusName,
	})
	require.NoError(t, err)
	require.NotNil(t, eqStatus)
	require.NoError(t, tx.Rollback())

}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_IsAvailableByPeriod_IntersectsWithAnotherStatusEnding() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)

	start := s.order.RentEnd.AddDate(0, 0, -2)
	end := start.AddDate(0, 0, 10)

	isAvailable, err := s.repository.HasStatusByPeriod(ctx, domain.EquipmentStatusAvailable, s.equipment.ID, start, end)
	require.NoError(t, err)
	require.False(t, isAvailable)
	require.NoError(t, tx.Rollback())
}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_IsAvailableByPeriod_IntersectsWithAnotherStatusBeginning() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	start := time.Now()
	end := s.order.RentStart.AddDate(0, 0, 2)

	isAvailable, err := s.repository.HasStatusByPeriod(ctx, domain.EquipmentStatusAvailable, s.equipment.ID, start, end)
	require.NoError(t, err)
	require.False(t, isAvailable)
	require.NoError(t, tx.Rollback())
}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_IsAvailableByPeriod_IntersectsExistingStatus() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	start := s.order.RentStart.AddDate(0, 0, -1)
	end := s.order.RentEnd.AddDate(0, 0, 1)

	isAvailable, err := s.repository.HasStatusByPeriod(ctx, domain.EquipmentStatusAvailable, s.equipment.ID, start, end)
	require.NoError(t, err)
	require.False(t, isAvailable)
	require.NoError(t, tx.Rollback())
}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_IsAvailableByPeriod_OK() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	start := time.Now()
	end := start.AddDate(0, 0, 5)

	isAvailable, err := s.repository.HasStatusByPeriod(ctx, domain.EquipmentStatusAvailable, s.equipment.ID, start, end)
	require.NoError(t, err)
	require.True(t, isAvailable)
	require.NoError(t, tx.Rollback())
}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_GetEquipmentsStatusesByOrder_OrderNotExists() {
	t := s.T()
	orderID := s.order.ID + 10
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	eqStatus, err := s.repository.GetEquipmentsStatusesByOrder(ctx, orderID)
	require.NoError(t, err)
	require.Empty(t, eqStatus)
	require.NoError(t, tx.Rollback())
}

func (s *equipmentStatusTestSuite) TestEquipmentStatusRepository_GetEquipmentsStatusesByOrder_OK() {
	t := s.T()
	orderID := s.order.ID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	eqStatus, err := s.repository.GetEquipmentsStatusesByOrder(ctx, orderID)
	require.NoError(t, err)
	require.NotEmpty(t, eqStatus)
	require.Greater(t, len(eqStatus), 0)
	require.NoError(t, tx.Rollback())
}
