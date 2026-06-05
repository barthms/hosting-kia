package repositories

import (
	"context"
	"errors"
	"monitoring-service/app/models"

	"gorm.io/gorm"
)

type PemeriksaanRepository interface {
    CreatePemeriksaan(ctx context.Context, p *models.Pemeriksaan) error
    GetRiwayatByPenduduk(ctx context.Context, pendudukID uint, kelompok string) ([]models.Pemeriksaan, error)
    GetPemeriksaanByID(ctx context.Context, id uint) (*models.Pemeriksaan, error)
    GetPendudukByID(ctx context.Context, id uint) (*models.Kependudukan, error)
    GetUserByID(ctx context.Context, id uint) (*models.User, error)
	CountDistinctPendudukByKelompokAndIDs(kelompok string, pendudukIDs []int32) (int64, error)
	GetLatestRiskCountByPendudukIDs(kelompok string, pendudukIDs []int32) (map[string]int, error)
	GetPendudukByRisk(kelompok string, risiko string, desaID *int32, role string) ([]models.PendudukRiskResponse, error)
	GetRiwayatByPendudukID(ctx context.Context, pendudukID uint) ([]models.Pemeriksaan, error)
	GetLatestPemeriksaanByPenduduk(ctx context.Context, pendudukID uint, kelompok string) (*models.Pemeriksaan, error)
}

type pemeriksaanRepository struct {
	db *gorm.DB
}

func NewPemeriksaanRepository(db *gorm.DB) PemeriksaanRepository {
	return &pemeriksaanRepository{db: db}
}

func (r *pemeriksaanRepository) CreatePemeriksaan(ctx context.Context, p *models.Pemeriksaan) error {
	return r.db.WithContext(ctx).Create(p).Error
}

// GetCakupanAktual menghitung jumlah penduduk unik dan total pemeriksaan per kelompok
// berdasarkan data di tabel pemeriksaans, dengan filter desa (opsional) dan status penduduk aktif.
func (r *pemeriksaanRepository) GetRiwayatByPenduduk(ctx context.Context, pendudukID uint, kelompok string) ([]models.Pemeriksaan, error) {
	var list []models.Pemeriksaan
	query := r.db.WithContext(ctx).Where("penduduk_id = ?", pendudukID)
	if kelompok != "" {
		query = query.Where("kelompok = ?", kelompok)
	}
	err := query.Order("tanggal_pemeriksaan desc").Find(&list).Error
	return list, err
}

