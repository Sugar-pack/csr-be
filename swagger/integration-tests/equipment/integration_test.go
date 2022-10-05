package equipment

import (
	"context"
	"os"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/client"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/categories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/equipment"
	eqStatusName "git.epam.com/epm-lstr/epm-lstr-lc/be/client/equipment_status_name"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/pet_kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/pet_size"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/photos"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/integration-tests/common"
)

func TestIntegration_CreateEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	l, p, err := utils.GenerateLoginAndPassword()
	require.NoError(t, err)

	_, err = utils.CreateUser(ctx, client, l, p)
	require.NoError(t, err)

	loginUser, err := utils.LoginUser(ctx, client, l, p)
	require.NoError(t, err)

	auth := utils.AuthInfoFunc(loginUser.GetPayload().AccessToken)

	t.Run("Create Equipment", func(t *testing.T) {
		params := equipment.NewCreateNewEquipmentParamsWithContext(ctx)
		model, err := setParameters(ctx, client, auth)
		require.NoError(t, err)

		params.NewEquipment = model

		res, err := client.Equipment.CreateNewEquipment(params, auth)
		require.NoError(t, err)

		// location returned as nil, in discussion we decided that this parameter will have more that one values
		// for now it is not handled
		// todo: uncomment string below when it's handled properly
		assert.Equal(t, model.Category, res.Payload.Category)
		assert.Equal(t, model.CompensationCost, res.Payload.CompensationCost)
		assert.Equal(t, model.Condition, res.Payload.Condition)
		assert.Equal(t, model.Description, res.Payload.Description)
		assert.Equal(t, model.InventoryNumber, res.Payload.InventoryNumber)
		assert.Equal(t, model.Category, res.Payload.Category)
		//assert.Equal(t, location, *res.Payload.Location)
		assert.Equal(t, model.MaximumAmount, res.Payload.MaximumAmount)
		assert.Equal(t, model.MaximumDays, res.Payload.MaximumDays)
		assert.Equal(t, model.Name, res.Payload.Name)
		assert.Equal(t, model.PetKinds[0], res.Payload.PetKinds[0].ID)
		assert.Equal(t, model.PetSize, res.Payload.PetSize)
		assert.Contains(t, *res.Payload.PhotoID, *model.PhotoID)
		assert.Equal(t, model.ReceiptDate, res.Payload.ReceiptDate)
		assert.Equal(t, model.Status, res.Payload.Status)
		assert.Equal(t, model.Supplier, res.Payload.Supplier)
		assert.Equal(t, model.TechnicalIssues, res.Payload.TechnicalIssues)
		assert.Equal(t, model.Title, res.Payload.Title)
	})

	t.Run("Create Equipment failed: foreign key constraint error", func(t *testing.T) {
		params := equipment.NewCreateNewEquipmentParamsWithContext(ctx)
		model, err := setParameters(ctx, client, auth)
		require.NoError(t, err)

		id := ""
		model.PhotoID = &id
		params.NewEquipment = model

		_, gotErr := client.Equipment.CreateNewEquipment(params, auth)
		require.Error(t, gotErr)

		wantErr := equipment.NewCreateNewEquipmentDefault(500)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{
			Message: "Error while creating equipment",
		}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Create Equipment failed: authorization error 500 Invalid token", func(t *testing.T) {
		params := equipment.NewCreateNewEquipmentParamsWithContext(ctx)
		token := utils.TokenNotExist
		model, err := setParameters(ctx, client, auth)
		require.NoError(t, err)

		params.NewEquipment = model

		_, gotErr := client.Equipment.CreateNewEquipment(params, utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := equipment.NewCreateNewEquipmentDefault(500)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_GetAllEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	l, p, err := utils.GenerateLoginAndPassword()
	require.NoError(t, err)

	_, err = utils.CreateUser(ctx, client, l, p)
	require.NoError(t, err)

	loginUser, err := utils.LoginUser(ctx, client, l, p)
	require.NoError(t, err)

	auth := utils.AuthInfoFunc(loginUser.GetPayload().AccessToken)

	t.Run("Get All Equipment", func(t *testing.T) {
		params := equipment.NewGetAllEquipmentParamsWithContext(ctx)
		require.NoError(t, err)

		res, err := client.Equipment.GetAllEquipment(params, auth)
		require.NoError(t, err)
		assert.NotZero(t, len(res.Payload.Items))
	})

	t.Run("Get All Equipment: authorization error 500 Invalid token", func(t *testing.T) {
		params := equipment.NewGetAllEquipmentParamsWithContext(ctx)
		token := utils.TokenNotExist

		_, gotErr := client.Equipment.GetAllEquipment(params, utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := equipment.NewGetAllEquipmentDefault(500)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_GetEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	l, p, err := utils.GenerateLoginAndPassword()
	require.NoError(t, err)

	_, err = utils.CreateUser(ctx, client, l, p)
	require.NoError(t, err)

	loginUser, err := utils.LoginUser(ctx, client, l, p)
	require.NoError(t, err)

	auth := utils.AuthInfoFunc(loginUser.GetPayload().AccessToken)

	model, err := setParameters(ctx, client, auth)
	require.NoError(t, err)

	created, err := createEquipment(ctx, client, auth, model)
	require.NoError(t, err)

	t.Run("Get Equipment", func(t *testing.T) {
		params := equipment.NewGetEquipmentParamsWithContext(ctx)
		require.NoError(t, err)

		params.EquipmentID = *created.Payload.ID
		res, err := client.Equipment.GetEquipment(params, auth)
		require.NoError(t, err)

		// location returned as nil, in discussion we decided that this parameter will have more that one values
		// for now it is not handled
		// todo: uncomment string below when it's handled properly
		assert.Equal(t, model.Category, res.Payload.Category)
		assert.Equal(t, model.CompensationCost, res.Payload.CompensationCost)
		assert.Equal(t, model.Condition, res.Payload.Condition)
		assert.Equal(t, model.Description, res.Payload.Description)
		assert.Equal(t, model.InventoryNumber, res.Payload.InventoryNumber)
		assert.Equal(t, model.Category, res.Payload.Category)
		assert.Equal(t, model.MaximumAmount, res.Payload.MaximumAmount)
		assert.Equal(t, model.MaximumDays, res.Payload.MaximumDays)
		assert.Equal(t, model.Name, res.Payload.Name)
		//assert.Equal(t, model.Location, res.Payload.Location)
		assert.Equal(t, model.PetKinds[0], res.Payload.PetKinds[0].ID)
		assert.Equal(t, model.PetSize, res.Payload.PetSize)
		assert.Contains(t, *res.Payload.PhotoID, *model.PhotoID)
		assert.Equal(t, model.ReceiptDate, res.Payload.ReceiptDate)
		assert.Equal(t, model.Status, res.Payload.Status)
		assert.Equal(t, model.Supplier, res.Payload.Supplier)
		assert.Equal(t, model.TechnicalIssues, res.Payload.TechnicalIssues)
		assert.Equal(t, model.Title, res.Payload.Title)
	})

	t.Run("Get Equipment failed: passed invalid equipment id", func(t *testing.T) {
		params := equipment.NewGetEquipmentParamsWithContext(ctx)
		require.NoError(t, err)

		params.EquipmentID = int64(-10)
		_, gotErr := client.Equipment.GetEquipment(params, auth)
		require.Error(t, gotErr)

		wantErr := equipment.NewGetEquipmentDefault(500)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{Message: "Error while getting equipment"}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Get Equipment failed: authorization error 500 Invalid token", func(t *testing.T) {
		params := equipment.NewGetEquipmentParamsWithContext(ctx)
		require.NoError(t, err)

		params.EquipmentID = *created.Payload.ID
		token := utils.TokenNotExist
		_, gotErr := client.Equipment.GetEquipment(params, utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := equipment.NewGetEquipmentDefault(500)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_FindEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	l, p, err := utils.GenerateLoginAndPassword()
	require.NoError(t, err)

	_, err = utils.CreateUser(ctx, client, l, p)
	require.NoError(t, err)

	loginUser, err := utils.LoginUser(ctx, client, l, p)
	require.NoError(t, err)

	auth := utils.AuthInfoFunc(loginUser.GetPayload().AccessToken)
	model, err := setParameters(ctx, client, auth)
	require.NoError(t, err)

	_, err = createEquipment(ctx, client, auth, model)
	require.NoError(t, err)

	t.Run("Find Equipment", func(t *testing.T) {
		params := equipment.NewFindEquipmentParamsWithContext(ctx)
		params.FindEquipment = &models.EquipmentFilter{
			Category: *model.Category,
		}
		res, err := client.Equipment.FindEquipment(params, auth)
		require.NoError(t, err)

		assert.NotZero(t, *res.Payload.Total)
		for _, item := range res.Payload.Items {
			assert.Equal(t, model.Category, item.Category)
		}
	})

	t.Run("Find Equipment: limit = 1", func(t *testing.T) {
		params := equipment.NewFindEquipmentParamsWithContext(ctx)
		params.FindEquipment = &models.EquipmentFilter{
			Category: *model.Category,
		}
		limit := int64(1)
		params.WithLimit(&limit)

		res, err := client.Equipment.FindEquipment(params, auth)
		require.NoError(t, err)

		assert.Equal(t, int(limit), len(res.Payload.Items))
	})

	t.Run("Find Equipment failed: authorization error 500 Invalid token", func(t *testing.T) {
		params := equipment.NewFindEquipmentParamsWithContext(ctx)
		params.FindEquipment = &models.EquipmentFilter{
			Category: *model.Category,
		}

		token := utils.TokenNotExist
		_, gotErr := client.Equipment.FindEquipment(params, utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := equipment.NewFindEquipmentDefault(500)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Find Equipment: unknown parameters, zero items found", func(t *testing.T) {
		params := equipment.NewFindEquipmentParamsWithContext(ctx)
		params.FindEquipment = &models.EquipmentFilter{
			TermsOfUse: "unknown category",
		}

		res, gotErr := client.Equipment.FindEquipment(params, auth)
		require.NoError(t, gotErr)

		assert.Zero(t, len(res.Payload.Items))
	})
}

func TestIntegration_EditEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	l, p, err := utils.GenerateLoginAndPassword()
	require.NoError(t, err)

	_, err = utils.CreateUser(ctx, client, l, p)
	require.NoError(t, err)

	loginUser, err := utils.LoginUser(ctx, client, l, p)
	require.NoError(t, err)

	auth := utils.AuthInfoFunc(loginUser.GetPayload().AccessToken)
	model, err := setParameters(ctx, client, auth)
	require.NoError(t, err)

	created, err := createEquipment(ctx, client, auth, model)
	require.NoError(t, err)

	t.Run("Edit Equipment description", func(t *testing.T) {
		desc := "new description"
		model.Description = &desc
		params := equipment.NewEditEquipmentParamsWithContext(ctx).WithEquipmentID(*created.Payload.ID).
			WithEditEquipment(model)

		res, err := client.Equipment.EditEquipment(params, auth)
		require.NoError(t, err)

		assert.Equal(t, desc, *res.Payload.Description)
	})

	t.Run("Edit Equipment description failed: wrong Equipment ID", func(t *testing.T) {
		desc := "new description"
		model.Description = &desc
		params := equipment.NewEditEquipmentParamsWithContext(ctx).WithEquipmentID(int64(-10)).
			WithEditEquipment(model)

		_, gotErr := client.Equipment.EditEquipment(params, auth)
		require.Error(t, gotErr)

		wantErr := equipment.NewEditEquipmentDefault(500)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{Message: "Error while updating equipment"}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Edit Equipment failed: authorization error 500 Invalid token", func(t *testing.T) {
		params := equipment.NewEditEquipmentParamsWithContext(ctx)
		require.NoError(t, err)

		params.EquipmentID = *created.Payload.ID
		token := utils.TokenNotExist
		_, gotErr := client.Equipment.EditEquipment(params, utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := equipment.NewEditEquipmentDefault(500)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_DeleteEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	l, p, err := utils.GenerateLoginAndPassword()
	require.NoError(t, err)

	_, err = utils.CreateUser(ctx, client, l, p)
	require.NoError(t, err)

	loginUser, err := utils.LoginUser(ctx, client, l, p)
	require.NoError(t, err)

	auth := utils.AuthInfoFunc(loginUser.GetPayload().AccessToken)

	t.Run("Delete All Equipment", func(t *testing.T) {
		res, err := client.Equipment.GetAllEquipment(equipment.NewGetAllEquipmentParamsWithContext(ctx), auth)
		require.NoError(t, err)
		assert.NotZero(t, len(res.Payload.Items))

		params := equipment.NewDeleteEquipmentParamsWithContext(ctx)
		for _, item := range res.Payload.Items {
			params.WithEquipmentID(*item.ID)
			_, err = client.Equipment.DeleteEquipment(params, auth)
			require.NoError(t, err)
		}

		res, err = client.Equipment.GetAllEquipment(equipment.NewGetAllEquipmentParamsWithContext(ctx), auth)
		require.NoError(t, err)
		assert.Zero(t, len(res.Payload.Items))
	})

	t.Run("Delete Equipment failed: zero equipments, delete failed", func(t *testing.T) {
		params := equipment.NewDeleteEquipmentParamsWithContext(ctx)
		params.WithEquipmentID(int64(1))
		_, gotErr := client.Equipment.DeleteEquipment(params, auth)
		require.Error(t, gotErr)

		wantErr := equipment.NewDeleteEquipmentDefault(500)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{
			Message: "Error while getting equipment",
		}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Delete Equipment failed: auth failed", func(t *testing.T) {
		params := equipment.NewDeleteEquipmentParamsWithContext(ctx)
		params.WithEquipmentID(int64(1))
		token := utils.TokenNotExist

		_, gotErr := client.Equipment.DeleteEquipment(params, utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := equipment.NewDeleteEquipmentDefault(500)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func setParameters(ctx context.Context, client *client.Be, auth runtime.ClientAuthInfoWriterFunc) (*models.Equipment, error) {
	termsOfUse := "https://..."
	cost := int64(3900)
	condition := "удовлетворительное, местами облупляется краска"
	description := "удобная, подойдет для котов любых размеров"
	inventoryNumber := int64(1)

	category, err := client.Categories.GetCategoryByID(categories.NewGetCategoryByIDParamsWithContext(ctx).WithCategoryID(1), auth)
	if err != nil {
		return nil, err
	}

	location := int64(71)
	amount := int64(1)
	mdays := int64(10)
	catName := "Том"
	rDate := "2018"

	status, err := client.EquipmentStatusName.GetEquipmentStatusName(
		eqStatusName.NewGetEquipmentStatusNameParamsWithContext(ctx).WithStatusID(1), auth)
	if err != nil {
		return nil, err
	}

	f, err := os.Open("../common/cat.jpeg")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	petSize, err := client.PetSize.GetAllPetSize(pet_size.NewGetAllPetSizeParamsWithContext(ctx), auth)
	if err != nil {
		return nil, err
	}

	photo, err := client.Photos.CreateNewPhoto(photos.NewCreateNewPhotoParams().WithContext(ctx).WithFile(f), auth)
	if err != nil {
		return nil, err
	}

	cats, err := client.PetKind.GetPetKind(pet_kind.NewGetPetKindParamsWithContext(ctx).WithPetKindID(1), auth)
	if err != nil {
		return nil, err
	}

	supp := "ИП Григорьев Виталий Васильевич"
	techIss := "нет"
	title := "клетка midwest icrate 1"

	return &models.Equipment{
		TermsOfUse:       termsOfUse,
		CompensationCost: &cost,
		Condition:        condition,
		Description:      &description,
		InventoryNumber:  &inventoryNumber,
		Category:         category.Payload.Data.ID,
		Location:         &location,
		MaximumAmount:    &amount,
		MaximumDays:      &mdays,
		Name:             &catName,
		NameSubstring:    "box",
		PetKinds:         []int64{*cats.Payload.ID},
		PetSize:          &petSize.Payload[0].ID,
		PhotoID:          photo.Payload.Data.ID,
		ReceiptDate:      &rDate,
		Status:           status.Payload.Data.ID,
		Supplier:         &supp,
		TechnicalIssues:  &techIss,
		Title:            &title,
	}, nil
}

func createEquipment(ctx context.Context, client *client.Be, auth runtime.ClientAuthInfoWriterFunc, model *models.Equipment) (*equipment.CreateNewEquipmentCreated, error) {
	paramsCreate := equipment.NewCreateNewEquipmentParamsWithContext(ctx)
	paramsCreate.NewEquipment = model
	created, err := client.Equipment.CreateNewEquipment(paramsCreate, auth)
	if err != nil {
		return nil, err
	}
	return created, nil
}
