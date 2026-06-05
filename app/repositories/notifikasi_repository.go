package repositories

import (
	// "time"
	"monitoring-service/app/models"
)

func (r *Main) GetJadwalForReminder() ([]models.JadwalImunisasiJoin, error) {

	var data []models.JadwalImunisasiJoin

	err := r.postgres.
		Table("jadwal_imunisasi_anak j").
		Select(`
			j.id as jadwal_id,
			j.id_anak,
			a.id as anak_id,
			p.nama_lengkap as nama_anak,
			j.tanggal_estimasi,
			j.id_status_jadwal as status_id,
			s.nama_status,
			dv.nama_dosis
		`).
		Joins("JOIN anak a ON a.id = j.id_anak").
		Joins("JOIN penduduk p ON p.id = a.penduduk_id").
		Joins("JOIN dosis_vaksin dv ON dv.id = j.id_dosis_vaksin").
		Joins("JOIN status_jadwal s ON s.id = j.id_status_jadwal").
		Where("j.tanggal_estimasi IS NOT NULL").
		Where("id_status_jadwal IN ?", []int{1, 2}).
		Find(&data).Error

	return data, err
}
func (r *Main) GetFCMTokensByUserID(userID uint) ([]string, error) {
	var tokens []string

	err := r.postgres.
		Table("pengguna").
		Select("perangkat.fcm_token").
		Joins("JOIN perangkat ON perangkat.id_pengguna = pengguna.id").
		Where("pengguna.id = ?", userID).
		Scan(&tokens).Error

	return tokens, err
}

func (r *Main) CreateNotifikasi(n models.Notifikasi) error {
	return r.postgres.Table("notifikasi").Create(&n).Error
}

func (r *Main) UpdateReminderFlag(jadwalID uint, field string) error {
	return r.postgres.
		Table("jadwal_imunisasi_anak").
		Where("id = ?", jadwalID).
		Update(field, true).Error
}

func (r *Main) GetUserIDByAnakID(anakID uint) (uint, error) {

	var userID uint

	err := r.postgres.
		Table("anak a").
		Select("p.id as user_id").
		Joins("JOIN kehamilan k ON k.id = a.kehamilan_id").
		Joins("JOIN ibu i ON i.id = k.ibu_id").
		Joins("JOIN penduduk pd ON pd.id = i.penduduk_id").
		Joins("JOIN pengguna p ON p.penduduk_id = pd.id").
		Where("a.id = ?", anakID).
		Scan(&userID).Error

	return userID, err
}

func (r *Main) UpdateStatusJadwal(
	jadwalID uint,
	statusID uint,
) error {

	return r.postgres.
		Table("jadwal_imunisasi_anak").
		Where("id = ?", jadwalID).
		Update(
			"id_status_jadwal",
			statusID,
		).Error
}

func (r *Main) MarkSent(jadwalID uint, tipe string) error {

	update := map[string]interface{}{}

	switch tipe {
	case "h7":
		update["is_sent_h7"] = true
	case "h3":
		update["is_sent_h3"] = true
	case "h":
		update["is_sent_h"] = true
	}

	return r.postgres.
		Table("jadwal_imunisasi_anak").
		Where("id = ?", jadwalID).
		Updates(update).Error
}

func (r *Main) InsertNotifikasi(
	anakID uint,
	title string,
	message string,
) error {

	return r.postgres.Exec(`
		INSERT INTO notifikasi (id_anak, judul, pesan, created_at)
		VALUES (?, ?, ?, NOW())
	`, anakID, title, message).Error
}

func (r *Main) GetFCMTokensByAnakID(anakID uint) ([]string, error) {

	var tokens []string

	err := r.postgres.
		Table("pengguna p").
		Select("d.fcm_token").
		Joins("JOIN anak a ON a.id = ?", anakID).
		Joins("JOIN pengguna u ON u.id = p.id").
		Joins("JOIN perangkat d ON d.id_pengguna = u.id").
		Scan(&tokens).Error

	return tokens, err
}
