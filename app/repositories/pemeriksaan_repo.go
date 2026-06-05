package repositories

// import (
//     "context"
//     "monitoring-service/app/models"
//     "gorm.io/gorm"
// )

// type PemeriksaanRepository interface {
//     Create(ctx context.Context, p *models.Pemeriksaan) error
//     GetByID(ctx context.Context, id uint) (*models.Pemeriksaan, error)
//     GetRiwayatByPenduduk(ctx context.Context, pendudukID uint, kelompok string) ([]models.Pemeriksaan, error)
//     GetLatestByPenduduk(ctx context.Context, pendudukID uint, kelompok string) (*models.Pemeriksaan, error)
// }

// type pemeriksaanRepository struct {
//     db *gorm.DB
// }

// func NewPemeriksaanRepository(db *gorm.DB) PemeriksaanRepository {
//     return &pemeriksaanRepository{db: db}
// }

// func (r *pemeriksaanRepository) Create(ctx context.Context, p *models.Pemeriksaan) error {
//     return r.db.WithContext(ctx).Create(p).Error
// }

// func (r *pemeriksaanRepository) GetByID(ctx context.Context, id uint) (*models.Pemeriksaan, error) {
//     var p models.Pemeriksaan
//     err := r.db.WithContext(ctx).Preload("Penduduk").Preload("FormVersi").First(&p, id).Error
//     if err != nil {
//         return nil, err
//     }
//     return &p, nil
// }

// func (r *pemeriksaanRepository) GetRiwayatByPenduduk(ctx context.Context, pendudukID uint, kelompok string) ([]models.Pemeriksaan, error) {
//     var list []models.Pemeriksaan
//     query := r.db.WithContext(ctx).Where("penduduk_id = ?", pendudukID)
//     if kelompok != "" {
//         query = query.Where("kelompok = ?", kelompok)
//     }
//     err := query.Order("tanggal_pemeriksaan DESC").Find(&list).Error
//     return list, err
// }

// func (r *pemeriksaanRepository) GetLatestByPenduduk(ctx context.Context, pendudukID uint, kelompok string) (*models.Pemeriksaan, error) {
//     var p models.Pemeriksaan
//     err := r.db.WithContext(ctx).Where("penduduk_id = ? AND kelompok = ?", pendudukID, kelompok).
//         Order("tanggal_pemeriksaan DESC, id DESC").
//         First(&p).Error
//     if err == gorm.ErrRecordNotFound {
//         return nil, nil
//     }
//     return &p, err
// }