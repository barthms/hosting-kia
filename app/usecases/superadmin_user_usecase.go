package usecases

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"monitoring-service/app/models"
	"monitoring-service/app/repositories"
	"monitoring-service/pkg/customerror"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SuperadminCreateBidanUserRequest struct {
	PendudukID  int32  `json:"penduduk_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	DesaID      *int32 `json:"desa_id,omitempty"`
	NoSTR       string `json:"no_str"`
	NoSIPB      string `json:"no_sipb"`
}

type SuperadminCreateAdminDesaUserRequest struct {
	PendudukID  *int32 `json:"penduduk_id,omitempty"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	DesaID      *int32 `json:"desa_id,omitempty"`
}

type SuperadminCreateKaderUserRequest struct {
	PendudukID  int32  `json:"penduduk_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	DesaID      *int32 `json:"desa_id,omitempty"`
	PosyanduID  *int64 `json:"posyandu_id,omitempty"`
}

type SuperadminCreateUserRequest struct {
	PendudukID  *int64 `json:"penduduk_id,omitempty"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	RoleName    string `json:"role_name"`
	DesaID      *int32 `json:"desa_id,omitempty"`
}

type SuperadminUpdateUserRoleRequest struct {
	RoleName string `json:"role_name"`
}

type SuperadminResetPasswordRequest struct {
	Password string `json:"password"`
}

type SuperadminUserUsecase struct {
	repo *repositories.Main
}

type SuperadminUserActionResponse struct {
	User  *models.User  `json:"user"`
	Bidan *models.Bidan `json:"bidan,omitempty"`
	Kader *models.Kader `json:"kader,omitempty"`
}

func NewSuperadminUserUsecase(repo *repositories.Main) *SuperadminUserUsecase {
	return &SuperadminUserUsecase{repo: repo}
}

func (u *SuperadminUserUsecase) isNotFound(err error) bool {
	return err != nil && errors.Is(err, gorm.ErrRecordNotFound)
}

func (u *SuperadminUserUsecase) getPendudukDesaID(pendudukID int32) (*int32, error) {
	penduduk, err := u.repo.Kependudukan.FindByID(pendudukID)
	if err != nil {
		return nil, customerror.NewNotFoundError("penduduk tidak ditemukan")
	}
	if penduduk.DesaID == nil || *penduduk.DesaID == 0 {
		return nil, customerror.NewBadRequestError("penduduk belum memiliki desa_id")
	}
	if _, err := u.repo.Desa.GetByID(*penduduk.DesaID); err != nil {
		return nil, customerror.NewNotFoundError("desa pada data penduduk tidak ditemukan")
	}
	return penduduk.DesaID, nil
}

func (u *SuperadminUserUsecase) validateDesaConsistency(requestDesaID, actualDesaID *int32) error {
	if requestDesaID == nil {
		return nil
	}
	if *requestDesaID == 0 {
		return customerror.NewBadRequestError("desa_id tidak valid")
	}
	if _, err := u.repo.Desa.GetByID(*requestDesaID); err != nil {
		return customerror.NewNotFoundError("desa tidak ditemukan")
	}
	if actualDesaID != nil && *actualDesaID != *requestDesaID {
		return customerror.NewBadRequestError("desa_id harus sama dengan desa pada data penduduk")
	}
	return nil
}

func (u *SuperadminUserUsecase) validateBaseUserFields(name, email, phoneNumber, password string) (string, string, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	phoneNumber = strings.TrimSpace(phoneNumber)
	password = strings.TrimSpace(password)

	if name == "" || email == "" || phoneNumber == "" || password == "" {
		return "", "", customerror.NewBadRequestError("name, email, phone_number, dan password wajib diisi")
	}
	if len(password) < 8 {
		return "", "", customerror.NewBadRequestError("password minimal 8 karakter")
	}

	normalizedPhone, err := normalizePhoneNumber(phoneNumber)
	if err != nil {
		return "", "", err
	}

	return email, normalizedPhone, nil
}

func (u *SuperadminUserUsecase) preparePassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		buf := make([]byte, 8)
		if _, err := rand.Read(buf); err != nil {
			return "", customerror.NewInternalServiceError("gagal membuat password sementara")
		}
		password = hex.EncodeToString(buf)
	}
	if len(password) < 8 {
		return "", customerror.NewBadRequestError("password minimal 8 karakter")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", customerror.NewInternalServiceError("gagal memproses password")
	}

	return string(hashedPassword), nil
}

func (u *SuperadminUserUsecase) ListUsers(search, role, desa string) ([]repositories.UserListItem, error) {
	return u.repo.User.List(search, role, desa)
}

