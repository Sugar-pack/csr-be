package messages

// Keep in mind that these messages will be used by frontend team.
// So, if you change it, you should notify them.
var (
	MsgAllOk = "all ok"

	// Area

	ErrQueryTotalAreas = "failed to query total active areas"
	ErrQueryAreas      = "failed to query active areas"

	// Category

	ErrCreateCategory       = "cant create new category"
	ErrQueryTotalCategories = "cant get total amount of categories"
	ErrQueryCategories      = "cant get all categories"
	ErrGetCategory          = "failed to get category"
	ErrDeleteCategory       = "delete category failed"
	ErrUpdateCategory       = "cant update category"
	MsgCategoryDeleted      = "category deleted"

	// Email Confirmation

	ErrEmailConfirm   = "failed to verify email confirmation token"
	MsgEmailConfirmed = "you have successfully confirmed new email"

	// Equipment Periods

	ErrGetUnavailableEqStatus = "can't find unavailable equipment status dates"

	// Equipment Status Names

	ErrCreateEqStatus  = "can't create equipment status"
	ErrQueryEqStatuses = "can't get equipment statuses"
	ErrGetEqStatus     = "can't get equipment status"
	ErrDeleteEqStatus  = "can't delete equipment status"

	// Equipment Status

	ErrWrongEqStatus            = "wrong new equipment status, status should be only 'not available'"
	ErrGetEqStatusByID          = "can't find equipment status by provided id"
	ErrOrderAndUserByEqStatusID = "can't receive order and user data during checking equipment status"
	ErrUpdateEqStatus           = "can't update equipment status"

	// Equipment

	ErrCreateEquipment           = "error while creating equipment"
	ErrMapEquipment              = "error while mapping equipment"
	ErrGetEquipment              = "error while getting equipment"
	ErrEquipmentNotFound         = "equipment not found"
	ErrEquipmentArchive          = "error while archiving equipment"
	ErrEquipmentBlock            = "error while blocking equipment"
	ErrEquipmentUnblock          = "error while unblocking equipment"
	ErrDeleteEquipment           = "error while deleting equipment"
	ErrQueryTotalEquipments      = "error while getting total of all equipments"
	ErrQueryEquipments           = "error while getting all equipments"
	ErrUpdateEquipment           = "error while updating equipment"
	ErrFindEquipment             = "error while finding equipment"
	ErrEquipmentBlockForbidden   = "you don't have rights to block the equipment"
	ErrEquipmentUnblockForbidden = "you don't have rights to unblock the equipment"
	ErrStartDateAfterEnd         = "start date should be before end date"
	MsgEquipmentDeleted          = "equipment deleted"

	// Order Status

	ErrQueryOrderHistory                 = "can't get order history"
	ErrQueryOrderHistoryForbidden        = "you don't have rights to see this order"
	ErrCreateOrderStatusForbidden        = "you don't have rights to add a new status"
	ErrOrderStatusEmpty                  = "order status is empty"
	ErrGetOrderStatus                    = "can't get order current status"
	ErrUpdateOrderStatus                 = "can't update status"
	ErrQueryTotalOrdersByStatus          = "can't get total count of orders by status"
	ErrQueryOrdersByStatus               = "can't get orders by status"
	ErrQueryTotalOrdersByPeriodAndStatus = "can't get total amount of orders by period and status"
	ErrQueryOrdersByPeriodAndStatus      = "can't get orders by period and status"
	ErrMapOrderStatus                    = "can't map order status name"
	ErrQueryStatusNames                  = "can't get all status names"

	// Order

	ErrOrderNotFound       = "no order with such id"
	ErrMapOrder            = "can't map order"
	ErrQueryOrders         = "can't get orders"
	ErrQueryTotalOrders    = "error while getting total of orders"
	ErrUpdateOrder         = "update order failed"
	ErrEquipmentIsNotFree  = "requested equipment is not free"
	ErrCheckEqStatusFailed = "error while checking if equipment is available for period"
	ErrSmallRentPeriod     = "small rent period"

	// Password Reset

	ErrLoginRequired          = "login is required"
	MsgPasswordResetSuccesful = "check your email for a reset link"

	// Pet Kind

	ErrCreatePetKind   = "error while creating pet kind"
	ErrGetPetKind      = "error while getting pet kind"
	ErrPetKindNotFound = "no pet kind found"
	ErrUpdatePetKind   = "error while updating pet kind"
	ErrDeletePetKind   = "error while deleting pet kind"
	MsgPetKindDeleted  = "pet kind deleted"

	// Pet Size

	ErrCreatePetSize        = "error while creating pet size"
	ErrPetSizeAlreadyExists = "error while creating pet size: the name already exist"
	ErrGetPetSize           = "error while getting pet size"
	ErrPetSizeNotFound      = "no pet size found"
	ErrUpdatePetSize        = "error while updating pet size"
	ErrDeletePetSize        = "error while deleting pet size"
	MsgPetSizeDeleted       = "pet size deleted"

	// Photo

	ErrCreatePhoto  = "failed to save photo"
	ErrFileEmpty    = "File is empty"
	ErrWrongFormat  = "Wrong file format. File should be jpg or jpeg"
	ErrGetPhoto     = "failed to get photo"
	ErrDeletePhoto  = "failed to delete photo"
	MsgPhotoDeleted = "photo deleted"

	// Registration Confirm

	ErrRegistrationAlreadyConfirmed = "Registration is already confirmed."
	ErrRegistrationCannotFindUser   = "Can't find this user, registration confirmation link wasn't send"
	ErrRegistrationCannotSend       = "Can't send registration confirmation link. Please try again later"
	ErrFailedToConfirm              = "Failed to verify confirmation token. Please try again later"
	MsgConfirmationNotRequired      = "Confirmation link was not sent to email, sending parameter was set to false and not required"
	MsgConfirmationSent             = "Confirmation link was sent"
	MsgRegistrationConfirmed        = "You have successfully confirmed registration"

	// Roles

	ErrQueryRoles = "can't get all roles"

	// Subcategory

	ErrCreateSubcategory   = "failed to create new subcategory"
	ErrMapSubcategory      = "failed to map new subcategory"
	ErrQuerySCatByCategory = "failed to list subcategories by category id"
	ErrGetSubcategory      = "failed to get subcategory"
	ErrDeleteSubcategory   = "failed to delete subcategory"
	ErrUpdateSubcategory   = "failed to update subcategory"
	MsgSubcategoryDeleted  = "subcategory deleted"

	// User

	ErrInvalidLoginOrPass   = "invalid login or password"
	ErrLoginInUse           = "login is already used"
	ErrCreateUser           = "failed to create user"
	ErrInvalidToken         = "token is invalid"
	ErrTokenRefresh         = "error while refreshing token"
	ErrMapUser              = "map user error"
	ErrUserNotFound         = "can't find user by id"
	ErrUpdateUser           = "can't update user"
	ErrRoleRequired         = "role id is required"
	ErrSetUserRole          = "set user role error"
	ErrQueryTotalUsers      = "failed get user total amount"
	ErrQueryUsers           = "failed to get user list"
	ErrDeleteUser           = "can't delete user"
	ErrDeleteUserNotRO      = "user must be readonly for deletion"
	ErrUserPasswordChange   = "error while changing password"
	ErrWrongPassword        = "wrong password"
	ErrPasswordsAreSame     = "old and new passwords are the same"
	ErrPasswordPatchEmpty   = "password patch is empty"
	ErrUpdateROAccess       = "error while updating readonly access"
	ErrChangeEmail          = "error while changing email"
	ErrEmailPatchEmpty      = "email patch is empty"
	ErrNewEmailConfirmation = "can't send link for confirmation new email"
	MsgLogoutSuccessful     = "successfully logged out"
	MsgRoleAssigned         = "role assigned"
)
