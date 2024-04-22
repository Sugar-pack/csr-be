package domain

import (
	"context"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
)

type ActiveAreaRepository interface {
	AllActiveAreas(ctx context.Context, limit, offset int, orderBy, orderColumn string) ([]*ent.ActiveArea, error)
	TotalActiveAreas(ctx context.Context) (int, error)
}

type Filter struct {
	Limit, Offset        int
	OrderBy, OrderColumn string
}

type CategoryFilter struct {
	Filter
	HasEquipments bool
}

type OrderFilter struct {
	Filter
	Status      *string
	EquipmentID *int
}

type CategoryRepository interface {
	CreateCategory(ctx context.Context, newCategory models.CreateNewCategory) (*ent.Category, error)
	AllCategories(ctx context.Context, filter CategoryFilter) ([]*ent.Category, error)
	AllCategoriesTotal(ctx context.Context) (int, error)
	CategoryByID(ctx context.Context, id int) (*ent.Category, error)
	DeleteCategoryByID(ctx context.Context, id int) error
	UpdateCategory(ctx context.Context, id int, update models.UpdateCategoryRequest) (*ent.Category, error)
}

type EquipmentRepository interface {
	EquipmentsByFilter(ctx context.Context, filter models.EquipmentFilter, limit, offset int,
		orderBy, orderColumn string) ([]*ent.Equipment, error)
	CreateEquipment(ctx context.Context, eq models.Equipment, status *ent.EquipmentStatusName) (*ent.Equipment, error)
	EquipmentByID(ctx context.Context, id int) (*ent.Equipment, error)
	DeleteEquipmentByID(ctx context.Context, id int) error
	DeleteEquipmentPhoto(ctx context.Context, id string) error
	AllEquipments(ctx context.Context, limit, offset int, orderBy, orderColumn string) ([]*ent.Equipment, error)
	UpdateEquipmentByID(ctx context.Context, id int, eq *models.Equipment) (*ent.Equipment, error)
	AllEquipmentsTotal(ctx context.Context) (int, error)
	EquipmentsByFilterTotal(ctx context.Context, filter models.EquipmentFilter) (int, error)
	ArchiveEquipment(ctx context.Context, id int) error
	BlockEquipment(ctx context.Context, id int, startDate, endDate time.Time, userID int) error
	UnblockEquipment(ctx context.Context, id int) error
}

type EquipmentStatusRepository interface {
	Create(ctx context.Context, data *models.NewEquipmentStatus) (*ent.EquipmentStatus, error)
	GetEquipmentsStatusesByOrder(ctx context.Context, orderID int) ([]*ent.EquipmentStatus, error)
	HasStatusByPeriod(ctx context.Context, status string, eqID int, startDate, endDate time.Time) (bool, error)
	Update(ctx context.Context, data *models.EquipmentStatus) (*ent.EquipmentStatus, error)
	GetOrderAndUserByEquipmentStatusID(ctx context.Context, id int) (*ent.Order, *ent.User, error)
	GetEquipmentStatusByID(ctx context.Context, equipmentStatusID int) (*ent.EquipmentStatus, error)
	GetUnavailableEquipmentStatusByEquipmentID(ctx context.Context, equipmentID int) ([]*ent.EquipmentStatus, error)
	GetLastEquipmentStatusByEquipmentID(ctx context.Context, equipmentID int) (*ent.EquipmentStatus, error)
}

type EquipmentStatusNameRepository interface {
	Create(ctx context.Context, name string) (*ent.EquipmentStatusName, error)
	GetAll(ctx context.Context) ([]*ent.EquipmentStatusName, error)
	Get(ctx context.Context, id int) (*ent.EquipmentStatusName, error)
	GetByName(ctx context.Context, name string) (*ent.EquipmentStatusName, error)
	Delete(ctx context.Context, id int) (*ent.EquipmentStatusName, error)
}

type OrderRepository interface {
	List(ctx context.Context, ownerId *int, filter OrderFilter) ([]*ent.Order, error)
	OrdersTotal(ctx context.Context, ownerId *int) (int, error)
	Create(ctx context.Context, data *models.OrderCreateRequest, ownerId int, equipmentIDs []int) (*ent.Order, error)
	Update(ctx context.Context, id int, data *models.OrderUpdateRequest, ownerId int) (*ent.Order, error)
}

type OrderRepositoryWithFilter interface {
	OrdersByStatus(ctx context.Context, status string, limit, offset int,
		orderBy, orderColumn string) ([]*ent.Order, error)
	OrdersByStatusTotal(ctx context.Context, status string) (int, error)
	OrdersByPeriodAndStatus(ctx context.Context, from, to time.Time, status string, limit, offset int,
		orderBy, orderColumn string) ([]*ent.Order, error)
	OrdersByPeriodAndStatusTotal(ctx context.Context, from, to time.Time, status string) (int, error)
}