func (u *SuperadminUserUsecase) ListPenduduk(search string) ([]repositories.EligiblePendudukItem, error) {
	return u.repo.Kependudukan.ListAvailableForSuperadmin(search)
}

func (u *SuperadminUserUsecase) GetUser(id int32) (*models.User, error) {
	if id == 0 {
		return nil, customerror.NewBadRequestError("id user tidak valid")
	}
	user, err := u.repo.User.FindByID(id)
	if err != nil {
		return nil, customerror.NewNotFoundError("user tidak ditemukan")
	}
	return user, nil
}

func (u *SuperadminUserUsecase) CreateBidanUser(req *SuperadminCreateBidanUserRequest) (*SuperadminUserActionResponse, error) {
	if req == nil {
		return nil, customerror.NewBadRequestError("request tidak valid")
	}
	if req.PendudukID == 0 {
		return nil, customerror.NewBadRequestError("penduduk_id wajib diisi")
	}
	if strings.TrimSpace(req.NoSIPB) == "" {
		return nil, customerror.NewBadRequestError("no_sipb wajib diisi")
	}

	email, normalizedPhone, err := u.validateBaseUserFields(req.Name, req.Email, req.PhoneNumber, req.Password)
	if err != nil {
		return nil, err
	}
	pendudukDesaID, err := u.getPendudukDesaID(req.PendudukID)
	if err != nil {
		return nil, err
	}
	if err := u.validateDesaConsistency(req.DesaID, pendudukDesaID); err != nil {
		return nil, err
	}

	if _, err := u.repo.Bidan.FindByPendudukID(req.PendudukID); err == nil {
		return nil, customerror.NewConflictError("penduduk sudah terdaftar sebagai bidan")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi data bidan")
	}
	if _, err := u.repo.Kader.FindByPendudukID(req.PendudukID); err == nil {
		return nil, customerror.NewConflictError("penduduk sudah terdaftar sebagai kader")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi data kader")
	}
	if _, err := u.repo.User.FindByPendudukID(int64(req.PendudukID)); err == nil {
		return nil, customerror.NewConflictError("penduduk sudah memiliki akun pengguna")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi akun pengguna")
	}
	if _, err := u.repo.User.FindByEmail(email); err == nil {
		return nil, customerror.NewConflictError("email sudah terdaftar")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi email")
	}
	if _, err := u.repo.User.FindByPhoneNumber(normalizedPhone); err == nil {
		return nil, customerror.NewConflictError("nomor hp sudah terdaftar")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi nomor hp")
	}

	role, err := u.repo.Role.FindByName("Bidan")
	if err != nil {
		return nil, customerror.NewNotFoundError("role bidan tidak ditemukan")
	}
	hashedPassword, err := u.preparePassword(req.Password)
	if err != nil {
		return nil, err
	}

	var createdUser models.User
	var createdBidan models.Bidan
	pendudukID64 := int64(req.PendudukID)
	err = u.repo.DB().Transaction(func(tx *gorm.DB) error {
		createdUser = models.User{
			Name:        strings.TrimSpace(req.Name),
			Email:       email,
			PhoneNumber: normalizedPhone,
			IsActive:    true,
			Password:    hashedPassword,
			RoleID:      role.ID,
			PendudukID:  &pendudukID64,
		}
		if err := tx.Create(&createdUser).Error; err != nil {
			return err
		}
		createdBidan = models.Bidan{
			PendudukID: req.PendudukID,
			DesaID:     pendudukDesaID,
			NoSTR:      strings.TrimSpace(req.NoSTR),
			NoSIPB:     strings.TrimSpace(req.NoSIPB),
			Status:     "aktif",
		}
		if err := tx.Create(&createdBidan).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, customerror.NewConflictError("data bidan atau akun pengguna sudah terdaftar")
		}
		return nil, customerror.NewInternalServiceError("gagal membuat akun bidan: " + err.Error())
	}
	return &SuperadminUserActionResponse{User: &createdUser, Bidan: &createdBidan}, nil
}

