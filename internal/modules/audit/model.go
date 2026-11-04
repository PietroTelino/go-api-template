package audit

type Action string

const (
	ActionUserRegister         Action = "USER_REGISTER"
	ActionUserLogin            Action = "USER_LOGIN"
	ActionUserLogout           Action = "USER_LOGOUT"
	ActionUserLogoutAll        Action = "USER_LOGOUT_ALL"
	ActionUserCreate           Action = "USER_CREATE"
	ActionUserUpdate           Action = "USER_UPDATE"
	ActionUserDelete           Action = "USER_DELETE"
	ActionUserSelfDelete       Action = "USER_SELF_DELETE"
	ActionUserInactivate       Action = "USER_INACTIVATE"
	ActionUserReactivate       Action = "USER_REACTIVATE"
	ActionUserRestore          Action = "USER_RESTORE"
	ActionUserHardDelete       Action = "USER_HARD_DELETE"
	ActionPasswordChange       Action = "PASSWORD_CHANGE"
	ActionPasswordResetRequest Action = "PASSWORD_RESET_REQUEST"
	ActionPasswordResetConfirm Action = "PASSWORD_RESET_CONFIRM"
	ActionPasswordAdminReset   Action = "PASSWORD_ADMIN_RESET"
	ActionPreferencesUpdate    Action = "PREFERENCES_UPDATE"
	ActionSessionLogout        Action = "SESSION_LOGOUT"
)

type Entity string

const (
	EntityUser        Entity = "USER"
	EntityPassword    Entity = "PASSWORD"
	EntityPreferences Entity = "PREFERENCES"
	EntitySession     Entity = "SESSION"
)

type Entry struct {
	UserID        *string
	PerformedByID *string
	Action        Action
	Entity        Entity
	EntityID      *string
	Details       map[string]any
	IPAddress     *string
	UserAgent     *string
}
