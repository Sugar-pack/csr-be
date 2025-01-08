package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	eqPeriods "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/equipment"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

func TestSetEquipmentPeriodsHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:eqstatushandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}

	api := operations.NewBeAPI(swaggerSpec)
	SetEquipmentPeriodsHandler(logger, api)
	require.NotEmpty(t, api.EquipmentGetUnavailabilityPeriodsByEquipmentIDHandler)
}

type EquipmentPeriodsTestSuite struct {
	suite.Suite
	logger                    *zap.Logger
	equipmentStatusRepository *mocks.EquipmentStatusRepository
	handler                   *EquipmentPeriods
}

func TestEquipmentPeriodsSuite(t *testing.T) {
	suite.Run(t, new(EquipmentPeriodsTestSuite))
}

func (s *EquipmentPeriodsTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.equipmentStatusRepository = &mocks.EquipmentStatusRepository{}
	s.handler = NewEquipmentPeriods(s.logger)
}

func (s *EquipmentPeriodsTestSuite) Test_Get_EquipmentUnavailableDatesFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := eqPeriods.GetUnavailabilityPeriodsByEquipmentIDParams{
		HTTPRequest: &request,
		EquipmentID: int64(1),
	}

	startDate := time.Date(2023, time.February, 14, 12, 34, 56, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 10)
	eqStatus := ent.EquipmentStatus{
		StartDate: startDate,
		EndDate:   endDate,
	}

	var eqStatusResponse []*ent.EquipmentStatus
	eqStatusResponse = append(eqStatusResponse, &eqStatus)

	s.equipmentStatusRepository.On(
		"GetUnavailableEquipmentStatusByEquipmentID",
		ctx, int(data.EquipmentID),
	).Return(eqStatusResponse, nil)

	handlerFunc := s.handler.GetEquipmentUnavailableDatesFunc(
		s.equipmentStatusRepository,
	)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)

	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentStatusRepository.AssertExpectations(t)

	actualResponse := &models.EquipmentUnavailabilityPeriodsResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), actualResponse)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}

	require.Equal(
		t, (*strfmt.DateTime)(&startDate),
		actualResponse.Items[0].StartDate,
	)

	require.Equal(
		t, (*strfmt.DateTime)(&endDate),
		actualResponse.Items[0].EndDate,
	)
}