func (u *SuperadminUserUsecase) CreateAdminDesaUser(req *SuperadminCreateAdminDesaUserRequest) (*models.User, error) {
	if req == nil {
		return nil, customerror.NewBadRequestError("request tidak valid")
	}
	if req.PendudukID == nil || *req.PendudukID == 0 {
		return nil, customerror.NewBadRequestError("penduduk_id wajib diisi agar desa mengikuti data kependudukan")
	}

	email, normalizedPhone, err := u.validateBaseUserFields(req.Name, req.Email, req.PhoneNumber, req.Password)
	if err != nil {
		return nil, err
	}
	pendudukDesaID, err := u.getPendudukDesaID(*req.PendudukID)
	if err != nil {
		return nil, err
	}
	if err := u.validateDesaConsistency(req.DesaID, pendudukDesaID); err != nil {
		return nil, err
	}

	pendudukID64 := int64(*req.PendudukID)
	pendudukID := &pendudukID64
	if _, err := u.repo.User.FindByEmail(email); err == nil {
		return nil, customerror.NewConflictError("email sudah terdaftar")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi email")
	}
	if _, err := u.repo.User.FindByPhoneNumber(normalizedPhone); err == nil {
		return nil, customerror.NewConflictError("nomor hp sudah terdaftar")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi nomor hp")
	}
	if _, err := u.repo.User.FindByPendudukID(*pendudukID); err == nil {
		return nil, customerror.NewConflictError("penduduk sudah memiliki akun pengguna")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi akun pengguna")
	}

	role, err := u.repo.Role.FindByName("Admin")
	if err != nil {
		role, err = u.repo.Role.FindByName("Admin_desa")
		if err != nil {
			return nil, customerror.NewNotFoundError("role admin tidak ditemukan")
		}
	}
	hashedPassword, err := u.preparePassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:        strings.TrimSpace(req.Name),
		Email:       email,
		PhoneNumber: normalizedPhone,
		IsActive:    true,
		Password:    hashedPassword,
		RoleID:      role.ID,
		PendudukID:  pendudukID,
	}
	if err := u.repo.DB().Create(user).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, customerror.NewConflictError("data user sudah terdaftar")
		}
		return nil, customerror.NewInternalServiceError("gagal membuat akun admin desa")
	}
	return user, nil
}

func (u *SuperadminUserUsecase) CreateKaderUser(req *SuperadminCreateKaderUserRequest) (*SuperadminUserActionResponse, error) {
	if req == nil {
		return nil, customerror.NewBadRequestError("request tidak valid")
	}
	if req.PendudukID == 0 {
		return nil, customerror.NewBadRequestError("penduduk_id wajib diisi")
	}

	email, normalizedPhone, err := u.validateBaseUserFields(req.Name, req.Email, req.PhoneNumber, req.Password)
	if err != nil {
		return nil, err
	}
	pendudukDesaID, err := u.getPendudukDesaID(req.PendudukID)
	if err != nil {
		return nil, err
	}
	if err := u.validateDesaConsistency(req.DesaID, pendudukDesaID); err != nil {
		return nil, err
	}

	if _, err := u.repo.Bidan.FindByPendudukID(req.PendudukID); err == nil {
		return nil, customerror.NewConflictError("penduduk sudah terdaftar sebagai bidan")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi data bidan")
	}
	if _, err := u.repo.Kader.FindByPendudukID(req.PendudukID); err == nil {
		return nil, customerror.NewConflictError("penduduk sudah terdaftar sebagai kader")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi data kader")
	}
	if _, err := u.repo.User.FindByPendudukID(int64(req.PendudukID)); err == nil {
		return nil, customerror.NewConflictError("penduduk sudah memiliki akun pengguna")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi akun pengguna")
	}
	if _, err := u.repo.User.FindByEmail(email); err == nil {
		return nil, customerror.NewConflictError("email sudah terdaftar")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi email")
	}
	if _, err := u.repo.User.FindByPhoneNumber(normalizedPhone); err == nil {
		return nil, customerror.NewConflictError("nomor hp sudah terdaftar")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi nomor hp")
	}

	role, err := u.repo.Role.FindByName("Kader")
	if err != nil {
		return nil, customerror.NewNotFoundError("role kader tidak ditemukan")
	}
	hashedPassword, err := u.preparePassword(req.Password)
	if err != nil {
		return nil, err
	}

	pendudukID64 := int64(req.PendudukID)
	var createdUser models.User
	var createdKader models.Kader
	err = u.repo.DB().Transaction(func(tx *gorm.DB) error {
		createdUser = models.User{
			Name:        strings.TrimSpace(req.Name),
			Email:       email,
			PhoneNumber: normalizedPhone,
			IsActive:    true,
			Password:    hashedPassword,
			RoleID:      role.ID,
			PendudukID:  &pendudukID64,
		}
		if err := tx.Create(&createdUser).Error; err != nil {
			return err
		}
		createdKader = models.Kader{
			PendudukID: req.PendudukID,
			PosyanduID: req.PosyanduID,
			Status:     "aktif",
		}
		if err := tx.Create(&createdKader).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, customerror.NewConflictError("data kader atau akun pengguna sudah terdaftar")
		}
		return nil, customerror.NewInternalServiceError("gagal membuat akun kader: " + err.Error())
	}
	return &SuperadminUserActionResponse{User: &createdUser, Kader: &createdKader}, nil
}

