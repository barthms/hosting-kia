package usecases

import (
	"fmt"
	"log"
	"monitoring-service/app/middlewares"
	"monitoring-service/app/models"
	"monitoring-service/app/utils"
)

type DashboardUsecase interface {
	GetJumlahPerKelompokUsia(desaID *int32, role string) (*models.JumlahKelompokUsia, error)
	GetKesehatanPerKelompok(desaID *int32, role string) (models.KesehatanKelompokResponse, error)
	GetCakupanPemeriksaan(desaID *int32, role string) ([]models.CakupanPemeriksaan, error)
}

type dashboardUsecase struct {
	kependudukanUsecase KependudukanUsecase
	anakUsecase         PemeriksaanAnakUsecase
	remajaUsecase       PemeriksaanRemajaUsecase
	dewasaUsecase       PemeriksaanDewasaUsecase
	lansiaUsecase       PemeriksaanLansiaUsecase
}

func NewDashboardUsecase(
	kependudukanUsecase KependudukanUsecase,
	anakUsecase PemeriksaanAnakUsecase,
	remajaUsecase PemeriksaanRemajaUsecase,
	dewasaUsecase PemeriksaanDewasaUsecase,
	lansiaUsecase PemeriksaanLansiaUsecase,
) DashboardUsecase {
	return &dashboardUsecase{
		kependudukanUsecase: kependudukanUsecase,
		anakUsecase:         anakUsecase,
		remajaUsecase:       remajaUsecase,
		dewasaUsecase:       dewasaUsecase,
		lansiaUsecase:       lansiaUsecase,
	}
}

func safePercent(done, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(done) / float64(total) * 100
}

// getFilteredPenduduk mengembalikan daftar penduduk aktif dengan filter desa jika diperlukan
func (u *dashboardUsecase) getFilteredPenduduk(desaID *int32, role string) ([]models.Kependudukan, error) {
	if middlewares.HasFullAccess(role) {
		return u.kependudukanUsecase.GetAllActive()
	}
	if desaID != nil {
		return u.kependudukanUsecase.GetAllActiveByDesaID(*desaID)
	}
	return []models.Kependudukan{}, nil
}

func (u *dashboardUsecase) GetJumlahPerKelompokUsia(desaID *int32, role string) (*models.JumlahKelompokUsia, error) {
	penduduks, err := u.getFilteredPenduduk(desaID, role)
	if err != nil {
		return nil, err
	}

	result := &models.JumlahKelompokUsia{}
	for _, p := range penduduks {
		umur := utils.HitungUmur(p.TanggalLahir)
		switch {
		case umur <= 5:
			result.Balita++
		case umur <= 12:
			result.Anak++
		case umur <= 18:
			result.Remaja++
		case umur <= 59:
			result.Dewasa++
		default:
			result.Lansia++
		}
	}
	return result, nil
}

func (u *dashboardUsecase) GetKesehatanPerKelompok(desaID *int32, role string) (models.KesehatanKelompokResponse, error) {
	penduduks, err := u.getFilteredPenduduk(desaID, role)
	if err != nil {
		return nil, err
	}

	kelompokIDs := map[string][]int32{
		"balita": {}, "anak": {}, "remaja": {}, "dewasa": {}, "lansia": {},
	}
	for _, p := range penduduks {
		umur := utils.HitungUmur(p.TanggalLahir)
		id := p.IDKependudukan
		switch {
		case umur <= 5:
			kelompokIDs["balita"] = append(kelompokIDs["balita"], id)
		case umur <= 12:
			kelompokIDs["anak"] = append(kelompokIDs["anak"], id)
		case umur <= 18:
			kelompokIDs["remaja"] = append(kelompokIDs["remaja"], id)
		case umur <= 59:
			kelompokIDs["dewasa"] = append(kelompokIDs["dewasa"], id)
		default:
			kelompokIDs["lansia"] = append(kelompokIDs["lansia"], id)
		}
	}

	result := make(models.KesehatanKelompokResponse)
	result["balita"] = models.RiskCount{"Rendah": 0, "Sedang": 0, "Tinggi": 0}

	if anakRisk, err := u.anakUsecase.GetLatestRiskCountByPendudukIDs(kelompokIDs["anak"]); err == nil {
		result["anak"] = anakRisk
	} else {
		result["anak"] = models.RiskCount{"Rendah": 0, "Sedang": 0, "Tinggi": 0}
	}

	if remajaRisk, err := u.remajaUsecase.GetLatestRiskCountByPendudukIDs(kelompokIDs["remaja"]); err == nil {
		result["remaja"] = remajaRisk
	} else {
		result["remaja"] = models.RiskCount{"Rendah": 0, "Sedang": 0, "Tinggi": 0}
	}

	if dewasaRisk, err := u.dewasaUsecase.GetLatestRiskCountByPendudukIDs(kelompokIDs["dewasa"]); err == nil {
		result["dewasa"] = dewasaRisk
	} else {
		result["dewasa"] = models.RiskCount{"Rendah": 0, "Sedang": 0, "Tinggi": 0}
	}

	if lansiaRisk, err := u.lansiaUsecase.GetLatestRiskCountByPendudukIDs(kelompokIDs["lansia"]); err == nil {
		result["lansia"] = lansiaRisk
	} else {
		result["lansia"] = models.RiskCount{"Rendah": 0, "Sedang": 0, "Tinggi": 0}
	}

	return result, nil
}

