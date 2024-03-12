package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	eqStatus "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/equipment_status"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

func TestSetEquipmentStatusHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:eqstatushandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)
	SetEquipmentStatusHandler(logger, api)
	require.NotEmpty(t, api.EquipmentStatusCheckEquipmentStatusHandler)
	require.NotEmpty(t, api.EquipmentStatusUpdateEquipmentStatusOnAvailableHandler)
	require.NotEmpty(t, api.EquipmentStatusUpdateEquipmentStatusOnUnavailableHandler)
	require.NotEmpty(t, api.EquipmentStatusUpdateRepairedEquipmentStatusDatesHandler)
}

type EquipmentStatusTestSuite struct {
	suite.Suite
	logger                    *zap.Logger
	equipmentStatusRepository *mocks.EquipmentStatusRepository
	orderStatusRepository     *mocks.OrderStatusRepository
	handler                   *EquipmentStatus
}

func TestStatusSuite(t *testing.T) {
	suite.Run(t, new(EquipmentStatusTestSuite))
}

func (s *EquipmentStatusTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.equipmentStatusRepository = &mocks.EquipmentStatusRepository{}
	s.orderStatusRepository = &mocks.OrderStatusRepository{}
	s.handler = NewEquipmentStatus(s.logger)
}

func (s *EquipmentStatusTestSuite) Test_Put_EquipmentStatusInRepairFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusName := domain.EquipmentStatusNotAvailable
	startDate := time.Date(2023, time.February, 14, 12, 34, 56, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 10)

	data := eqStatus.UpdateEquipmentStatusOnUnavailableParams{
		HTTPRequest:       &request,
		EquipmentstatusID: 11,
		Name: &models.EquipmentStatusInRepairRequest{
			EndDate:    (*strfmt.DateTime)(&endDate),
			StartDate:  (*strfmt.DateTime)(&startDate),
			StatusName: &statusName,
		},
	}

	reduceOneDayFromCurrentStartDate := strfmt.DateTime(
		time.Time(startDate).AddDate(0, 0, -1),
	)

	addOneDayToCurrentEndDate := strfmt.DateTime(
		time.Time(endDate).AddDate(0, 0, 1),
	)

	eqStatusModel := models.EquipmentStatus{
		StartDate:  &reduceOneDayFromCurrentStartDate,
		EndDate:    &addOneDayToCurrentEndDate,
		StatusName: &statusName,
		ID:         &data.EquipmentstatusID,
	}

	timeNowEquipmentStatus = func() time.Time {
		return time.Date(2023, 02, 15, 13, 0, 0, 0, time.UTC)
	}
	timeNow := timeNowEquipmentStatus()

	eqStatusResponseModel := ent.EquipmentStatus{
		ID:        int(data.EquipmentstatusID),
		StartDate: startDate,
		EndDate:   endDate,
		CreatedAt: timeNow,
	}

	s.equipmentStatusRepository.On("Update", ctx, &eqStatusModel).Return(&eqStatusResponseModel, nil)

	orderResult := ent.Order{ID: 22}
	userResult := ent.User{ID: 33}
	s.equipmentStatusRepository.On(
		"GetOrderAndUserByEquipmentStatusID",
		ctx,
		int(*eqStatusModel.ID),
	).Return(&orderResult, &userResult, nil)

	comment := EQUIPMENT_UNDER_REPAIR_COMMENT_FOR_ORDER
	orderID := int64(orderResult.ID)
	orderModel := models.NewOrderStatus{
		Comment:   &comment,
		CreatedAt: (*strfmt.DateTime)(&timeNow),
		OrderID:   &orderID,
		Status:    &domain.OrderStatusRejected,
	}

	s.orderStatusRepository.On("UpdateStatus", ctx, userResult.ID, orderModel).Return(nil)

	eqStatusResponseModel.Edges = ent.EquipmentStatusEdges{Equipments: &ent.Equipment{ID: 1},
		EquipmentStatusName: &ent.EquipmentStatusName{Name: "testStatusName"}}
	s.equipmentStatusRepository.On(
		"GetEquipmentStatusByID",
		ctx,
		int(*eqStatusModel.ID),
	).Return(&eqStatusResponseModel, nil)

	handlerFunc := s.handler.PutEquipmentStatusInRepairFunc(
		s.equipmentStatusRepository, s.orderStatusRepository,
	)

	principal := &models.Principal{}
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)

	require.Equal(t, http.StatusOK, responseRecorder.Code)

	s.equipmentStatusRepository.AssertExpectations(t)
	s.orderStatusRepository.AssertExpectations(t)

	actualEquipmentStatusResponse := &models.EquipmentStatusRepairResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), actualEquipmentStatusResponse)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}

	require.Equal(t, eqStatusModel.ID, actualEquipmentStatusResponse.Data.ID)
	require.Equal(
		t, int64(eqStatusResponseModel.Edges.Equipments.ID),
		*actualEquipmentStatusResponse.Data.EquipmentID,
	)
	require.Equal(
		t, (*strfmt.DateTime)(&endDate),
		actualEquipmentStatusResponse.Data.EndDate,
	)
	require.Equal(
		t, (*strfmt.DateTime)(&startDate),
		actualEquipmentStatusResponse.Data.StartDate,
	)
	require.Equal(
		t, (strfmt.DateTime)(timeNow),
		actualEquipmentStatusResponse.Data.CreatedAt,
	)
	require.Equal(
		t, eqStatusResponseModel.Edges.EquipmentStatusName.Name,
		*actualEquipmentStatusResponse.Data.StatusName,
	)
}