func (u *SuperadminUserUsecase) CreateUser(req *SuperadminCreateUserRequest) (*models.User, error) {
	if req == nil {
		return nil, customerror.NewBadRequestError("request tidak valid")
	}
	roleName := strings.TrimSpace(req.RoleName)
	if roleName == "" {
		return nil, customerror.NewBadRequestError("role_name wajib diisi")
	}
	if normalizeKey(roleName) == "admin" || normalizeKey(roleName) == "admindesa" || normalizeKey(roleName) == "bidan" || normalizeKey(roleName) == "kader" || normalizeKey(roleName) == "superadmin" {
		return nil, customerror.NewBadRequestError("role admin, bidan, kader, dan superadmin diatur dari halaman kelola akun khusus")
	}

	email, normalizedPhone, err := u.validateBaseUserFields(req.Name, req.Email, req.PhoneNumber, req.Password)
	if err != nil {
		return nil, err
	}

	var pendudukID *int64
	var pendudukDesaID *int32
	if req.PendudukID != nil && *req.PendudukID > 0 {
		pendudukDesaID, err = u.getPendudukDesaID(int32(*req.PendudukID))
		if err != nil {
			return nil, err
		}
		pendudukID = req.PendudukID
	}
	if err := u.validateDesaConsistency(req.DesaID, pendudukDesaID); err != nil {
		return nil, err
	}

	if _, err := u.repo.User.FindByEmail(email); err == nil {
		return nil, customerror.NewConflictError("email sudah terdaftar")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi email")
	}
	if _, err := u.repo.User.FindByPhoneNumber(normalizedPhone); err == nil {
		return nil, customerror.NewConflictError("nomor hp sudah terdaftar")
	} else if !u.isNotFound(err) {
		return nil, customerror.NewInternalServiceError("gagal memvalidasi nomor hp")
	}
	if pendudukID != nil {
		if _, err := u.repo.User.FindByPendudukID(*pendudukID); err == nil {
			return nil, customerror.NewConflictError("penduduk sudah memiliki akun pengguna")
		} else if !u.isNotFound(err) {
			return nil, customerror.NewInternalServiceError("gagal memvalidasi akun pengguna")
		}
		if _, err := u.repo.Bidan.FindByPendudukID(int32(*pendudukID)); err == nil {
			return nil, customerror.NewConflictError("penduduk sudah terdaftar sebagai bidan")
		} else if !u.isNotFound(err) {
			return nil, customerror.NewInternalServiceError("gagal memvalidasi data bidan")
		}
		if _, err := u.repo.Kader.FindByPendudukID(int32(*pendudukID)); err == nil {
			return nil, customerror.NewConflictError("penduduk sudah terdaftar sebagai kader")
		} else if !u.isNotFound(err) {
			return nil, customerror.NewInternalServiceError("gagal memvalidasi data kader")
		}
	}

	role, err := u.repo.Role.FindByName(normalizeRoleName(roleName))
	if err != nil {
		return nil, customerror.NewNotFoundError("role tidak ditemukan")
	}
	hashedPassword, err := u.preparePassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:        strings.TrimSpace(req.Name),
		Email:       email,
		PhoneNumber: normalizedPhone,
		IsActive:    true,
		Password:    hashedPassword,
		RoleID:      role.ID,
		PendudukID:  pendudukID,
	}
	if err := u.repo.DB().Create(user).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, customerror.NewConflictError("data user sudah terdaftar")
		}
		return nil, customerror.NewInternalServiceError("gagal membuat akun user")
	}
	return user, nil
}

func (u *SuperadminUserUsecase) ResetPassword(id int32, req *SuperadminResetPasswordRequest) (*models.User, error) {
	if id == 0 {
		return nil, customerror.NewBadRequestError("id user tidak valid")
	}
	if req == nil {
		return nil, customerror.NewBadRequestError("request tidak valid")
	}
	user, err := u.repo.User.FindByID(id)
	if err != nil {
		return nil, customerror.NewNotFoundError("user tidak ditemukan")
	}
	hashedPassword, err := u.preparePassword(req.Password)
	if err != nil {
		return nil, err
	}
	if err := u.repo.User.SetPassword(id, hashedPassword); err != nil {
		return nil, customerror.NewInternalServiceError("gagal memperbarui password")
	}
	user.Password = hashedPassword
	return user, nil
}

