package usecases

import (
	"context"
	"monitoring-service/app/models"
	"monitoring-service/app/repositories"
	"monitoring-service/app/utils"
	"sort"
)

type RiwayatCardUsecase interface {
	GetRiwayatCard(ctx context.Context, pendudukID int32) (*models.PendudukRiwayatCardResponse, error)
}

type riwayatCardUsecase struct {
	kependudukanRepo *repositories.KependudukanRepository
    pemeriksaanRepo  repositories.PemeriksaanRepository
}

func NewRiwayatCardUsecase(
	kependudukanRepo *repositories.KependudukanRepository,
    pemeriksaanRepo repositories.PemeriksaanRepository,
) RiwayatCardUsecase {
	return &riwayatCardUsecase{
		 kependudukanRepo: kependudukanRepo,
        pemeriksaanRepo:  pemeriksaanRepo,
	}
}
func (u *riwayatCardUsecase) GetRiwayatCard(ctx context.Context, pendudukID int32) (*models.PendudukRiwayatCardResponse, error) {
    // 1. Ambil data penduduk
    penduduk, err := u.kependudukanRepo.FindByID(pendudukID)
    if err != nil {
        return nil, err
    }
    usia := utils.HitungUmur(penduduk.TanggalLahir)

    dataDiri := models.DataDiriResponse{
        ID:           penduduk.IDKependudukan,
        NIK:          getStringValue(penduduk.NIK),
        NamaLengkap:  penduduk.NamaLengkap,
        Dusun:        penduduk.Dusun,
        TanggalLahir: penduduk.TanggalLahir,
        Usia:         usia,
        JenisKelamin: penduduk.JenisKelamin,
        Agama:        penduduk.Agama,
        Pekerjaan:    penduduk.Pekerjaan,
    }

    // 2. Ambil riwayat dari tabel pemeriksaans
    pemeriksaanList, err := u.pemeriksaanRepo.GetRiwayatByPendudukID(ctx, uint(pendudukID))
    if err != nil {
        // Jika error, tetap kembalikan data diri dengan riwayat kosong
        return &models.PendudukRiwayatCardResponse{
            DataDiri: dataDiri,
            Riwayat:  []models.RiwayatCard{},
        }, nil
    }

    // 3. Konversi ke RiwayatCard
    allRiwayat := make([]models.RiwayatCard, 0, len(pemeriksaanList))
    for _, ex := range pemeriksaanList {
        riwayatCard := models.RiwayatCard{
            ID:                 ex.ID,
            Kategori:           mapKelompokToKategori(ex.Kelompok),
            TanggalPemeriksaan: ex.TanggalPemeriksaan,
            KategoriRisiko:     utils.NormalizeRisk(ex.KategoriRisiko),
            Rekomendasi:        ex.Rekomendasi,
            // CatatanKhusus:      ex.CatatanKhusus,
            // Field lain (IMT, berat, tinggi, dll) bisa diambil dari parsing Jawaban jika diperlukan
        }
        allRiwayat = append(allRiwayat, riwayatCard)
    }

    // 4. Urutkan berdasarkan tanggal terbaru
    sort.Slice(allRiwayat, func(i, j int) bool {
        return allRiwayat[i].TanggalPemeriksaan.After(allRiwayat[j].TanggalPemeriksaan)
    })

    return &models.PendudukRiwayatCardResponse{
        DataDiri: dataDiri,
        Riwayat:  allRiwayat,
    }, nil
}

func mapKelompokToKategori(kelompok string) string {
    switch kelompok {
    case "anak":
        return "Anak"
    case "remaja":
        return "Remaja"
    case "dewasa":
        return "Dewasa"
    case "lansia":
        return "Lansia"
    default:
        return kelompok
    }
}

func getStringValue(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}