func (u *dashboardUsecase) GetCakupanPemeriksaan(desaID *int32, role string) ([]models.CakupanPemeriksaan, error) {
	log.Println(">>> GetCakupanPemeriksaan called")
	penduduks, err := u.getFilteredPenduduk(desaID, role)
	if err != nil {
		return nil, err
	}

	kelompokIDs := map[string][]int32{
		"balita": {}, "anak": {}, "remaja": {}, "dewasa": {}, "lansia": {},
	}
	for _, p := range penduduks {
		umur := utils.HitungUmur(p.TanggalLahir)
		id := p.IDKependudukan
		switch {
		case umur <= 5:
			kelompokIDs["balita"] = append(kelompokIDs["balita"], id)
		case umur <= 12:
			kelompokIDs["anak"] = append(kelompokIDs["anak"], id)
		case umur <= 18:
			kelompokIDs["remaja"] = append(kelompokIDs["remaja"], id)
		case umur <= 59:
			kelompokIDs["dewasa"] = append(kelompokIDs["dewasa"], id)
		default:
			kelompokIDs["lansia"] = append(kelompokIDs["lansia"], id)
		}
	}

	var result []models.CakupanPemeriksaan

	totalBalita := int64(len(kelompokIDs["balita"]))
	result = append(result, models.CakupanPemeriksaan{
		Kelompok:       "balita",
		TotalSasaran:   totalBalita,
		SudahDiperiksa: 0,
		BelumDiperiksa: totalBalita,
		Persentase:     0,
	})

	totalAnak := int64(len(kelompokIDs["anak"]))
	sudahAnak, err := u.anakUsecase.CountPendudukWithExamination(kelompokIDs["anak"])
	if err != nil {
		return nil, fmt.Errorf("gagal hitung pemeriksaan anak: %w", err)
	}
	result = append(result, models.CakupanPemeriksaan{
		Kelompok:       "anak",
		TotalSasaran:   totalAnak,
		SudahDiperiksa: sudahAnak,
		BelumDiperiksa: totalAnak - sudahAnak,
		Persentase:     safePercent(sudahAnak, totalAnak),
	})

	totalRemaja := int64(len(kelompokIDs["remaja"]))
	sudahRemaja, err := u.remajaUsecase.CountPendudukWithExamination(kelompokIDs["remaja"])
	if err != nil {
		return nil, fmt.Errorf("gagal hitung pemeriksaan remaja: %w", err)
	}
	result = append(result, models.CakupanPemeriksaan{
		Kelompok:       "remaja",
		TotalSasaran:   totalRemaja,
		SudahDiperiksa: sudahRemaja,
		BelumDiperiksa: totalRemaja - sudahRemaja,
		Persentase:     safePercent(sudahRemaja, totalRemaja),
	})

	totalDewasa := int64(len(kelompokIDs["dewasa"]))
	sudahDewasa, err := u.dewasaUsecase.CountPendudukWithExamination(kelompokIDs["dewasa"])
	if err != nil {
		return nil, fmt.Errorf("gagal hitung pemeriksaan dewasa: %w", err)
	}
	result = append(result, models.CakupanPemeriksaan{
		Kelompok:       "dewasa",
		TotalSasaran:   totalDewasa,
		SudahDiperiksa: sudahDewasa,
		BelumDiperiksa: totalDewasa - sudahDewasa,
		Persentase:     safePercent(sudahDewasa, totalDewasa),
	})

	totalLansia := int64(len(kelompokIDs["lansia"]))
	sudahLansia, err := u.lansiaUsecase.CountPendudukWithExamination(kelompokIDs["lansia"])
	if err != nil {
		return nil, fmt.Errorf("gagal hitung pemeriksaan lansia: %w", err)
	}
	result = append(result, models.CakupanPemeriksaan{
		Kelompok:       "lansia",
		TotalSasaran:   totalLansia,
		SudahDiperiksa: sudahLansia,
		BelumDiperiksa: totalLansia - sudahLansia,
		Persentase:     safePercent(sudahLansia, totalLansia),
	})

	return result, nil
}