type OrderStatusRepository interface {
	StatusHistory(ctx context.Context, orderId int) ([]*ent.OrderStatus, error)
	UpdateStatus(ctx context.Context, userID int, status models.NewOrderStatus) error
	GetOrderCurrentStatus(ctx context.Context, orderId int) (*ent.OrderStatus, error)
	GetUserStatusHistory(ctx context.Context, userId int) ([]*ent.OrderStatus, error)
}

type OrderStatusNameRepository interface {
	ListOfOrderStatusNames(ctx context.Context) ([]*ent.OrderStatusName, error)
}

type PasswordResetRepository interface {
	CreateToken(ctx context.Context, token string, ttl time.Time, userID int) error
	GetToken(ctx context.Context, token string) (*ent.PasswordReset, error)
	DeleteToken(ctx context.Context, token string) error
}

type EmailConfirmRepository interface {
	CreateToken(ctx context.Context, token string, ttl time.Time, userID int, email string) error
	GetToken(ctx context.Context, token string) (*ent.EmailConfirm, error)
	DeleteToken(ctx context.Context, token string) error
}

type PetKindRepository interface {
	Create(ctx context.Context, ps models.PetKind) (*ent.PetKind, error)
	GetByID(ctx context.Context, id int) (*ent.PetKind, error)
	GetAll(ctx context.Context) ([]*ent.PetKind, error)
	Delete(ctx context.Context, id int) error
	Update(ctx context.Context, id int, newPetKind *models.PetKind) (*ent.PetKind, error)
}

type PetSizeRepository interface {
	Create(ctx context.Context, ps models.PetSize) (*ent.PetSize, error)
	GetByID(ctx context.Context, id int) (*ent.PetSize, error)
	GetAll(ctx context.Context) ([]*ent.PetSize, error)
	Delete(ctx context.Context, id int) error
	Update(ctx context.Context, id int, newPetSize *models.PetSize) (*ent.PetSize, error)
}

type PhotoRepository interface {
	CreatePhoto(ctx context.Context, p *ent.Photo) (*ent.Photo, error)
	PhotoByID(ctx context.Context, id string) (*ent.Photo, error)
	DeletePhotoByID(ctx context.Context, id string) error
}

type RegistrationConfirmRepository interface {
	CreateToken(ctx context.Context, token string, ttl time.Time, userID int) error
	GetToken(ctx context.Context, token string) (*ent.RegistrationConfirm, error)
	DeleteToken(ctx context.Context, token string) error
}

type RoleRepository interface {
	GetRoles(ctx context.Context) ([]*ent.Role, error)
}
type SubcategoryRepository interface {
	CreateSubcategory(ctx context.Context, categoryID int, newSubcategory models.NewSubcategory) (*ent.Subcategory, error)
	ListSubcategories(ctx context.Context, categoryID int) ([]*ent.Subcategory, error)
	SubcategoryByID(ctx context.Context, id int) (*ent.Subcategory, error)
	DeleteSubcategoryByID(ctx context.Context, id int) error
	UpdateSubcategory(ctx context.Context, id int, update models.NewSubcategory) (*ent.Subcategory, error)
}

type TokenRepository interface {
	CreateTokens(ctx context.Context, ownerID int, accessToken, refreshToken string) error
	DeleteTokensByRefreshToken(ctx context.Context, refreshToken string) error
	UpdateAccessToken(ctx context.Context, accessToken, refreshToken string) error
}

type UserRepository interface {
	SetUserRole(ctx context.Context, userId int, roleId int) error
	UserByLogin(ctx context.Context, login string) (*ent.User, error)
	ChangePasswordByLogin(ctx context.Context, login string, password string) error
	ChangeEmailByLogin(ctx context.Context, login string, email string) error
	CreateUser(ctx context.Context, data *models.UserRegister) (*ent.User, error)
	GetUserByLogin(ctx context.Context, login string) (*ent.User, error)
	GetUserByID(ctx context.Context, id int) (*ent.User, error)
	UpdateUserByID(ctx context.Context, id int, patch *models.PatchUserRequest) error
	UserList(ctx context.Context, limit, offset int, orderBy, orderColumn string) ([]*ent.User, error)
	Delete(ctx context.Context, userId int) error
	UsersListTotal(ctx context.Context) (int, error)
	ConfirmRegistration(ctx context.Context, login string) error
	UnConfirmRegistration(ctx context.Context, login string) error
	SetIsReadonly(ctx context.Context, id int, isReadonly bool) error
}