func (u *SuperadminUserUsecase) UpdateUserRole(id int32, req *SuperadminUpdateUserRoleRequest) (*models.User, error) {
	if id == 0 {
		return nil, customerror.NewBadRequestError("id user tidak valid")
	}
	if req == nil {
		return nil, customerror.NewBadRequestError("request tidak valid")
	}
	roleName := strings.TrimSpace(req.RoleName)
	if roleName == "" {
		return nil, customerror.NewBadRequestError("role_name wajib diisi")
	}
	normalizedRole := normalizeRoleName(roleName)
	if normalizedRole == "Admin" || normalizedRole == "Bidan" || normalizedRole == "Kader" || normalizedRole == "Superadmin" {
		return nil, customerror.NewBadRequestError("role admin, bidan, kader, dan superadmin diatur dari halaman kelola akun desa")
	}
	user, err := u.repo.User.FindByID(id)
	if err != nil {
		return nil, customerror.NewNotFoundError("user tidak ditemukan")
	}
	if strings.EqualFold(strings.TrimSpace(user.Role.Name), "Superadmin") {
		return nil, customerror.NewBadRequestError("akun superadmin tidak dapat diubah")
	}
	role, err := u.repo.Role.FindByName(normalizedRole)
	if err != nil {
		return nil, customerror.NewNotFoundError("role tidak ditemukan")
	}
	if err := u.repo.DB().Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"role_id":    role.ID,
		"updated_at": gorm.Expr("NOW()"),
	}).Error; err != nil {
		return nil, customerror.NewInternalServiceError("gagal memperbarui role user")
	}
	user.Role = *role
	user.RoleID = role.ID
	return user, nil
}

func (u *SuperadminUserUsecase) DeactivateUser(id int32) (*models.User, error) {
	if id == 0 {
		return nil, customerror.NewBadRequestError("id user tidak valid")
	}
	user, err := u.repo.User.FindByID(id)
	if err != nil {
		return nil, customerror.NewNotFoundError("user tidak ditemukan")
	}
	if strings.EqualFold(strings.TrimSpace(user.Role.Name), "Superadmin") {
		return nil, customerror.NewBadRequestError("akun superadmin tidak dapat dinonaktifkan")
	}
	if err := u.repo.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": gorm.Expr("NOW()"),
		}).Error; err != nil {
			return err
		}
		switch strings.TrimSpace(user.Role.Name) {
		case "Bidan":
			if user.PendudukID != nil {
				if err := tx.Model(&models.Bidan{}).Where("penduduk_id = ? AND deleted_at IS NULL", *user.PendudukID).Updates(map[string]interface{}{
					"status":     "nonaktif",
					"updated_at": gorm.Expr("NOW()"),
				}).Error; err != nil {
					return err
				}
			}
		case "Kader":
			if user.PendudukID != nil {
				if err := tx.Model(&models.Kader{}).Where("penduduk_id = ? AND deleted_at IS NULL", *user.PendudukID).Updates(map[string]interface{}{
					"status":     "nonaktif",
					"updated_at": gorm.Expr("NOW()"),
				}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return nil, customerror.NewInternalServiceError("gagal menonaktifkan user")
	}
	user.IsActive = false
	return user, nil
}

func (u *SuperadminUserUsecase) ActivateUser(id int32) (*models.User, error) {
	if id == 0 {
		return nil, customerror.NewBadRequestError("id user tidak valid")
	}
	user, err := u.repo.User.FindByID(id)
	if err != nil {
		return nil, customerror.NewNotFoundError("user tidak ditemukan")
	}
	if strings.EqualFold(strings.TrimSpace(user.Role.Name), "Superadmin") {
		return nil, customerror.NewBadRequestError("akun superadmin tidak dapat diubah statusnya")
	}
	if err := u.repo.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
			"is_active":  true,
			"updated_at": gorm.Expr("NOW()"),
		}).Error; err != nil {
			return err
		}
		switch strings.TrimSpace(user.Role.Name) {
		case "Bidan":
			if user.PendudukID != nil {
				if err := tx.Model(&models.Bidan{}).Where("penduduk_id = ? AND deleted_at IS NULL", *user.PendudukID).Updates(map[string]interface{}{
					"status":     "aktif",
					"updated_at": gorm.Expr("NOW()"),
				}).Error; err != nil {
					return err
				}
			}
		case "Kader":
			if user.PendudukID != nil {
				if err := tx.Model(&models.Kader{}).Where("penduduk_id = ? AND deleted_at IS NULL", *user.PendudukID).Updates(map[string]interface{}{
					"status":     "aktif",
					"updated_at": gorm.Expr("NOW()"),
				}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return nil, customerror.NewInternalServiceError("gagal mengaktifkan user")
	}
	user.IsActive = true
	return user, nil
}