func (s *EquipmentStatusTestSuite) Test_Get_EquipmentStatusCheckDates_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusName := domain.EquipmentStatusNotAvailable
	startDate := time.Date(2023, time.February, 14, 12, 34, 56, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 10)

	data := eqStatus.CheckEquipmentStatusParams{
		HTTPRequest:       &request,
		EquipmentstatusID: 11,
		Name: &models.EquipmentStatusInRepairRequest{
			EndDate:    (*strfmt.DateTime)(&endDate),
			StartDate:  (*strfmt.DateTime)(&startDate),
			StatusName: &statusName,
		},
	}

	reduceOneDayFromCurrentStartDate := strfmt.DateTime(
		time.Time(startDate).AddDate(0, 0, -1),
	)

	addOneDayToCurrentEndDate := strfmt.DateTime(
		time.Time(endDate).AddDate(0, 0, 1),
	)

	eqStatusModel := models.EquipmentStatus{
		StartDate:  &reduceOneDayFromCurrentStartDate,
		EndDate:    &addOneDayToCurrentEndDate,
		StatusName: &statusName,
		ID:         &data.EquipmentstatusID,
	}

	timeNowEquipmentStatus = func() time.Time {
		return time.Date(2022, 01, 15, 13, 0, 0, 0, time.UTC)
	}

	timeNow := timeNowEquipmentStatus()

	eqStatusResponseModel := ent.EquipmentStatus{
		ID:        int(data.EquipmentstatusID),
		StartDate: startDate,
		EndDate:   endDate,
		CreatedAt: timeNow,
	}

	eqStatusResponseModel.Edges = ent.EquipmentStatusEdges{Equipments: &ent.Equipment{ID: 1},
		EquipmentStatusName: &ent.EquipmentStatusName{Name: "testStatusName"}}
	s.equipmentStatusRepository.On(
		"GetEquipmentStatusByID",
		ctx,
		int(*eqStatusModel.ID),
	).Return(&eqStatusResponseModel, nil)

	orderResult := ent.Order{ID: 22}
	userResult := ent.User{ID: 33, Email: "user@email"}

	s.equipmentStatusRepository.On(
		"GetOrderAndUserByEquipmentStatusID",
		ctx,
		int(*eqStatusModel.ID),
	).Return(&orderResult, &userResult, nil)

	handlerFunc := s.handler.GetEquipmentStatusCheckDatesFunc(
		s.equipmentStatusRepository,
	)

	principal := &models.Principal{}
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentStatusRepository.AssertExpectations(t)

	actualEquipmentStatusResponse := &models.EquipmentStatusRepairConfirmationResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), actualEquipmentStatusResponse)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}

	require.Equal(t, data.EquipmentstatusID, *actualEquipmentStatusResponse.Data.EquipmentStatusID)
	require.Equal(
		t, int64(eqStatusResponseModel.Edges.Equipments.ID),
		*actualEquipmentStatusResponse.Data.EquipmentID,
	)
	require.Equal(
		t, data.Name.EndDate,
		actualEquipmentStatusResponse.Data.EndDate,
	)
	require.Equal(
		t, data.Name.StartDate,
		actualEquipmentStatusResponse.Data.StartDate,
	)
	require.Equal(
		t, eqStatusResponseModel.Edges.EquipmentStatusName.Name,
		*actualEquipmentStatusResponse.Data.StatusName,
	)
	require.Equal(
		t, int64(orderResult.ID),
		*actualEquipmentStatusResponse.Data.OrderID,
	)
	require.Equal(
		t, userResult.Email,
		*actualEquipmentStatusResponse.Data.UserEmail,
	)

	// add many dates to end dates, in order for the dates to go out of range
	newEndDate := endDate.AddDate(0, 0, 20)
	newStartDate := startDate.AddDate(0, 0, 20)
	data.Name.EndDate = (*strfmt.DateTime)(&newEndDate)
	data.Name.StartDate = (*strfmt.DateTime)(&newStartDate)
	resp = handlerFunc(data, principal)
	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentStatusRepository.AssertExpectations(t)

	actualEquipmentStatusResponse = &models.EquipmentStatusRepairConfirmationResponse{}
	err = json.Unmarshal(responseRecorder.Body.Bytes(), actualEquipmentStatusResponse)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}

	require.Empty(t, actualEquipmentStatusResponse.Data)

	// subtract many dates from start and end dates
	newStartDate = startDate.AddDate(0, 0, -20)
	newEndDate = endDate.AddDate(0, 0, -20)
	data.Name.EndDate = (*strfmt.DateTime)(&newEndDate)
	data.Name.StartDate = (*strfmt.DateTime)(&newStartDate)

	resp = handlerFunc(data, principal)
	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentStatusRepository.AssertExpectations(t)

	actualEquipmentStatusResponse = &models.EquipmentStatusRepairConfirmationResponse{}
	err = json.Unmarshal(responseRecorder.Body.Bytes(), actualEquipmentStatusResponse)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	require.Empty(t, actualEquipmentStatusResponse.Data)

	// add one date to end date, the start date does not change
	newEndDate = endDate.AddDate(0, 0, 1)
	newStartDate = startDate
	data.Name.EndDate = (*strfmt.DateTime)(&newEndDate)
	data.Name.StartDate = (*strfmt.DateTime)(&newStartDate)

	resp = handlerFunc(data, principal)
	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentStatusRepository.AssertExpectations(t)

	actualEquipmentStatusResponse = &models.EquipmentStatusRepairConfirmationResponse{}
	err = json.Unmarshal(responseRecorder.Body.Bytes(), actualEquipmentStatusResponse)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	require.NotEmpty(t, actualEquipmentStatusResponse.Data)

	// subtract one date from start date, the end date does not change
	newStartDate = startDate.AddDate(0, 0, -1)
	newEndDate = endDate
	data.Name.EndDate = (*strfmt.DateTime)(&newEndDate)
	data.Name.StartDate = (*strfmt.DateTime)(&newStartDate)

	resp = handlerFunc(data, principal)
	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentStatusRepository.AssertExpectations(t)

	actualEquipmentStatusResponse = &models.EquipmentStatusRepairConfirmationResponse{}
	err = json.Unmarshal(responseRecorder.Body.Bytes(), actualEquipmentStatusResponse)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	require.NotEmpty(t, actualEquipmentStatusResponse)

	// add one date to start and end dates
	newStartDate = startDate.AddDate(0, 0, 1)
	newEndDate = endDate.AddDate(0, 0, 1)
	data.Name.EndDate = (*strfmt.DateTime)(&newEndDate)
	data.Name.StartDate = (*strfmt.DateTime)(&newStartDate)

	resp = handlerFunc(data, principal)
	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentStatusRepository.AssertExpectations(t)

	actualEquipmentStatusResponse = &models.EquipmentStatusRepairConfirmationResponse{}
	err = json.Unmarshal(responseRecorder.Body.Bytes(), actualEquipmentStatusResponse)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	require.NotEmpty(t, actualEquipmentStatusResponse)

	// subtract one date to start and end dates
	newStartDate = startDate.AddDate(0, 0, -1)
	newEndDate = endDate.AddDate(0, 0, -1)
	data.Name.EndDate = (*strfmt.DateTime)(&newEndDate)
	data.Name.StartDate = (*strfmt.DateTime)(&newStartDate)

	resp = handlerFunc(data, principal)
	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentStatusRepository.AssertExpectations(t)

	actualEquipmentStatusResponse = &models.EquipmentStatusRepairConfirmationResponse{}
	err = json.Unmarshal(responseRecorder.Body.Bytes(), actualEquipmentStatusResponse)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	require.NotEmpty(t, actualEquipmentStatusResponse)
}

