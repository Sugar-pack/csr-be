package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/kinds"
)

func TestSetKindHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:kindhandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)
	SetKindHandler(client, logger, api)
	assert.NotEmpty(t, api.KindsCreateNewKindHandler)
	assert.NotEmpty(t, api.KindsGetKindByIDHandler)
	assert.NotEmpty(t, api.KindsDeleteKindHandler)
	assert.NotEmpty(t, api.KindsGetAllKindsHandler)
	assert.NotEmpty(t, api.KindsPatchKindHandler)
}

type KindTestSuite struct {
	suite.Suite
	logger     *zap.Logger
	repository *repomock.KindRepository
	handler    *Kind
}

func TestKindSuite(t *testing.T) {
	suite.Run(t, new(KindTestSuite))
}

func (s *KindTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.repository = &repomock.KindRepository{}
	s.handler = NewKind(s.logger)
}

func (s *KindTestSuite) TestKind_CreateKind_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	maxReservationTime := int64(100)
	maxReservationUnit := int64(1)
	kindName := "test"
	newKind := models.CreateNewKind{
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnit,
		Name:                &kindName,
	}
	data := kinds.CreateNewKindParams{
		HTTPRequest: &request,
		NewKind:     &newKind,
	}
	err := errors.New("test")
	s.repository.On("CreateKind", ctx, newKind).Return(nil, err)

	handlerFunc := s.handler.CreateNewKindFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_CreateKind_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	maxReservationTime := int64(100)
	maxReservationUnit := int64(1)
	kindName := "test"
	newKind := models.CreateNewKind{
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnit,
		Name:                &kindName,
	}
	data := kinds.CreateNewKindParams{
		HTTPRequest: &request,
		NewKind:     &newKind,
	}

	kindToReturn := &ent.Kind{
		ID:                  1,
		Name:                kindName,
		MaxReservationTime:  maxReservationTime,
		MaxReservationUnits: maxReservationUnit,
	}
	s.repository.On("CreateKind", ctx, newKind).Return(kindToReturn, nil)

	handlerFunc := s.handler.CreateNewKindFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	returnedKind := models.CreateNewKindResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedKind)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, kindToReturn.ID, int(returnedKind.Data.ID))
	assert.Equal(t, kindToReturn.Name, *returnedKind.Data.Name)
	assert.Equal(t, kindToReturn.MaxReservationTime, returnedKind.Data.MaxReservationTime)
	assert.Equal(t, kindToReturn.MaxReservationUnits, returnedKind.Data.MaxReservationUnits)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_GetAllKinds_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := kinds.GetAllKindsParams{
		HTTPRequest: &request,
	}

	err := errors.New("test")
	s.repository.On("AllKindsTotal", ctx).Return(0, err)

	handlerFunc := s.handler.GetAllKindsFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_GetAllKinds_NotFound() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := kinds.GetAllKindsParams{
		HTTPRequest: &request,
	}

	s.repository.On("AllKindsTotal", ctx).Return(0, nil)

	handlerFunc := s.handler.GetAllKindsFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var returnedKinds models.ListOfKinds
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedKinds)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, int(*returnedKinds.Total))
	assert.Equal(t, 0, len(returnedKinds.Items))

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_GetAllKinds_EmptyParams() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := kind.FieldID
	data := kinds.GetAllKindsParams{
		HTTPRequest: &request,
	}

	kindsToReturn := []*ent.Kind{
		validKind(t, 1),
		validKind(t, 2),
		validKind(t, 3),
	}

	s.repository.On("AllKindsTotal", ctx).Return(len(kindsToReturn), nil)
	s.repository.On("AllKinds", ctx, limit, offset, orderBy, orderColumn).
		Return(kindsToReturn, nil)

	handlerFunc := s.handler.GetAllKindsFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var returnedKinds models.ListOfKinds
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedKinds)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(kindsToReturn), int(*returnedKinds.Total))
	assert.Equal(t, len(kindsToReturn), len(returnedKinds.Items))

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_GetAllKinds_LimitGreaterThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	limit := int64(10)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := kind.FieldID
	data := kinds.GetAllKindsParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	kindsToReturn := []*ent.Kind{
		validKind(t, 1),
		validKind(t, 2),
		validKind(t, 3),
	}

	s.repository.On("AllKindsTotal", ctx).Return(len(kindsToReturn), nil)
	s.repository.On("AllKinds", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(kindsToReturn, nil)

	handlerFunc := s.handler.GetAllKindsFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var returnedKinds models.ListOfKinds
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedKinds)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(kindsToReturn), int(*returnedKinds.Total))
	assert.Equal(t, len(kindsToReturn), len(returnedKinds.Items))
	assert.GreaterOrEqual(t, int(limit), len(returnedKinds.Items))

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_GetAllKinds_LimitLessThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	limit := int64(2)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := kind.FieldID
	data := kinds.GetAllKindsParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	kindsToReturn := []*ent.Kind{
		validKind(t, 1),
		validKind(t, 2),
		validKind(t, 3),
	}

	s.repository.On("AllKindsTotal", ctx).Return(len(kindsToReturn), nil)
	s.repository.On("AllKinds", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(kindsToReturn[:limit], nil)

	handlerFunc := s.handler.GetAllKindsFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var returnedKinds models.ListOfKinds
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedKinds)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(kindsToReturn), int(*returnedKinds.Total))
	assert.Greater(t, len(kindsToReturn), len(returnedKinds.Items))
	assert.GreaterOrEqual(t, int(limit), len(returnedKinds.Items))

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_GetAllKinds_SecondPage() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	limit := int64(2)
	offset := int64(2)
	orderBy := utils.AscOrder
	orderColumn := kind.FieldID
	data := kinds.GetAllKindsParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	kindsToReturn := []*ent.Kind{
		validKind(t, 1),
		validKind(t, 2),
		validKind(t, 3),
	}

	s.repository.On("AllKindsTotal", ctx).Return(len(kindsToReturn), nil)
	s.repository.On("AllKinds", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(kindsToReturn[offset:], nil)

	handlerFunc := s.handler.GetAllKindsFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var returnedKinds models.ListOfKinds
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedKinds)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(kindsToReturn), int(*returnedKinds.Total))
	assert.Greater(t, len(kindsToReturn), len(returnedKinds.Items))
	assert.GreaterOrEqual(t, int(limit), len(returnedKinds.Items))
	assert.Equal(t, len(kindsToReturn)-int(offset), len(returnedKinds.Items))

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_GetAllKinds_SeveralPages() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	limit := int64(3)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := kind.FieldID
	data := kinds.GetAllKindsParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	kindsToReturn := []*ent.Kind{
		validKind(t, 1),
		validKind(t, 2),
		validKind(t, 3),
		validKind(t, 4),
		validKind(t, 5),
	}

	s.repository.On("AllKindsTotal", ctx).Return(len(kindsToReturn), nil)
	s.repository.On("AllKinds", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(kindsToReturn[:limit], nil)

	handlerFunc := s.handler.GetAllKindsFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var firstPage models.ListOfKinds
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &firstPage)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(kindsToReturn), int(*firstPage.Total))
	assert.Greater(t, len(kindsToReturn), len(firstPage.Items))
	assert.Equal(t, int(limit), len(firstPage.Items))

	offset = limit
	data = kinds.GetAllKindsParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}
	s.repository.On("AllKindsTotal", ctx).Return(len(kindsToReturn), nil)
	s.repository.On("AllKinds", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(kindsToReturn[offset:], nil)

	resp = handlerFunc.Handle(data, access)
	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var secondPage models.ListOfKinds
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &secondPage)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(kindsToReturn), int(*secondPage.Total))
	assert.Greater(t, len(kindsToReturn), len(secondPage.Items))
	assert.GreaterOrEqual(t, int(limit), len(secondPage.Items))
	assert.Equal(t, len(kindsToReturn)-int(offset), len(secondPage.Items))

	assert.False(t, kindsDuplicated(t, firstPage.Items, secondPage.Items))
	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_GetKindByID_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := kinds.GetKindByIDParams{
		HTTPRequest: &request,
		KindID:      1,
	}

	err := errors.New("test")
	s.repository.On("KindByID", ctx, int(data.KindID)).Return(nil, err)

	handlerFunc := s.handler.GetKindByIDFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_GetKindByID_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := kinds.GetKindByIDParams{
		HTTPRequest: &request,
		KindID:      1,
	}

	kindToReturn := &ent.Kind{
		ID:                  1,
		Name:                "test",
		MaxReservationTime:  100,
		MaxReservationUnits: 1,
	}
	s.repository.On("KindByID", ctx, int(data.KindID)).Return(kindToReturn, nil)

	handlerFunc := s.handler.GetKindByIDFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	returnedKind := models.GetKindByIDResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedKind)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, kindToReturn.ID, int(returnedKind.Data.ID))
	assert.Equal(t, kindToReturn.Name, *returnedKind.Data.Name)
	assert.Equal(t, kindToReturn.MaxReservationTime, returnedKind.Data.MaxReservationTime)
	assert.Equal(t, kindToReturn.MaxReservationUnits, returnedKind.Data.MaxReservationUnits)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_DeleteKind_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := kinds.DeleteKindParams{
		HTTPRequest: &request,
		KindID:      1,
	}

	err := errors.New("test")
	s.repository.On("DeleteKindByID", ctx, int(data.KindID)).Return(err)

	handlerFunc := s.handler.DeleteKindFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_DeleteKind_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := kinds.DeleteKindParams{
		HTTPRequest: &request,
		KindID:      1,
	}

	s.repository.On("DeleteKindByID", ctx, int(data.KindID)).Return(nil)

	handlerFunc := s.handler.DeleteKindFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_PatchKind_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	patch := models.PatchKind{
		Name: "test",
	}
	data := kinds.PatchKindParams{
		HTTPRequest: &request,
		KindID:      1,
		PatchKind:   &patch,
	}

	err := errors.New("test")
	s.repository.On("UpdateKind", ctx, int(data.KindID), patch).Return(nil, err)

	handlerFunc := s.handler.PatchKindFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_PatchKind_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	patch := models.PatchKind{
		Name: "test",
	}
	data := kinds.PatchKindParams{
		HTTPRequest: &request,
		KindID:      1,
		PatchKind:   &patch,
	}

	updatedKind := &ent.Kind{
		ID:                  1,
		Name:                patch.Name,
		MaxReservationTime:  100,
		MaxReservationUnits: 1,
	}
	s.repository.On("UpdateKind", ctx, int(data.KindID), patch).Return(updatedKind, nil)

	handlerFunc := s.handler.PatchKindFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	returnedKind := models.PatchKindResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedKind)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, updatedKind.ID, int(returnedKind.Data.ID))
	assert.Equal(t, updatedKind.Name, *returnedKind.Data.Name)
	assert.Equal(t, updatedKind.MaxReservationTime, returnedKind.Data.MaxReservationTime)
	assert.Equal(t, updatedKind.MaxReservationUnits, returnedKind.Data.MaxReservationUnits)

	s.repository.AssertExpectations(t)
}

func validKind(t *testing.T, id int) *ent.Kind {
	t.Helper()
	return &ent.Kind{
		ID:                  id,
		Name:                fmt.Sprintf("kind %d", id),
		MaxReservationTime:  10,
		MaxReservationUnits: 5,
	}
}
func kindsDuplicated(t *testing.T, array1, array2 []*models.Kind) bool {
	t.Helper()
	diff := make(map[int64]int, len(array1))
	for _, v := range array1 {
		diff[v.ID] = 1
	}
	for _, v := range array2 {
		if _, ok := diff[v.ID]; ok {
			return true
		}
	}
	return false
}