func (r *pemeriksaanRepository) GetPemeriksaanByID(ctx context.Context, id uint) (*models.Pemeriksaan, error) {
	var p models.Pemeriksaan
	err := r.db.WithContext(ctx).Preload("Penduduk").First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *pemeriksaanRepository) GetPendudukByID(ctx context.Context, id uint) (*models.Kependudukan, error) {
	var penduduk models.Kependudukan
	err := r.db.WithContext(ctx).First(&penduduk, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &penduduk, err
}

func (r *pemeriksaanRepository) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *pemeriksaanRepository) CountDistinctPendudukByKelompokAndIDs(kelompok string, pendudukIDs []int32) (int64, error) {
	if len(pendudukIDs) == 0 {
		return 0, nil
	}
	var count int64
	err := r.db.Model(&models.Pemeriksaan{}).
		Where("kelompok = ? AND penduduk_id IN ?", kelompok, pendudukIDs).
		Distinct("penduduk_id").
		Count(&count).Error
	return count, err
}
func (r *pemeriksaanRepository) GetLatestRiskCountByPendudukIDs(kelompok string, pendudukIDs []int32) (map[string]int, error) {
	if len(pendudukIDs) == 0 {
		return map[string]int{"Rendah": 0, "Sedang": 0, "Tinggi": 0}, nil
	}
	
	// Subquery untuk mendapatkan pemeriksaan terbaru per penduduk (berdasarkan TanggalPemeriksaan)
	subQuery := r.db.Model(&models.Pemeriksaan{}).
		Select("penduduk_id, MAX(tanggal_pemeriksaan) as max_tanggal").
		Where("kelompok = ? AND penduduk_id IN ?", kelompok, pendudukIDs).
		Group("penduduk_id")
	
	type Result struct {
		KategoriRisiko string
		Count          int
	}
	var results []Result
	err := r.db.Table("pemeriksaans as p").
		Select("p.kategori_risiko, COUNT(*) as count").
		Joins("INNER JOIN (?) as latest ON p.penduduk_id = latest.penduduk_id AND p.tanggal_pemeriksaan = latest.max_tanggal", subQuery).
		Where("p.kelompok = ?", kelompok).
		Group("p.kategori_risiko").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	
	riskMap := map[string]int{"Rendah": 0, "Sedang": 0, "Tinggi": 0}
	for _, res := range results {
		riskMap[res.KategoriRisiko] = res.Count
	}
	return riskMap, nil
}
func (r *pemeriksaanRepository) GetPendudukByRisk(kategori string, risiko string, desaID *int32, role string) ([]models.PendudukRiskResponse, error) {
    // Gunakan DISTINCT ON untuk mengambil satu baris terbaru per penduduk
    query := r.db.Table("pemeriksaans as p").
        Select("DISTINCT ON (p.penduduk_id) p.penduduk_id, k.id, k.nama_lengkap, k.nik, k.dusun, EXTRACT(YEAR FROM AGE(CURRENT_DATE, k.tanggal_lahir)) as usia, p.kategori_risiko").
        Joins("JOIN penduduk as k ON p.penduduk_id = k.id").
        Where("p.kelompok = ?", kategori).
        Where("p.deleted_at IS NULL").
        Order("p.penduduk_id, p.tanggal_pemeriksaan DESC") // urutkan tanggal terbaru dulu

    // Filter risiko berdasarkan nilai yang ada di database (case-sensitive, sesuai data)
    if risiko != "" {
        var dbRisiko string
        switch risiko {
        case "Tinggi":
            dbRisiko = "Tinggi"   // sesuai dengan nilai di database
        case "Sedang":
            dbRisiko = "Sedang"
        case "Normal":
            dbRisiko = "Normal"
        default:
            dbRisiko = risiko
        }
        query = query.Where("p.kategori_risiko = ?", dbRisiko)
    }

    // Filter desa jika bukan superadmin
    hasFullAccess := role == "superadmin"
    if !hasFullAccess && desaID != nil {
        query = query.Where("k.desa_id = ?", *desaID)
    }

    type row struct {
        PendudukID     int32
        ID             int32
        NamaLengkap    string
        Nik            string
        Dusun          string
        Usia           int
        KategoriRisiko string
    }
    var rows []row
    if err := query.Scan(&rows).Error; err != nil {
        return nil, err
    }

    result := make([]models.PendudukRiskResponse, len(rows))
    for i, r := range rows {
        // Normalisasi risiko untuk frontend (sudah sesuai, tapi pastikan)
        normalizedRisk := r.KategoriRisiko // bisa langsung pakai karena sudah "Tinggi", "Sedang", "Normal"
        result[i] = models.PendudukRiskResponse{
            ID:          r.ID,
            NIK:         r.Nik,
            NamaLengkap: r.NamaLengkap,
            Dusun:       r.Dusun,
            Usia:        r.Usia,
            Risiko:      normalizedRisk,
        }
    }
    return result, nil
}

func (r *pemeriksaanRepository) GetRiwayatByPendudukID(ctx context.Context, pendudukID uint) ([]models.Pemeriksaan, error) {
    var riwayat []models.Pemeriksaan
    err := r.db.WithContext(ctx).
        Where("penduduk_id = ?", pendudukID).
        Order("tanggal_pemeriksaan DESC").
        Find(&riwayat).Error
    return riwayat, err
}

// repositories/pemeriksaan_repository.go
func (r *pemeriksaanRepository) GetLatestPemeriksaanByPenduduk(ctx context.Context, pendudukID uint, kelompok string) (*models.Pemeriksaan, error) {
    var pemeriksaan models.Pemeriksaan
    err := r.db.WithContext(ctx).
        Where("penduduk_id = ? AND kelompok = ?", pendudukID, kelompok).
        Order("tanggal_pemeriksaan DESC").
        First(&pemeriksaan).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &pemeriksaan, err
}