func (s *EquipmentStatusTestSuite) Test_Delete_EquipmentStatusFromRepairFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusName := domain.EquipmentStatusAvailable
	data := eqStatus.UpdateEquipmentStatusOnAvailableParams{
		HTTPRequest:       &request,
		EquipmentstatusID: 1,
		Name: &models.EquipmentStatusRemoveFromRepairRequest{
			StatusName: &statusName,
		},
	}

	timeNowEquipmentStatus = func() time.Time {
		return time.Date(2023, 02, 15, 13, 0, 0, 0, time.UTC)
	}
	timeNow := timeNowEquipmentStatus()

	addOneDayToCurrentEndDate := strfmt.DateTime(
		time.Time(timeNow).AddDate(0, 0, 1),
	)

	eqStatusModel := models.EquipmentStatus{
		EndDate:    &addOneDayToCurrentEndDate,
		StatusName: &statusName,
		ID:         &data.EquipmentstatusID,
	}

	eqStatusResponseModel := ent.EquipmentStatus{
		ID:        int(data.EquipmentstatusID),
		StartDate: timeNow,
		EndDate:   time.Time(addOneDayToCurrentEndDate),
		CreatedAt: timeNow,
	}

	s.equipmentStatusRepository.On("Update", ctx, &eqStatusModel).Return(&eqStatusResponseModel, nil)

	eqStatusResponseModel.Edges = ent.EquipmentStatusEdges{Equipments: &ent.Equipment{ID: 1},
		EquipmentStatusName: &ent.EquipmentStatusName{Name: "testStatusName"}}
	s.equipmentStatusRepository.On(
		"GetEquipmentStatusByID",
		ctx,
		int(*eqStatusModel.ID),
	).Return(&eqStatusResponseModel, nil)

	handlerFunc := s.handler.DeleteEquipmentStatusFromRepairFunc(
		s.equipmentStatusRepository, s.orderStatusRepository,
	)

	principal := &models.Principal{}
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)

	require.Equal(t, http.StatusOK, responseRecorder.Code)

	s.equipmentStatusRepository.AssertExpectations(t)
	s.orderStatusRepository.AssertExpectations(t)

	actualEquipmentStatusResponse := &models.EquipmentStatusRepairResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), actualEquipmentStatusResponse)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}

	require.Equal(t, eqStatusModel.ID, actualEquipmentStatusResponse.Data.ID)
	require.Equal(
		t, int64(eqStatusResponseModel.Edges.Equipments.ID),
		*actualEquipmentStatusResponse.Data.EquipmentID,
	)
	require.Equal(
		t, (*strfmt.DateTime)(&addOneDayToCurrentEndDate),
		actualEquipmentStatusResponse.Data.EndDate,
	)
	require.Equal(
		t, (*strfmt.DateTime)(&timeNow),
		actualEquipmentStatusResponse.Data.StartDate,
	)
	require.Equal(
		t, (strfmt.DateTime)(timeNow),
		actualEquipmentStatusResponse.Data.CreatedAt,
	)
	require.Equal(
		t, eqStatusResponseModel.Edges.EquipmentStatusName.Name,
		*actualEquipmentStatusResponse.Data.StatusName,
	)
}

