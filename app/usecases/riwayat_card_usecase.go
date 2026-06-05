package usecases

import (
	"monitoring-service/app/models"
	"monitoring-service/app/repositories"
	"monitoring-service/app/utils"
	"sort"
)

type RiwayatCardUsecase interface {
	GetRiwayatCard(pendudukID int32) (*models.PendudukRiwayatCardResponse, error)
}

type riwayatCardUsecase struct {
	kependudukanRepo *repositories.KependudukanRepository
	anakRepo         repositories.PemeriksaanAnakRepository
	remajaRepo       repositories.PemeriksaanRemajaRepository
	dewasaRepo       repositories.PemeriksaanDewasaRepository
	lansiaRepo       repositories.PemeriksaanLansiaRepository
}

func NewRiwayatCardUsecase(
	kependudukanRepo *repositories.KependudukanRepository,
	anakRepo repositories.PemeriksaanAnakRepository,
	remajaRepo repositories.PemeriksaanRemajaRepository,
	dewasaRepo repositories.PemeriksaanDewasaRepository,
	lansiaRepo repositories.PemeriksaanLansiaRepository,
) RiwayatCardUsecase {
	return &riwayatCardUsecase{
		kependudukanRepo: kependudukanRepo,
		anakRepo:         anakRepo,
		remajaRepo:       remajaRepo,
		dewasaRepo:       dewasaRepo,
		lansiaRepo:       lansiaRepo,
	}
}

// Helper untuk mengkonversi *float64 ke float64 (0 jika nil)
func float64Value(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func (u *riwayatCardUsecase) GetRiwayatCard(pendudukID int32) (*models.PendudukRiwayatCardResponse, error) {
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

	var allRiwayat []models.RiwayatCard

	// 2. Anak
	anakList, _ := u.anakRepo.GetAllByPendudukID(pendudukID)
	for _, ex := range anakList {
		allRiwayat = append(allRiwayat, models.RiwayatCard{
			ID:                 ex.ID,
			Kategori:           "Anak",
			TanggalPemeriksaan: ex.TanggalPemeriksaan,
			IMT:                ex.IMT,
			StatusGizi:         ex.StatusGizi,
			KategoriRisiko:     utils.NormalizeRisk(ex.KategoriRisiko),
			CatatanKhusus:      ex.CatatanKhusus,
			BeratBadan:         ex.BeratBadan,
			TinggiBadan:        ex.TinggiBadan,
		})
	}

	// 3. Remaja
	remajaList, _ := u.remajaRepo.GetAllByPendudukID(pendudukID)
	for _, ex := range remajaList {
		allRiwayat = append(allRiwayat, models.RiwayatCard{
			ID:                 ex.ID,
			Kategori:           "Remaja",
			TanggalPemeriksaan: ex.TanggalPemeriksaan,
			IMT:                ex.IMT,
			// StatusGizi:         ex.StatusGizi,
			KategoriRisiko:     utils.NormalizeRisk(ex.KategoriRisiko),
			CatatanKhusus:      ex.CatatanKhusus,
			BeratBadan:         ex.BeratBadan,
			TinggiBadan:        ex.TinggiBadan,
			TekananDarah:       ex.TekananDarah,
			// GulaDarah:          float64Value(ex.GulaDarah),
		})
	}

	// 4. Dewasa
	dewasaList, _ := u.dewasaRepo.GetAllByPendudukID(pendudukID)
	for _, ex := range dewasaList {
		allRiwayat = append(allRiwayat, models.RiwayatCard{
			ID:                 ex.ID,
			Kategori:           "Dewasa",
			TanggalPemeriksaan: ex.TanggalPemeriksaan,
			IMT:                ex.IMT,
			// StatusGizi:         ex.StatusGizi,
			KategoriRisiko:     utils.NormalizeRisk(ex.KategoriRisiko),
			CatatanKhusus:      ex.CatatanKhusus,
			BeratBadan:         ex.BeratBadan,
			TinggiBadan:        ex.TinggiBadan,
			TekananDarah:       ex.TekananDarah,
			GulaDarah:          float64Value(ex.GulaDarah),
			Kolesterol:         float64Value(ex.Kolesterol),
		})
	}

	// 5. Lansia
	lansiaList, _ := u.lansiaRepo.GetAllByPendudukID(pendudukID)
	for _, ex := range lansiaList {
		allRiwayat = append(allRiwayat, models.RiwayatCard{
			ID:                 ex.ID,
			Kategori:           "Lansia",
			TanggalPemeriksaan: ex.TanggalPemeriksaan,
			IMT:                ex.IMT,
			// StatusGizi:         ex.StatusGizi,
			KategoriRisiko:     utils.NormalizeRisk(ex.KategoriRisiko),
			CatatanKhusus:      ex.CatatanKhusus,
			BeratBadan:         ex.BeratBadan,
			TinggiBadan:        ex.TinggiBadan,
			TekananDarah:       ex.TekananDarah,
			GulaDarah:          float64Value(ex.GulaDarah),
			// Kolesterol:         float64Value(ex.Kolesterol),
			PenyakitKronis:     ex.PenyakitKronis,
			StatusKemandirian:  ex.StatusKemandirian,
		})
	}

	// 6. Urutkan berdasarkan tanggal terbaru
	sort.Slice(allRiwayat, func(i, j int) bool {
		return allRiwayat[i].TanggalPemeriksaan.After(allRiwayat[j].TanggalPemeriksaan)
	})

	return &models.PendudukRiwayatCardResponse{
		DataDiri: dataDiri,
		Riwayat:  allRiwayat,
	}, nil
}

// helper untuk pointer string
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}