func (s *EquipmentStatusTestSuite) Test_Patch_EquipmentStatusEditDatesFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	startDate := time.Date(2023, time.February, 14, 12, 34, 56, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 10)

	data := eqStatus.UpdateRepairedEquipmentStatusDatesParams{
		HTTPRequest:       &request,
		EquipmentstatusID: 11,
		Name: &models.EquipmentStatusEditDatesRequest{
			EndDate:   strfmt.DateTime(endDate),
			StartDate: strfmt.DateTime(startDate),
		},
	}

	reduceOneDayFromCurrentStartDate := strfmt.DateTime(
		time.Time(startDate).AddDate(0, 0, -1),
	)

	addOneDayToCurrentEndDate := strfmt.DateTime(
		time.Time(endDate).AddDate(0, 0, 1),
	)

	eqStatusModel := models.EquipmentStatus{
		StartDate: &reduceOneDayFromCurrentStartDate,
		EndDate:   &addOneDayToCurrentEndDate,
		ID:        &data.EquipmentstatusID,
	}

	updatedEqStatus := ent.EquipmentStatus{
		ID:        int(data.EquipmentstatusID),
		StartDate: time.Time(reduceOneDayFromCurrentStartDate),
		EndDate:   time.Time(addOneDayToCurrentEndDate),
	}

	updatedEqStatus.Edges = ent.EquipmentStatusEdges{Equipments: &ent.Equipment{ID: 1},
		EquipmentStatusName: &ent.EquipmentStatusName{Name: "testStatusName"}}
	s.equipmentStatusRepository.On("Update", ctx, &eqStatusModel).Return(&updatedEqStatus, nil)

	s.equipmentStatusRepository.On(
		"GetEquipmentStatusByID",
		ctx,
		int(*eqStatusModel.ID),
	).Return(&updatedEqStatus, nil)

	handlerFunc := s.handler.PatchEquipmentStatusEditDatesFunc(
		s.equipmentStatusRepository,
	)

	principal := &models.Principal{}
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)

	require.Equal(t, http.StatusOK, responseRecorder.Code)

	s.equipmentStatusRepository.AssertExpectations(t)
	s.orderStatusRepository.AssertExpectations(t)

	actualEquipmentStatusResponse := &models.EquipmentStatusRepairResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), actualEquipmentStatusResponse)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}

	require.Equal(t, eqStatusModel.ID, actualEquipmentStatusResponse.Data.ID)
	require.Equal(
		t, int64(updatedEqStatus.Edges.Equipments.ID),
		*actualEquipmentStatusResponse.Data.EquipmentID,
	)
	require.Equal(
		t, (*strfmt.DateTime)(&updatedEqStatus.EndDate),
		actualEquipmentStatusResponse.Data.EndDate,
	)
	require.Equal(
		t, (*strfmt.DateTime)(&updatedEqStatus.StartDate),
		actualEquipmentStatusResponse.Data.StartDate,
	)
	require.Equal(
		t, updatedEqStatus.Edges.EquipmentStatusName.Name,
		*actualEquipmentStatusResponse.Data.StatusName,
	)
}
