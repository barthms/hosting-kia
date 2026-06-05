// usercases/penduduk_risk_usecase.go (lengkap)
package usecases

import (
	"errors"
	"monitoring-service/app/models"
	"monitoring-service/app/repositories"
	"monitoring-service/app/utils"
)

type PendudukRiskUsecase interface {
	GetPendudukByRisk(kategori, risiko string) ([]models.PendudukRiskResponse, error)
}

type pendudukRiskUsecase struct {
	anakRepo   repositories.PemeriksaanAnakRepository
	remajaRepo repositories.PemeriksaanRemajaRepository
	dewasaRepo repositories.PemeriksaanDewasaRepository
	lansiaRepo repositories.PemeriksaanLansiaRepository
}

func NewPendudukRiskUsecase(
	anakRepo repositories.PemeriksaanAnakRepository,
	remajaRepo repositories.PemeriksaanRemajaRepository,
	dewasaRepo repositories.PemeriksaanDewasaRepository,
	lansiaRepo repositories.PemeriksaanLansiaRepository,
) PendudukRiskUsecase {
	return &pendudukRiskUsecase{
		anakRepo:   anakRepo,
		remajaRepo: remajaRepo,
		dewasaRepo: dewasaRepo,
		lansiaRepo: lansiaRepo,
	}
}

func (u *pendudukRiskUsecase) GetPendudukByRisk(kategori, risiko string) ([]models.PendudukRiskResponse, error) {
	switch kategori {
	case "anak":
		results, err := u.anakRepo.GetAllLatestExamination(risiko)
		if err != nil {
			return nil, err
		}
		return u.mapToResponse(results), nil
	case "remaja":
		results, err := u.remajaRepo.GetAllLatestExamination(risiko)
		if err != nil {
			return nil, err
		}
		return u.mapToResponse(results), nil
	case "dewasa":
		results, err := u.dewasaRepo.GetAllLatestExamination(risiko)
		if err != nil {
			return nil, err
		}
		return u.mapToResponse(results), nil
	case "lansia":
		results, err := u.lansiaRepo.GetAllLatestExamination(risiko)
		if err != nil {
			return nil, err
		}
		return u.mapToResponse(results), nil
	default:
		return nil, errors.New("kategori tidak valid. Gunakan: anak, remaja, dewasa, lansia")
	}
}

// mapToResponse melakukan konversi dari slice hasil repository ke slice PendudukRiskResponse
// mapToResponse melakukan konversi dari slice hasil repository ke slice PendudukRiskResponse
func (u *pendudukRiskUsecase) mapToResponse(exams interface{}) []models.PendudukRiskResponse {
	result := []models.PendudukRiskResponse{}

	switch v := exams.(type) {
	case []models.PemeriksaanAnak:
		for _, ex := range v {
			if ex.Penduduk == nil {
				continue
			}
			usia := utils.HitungUmur(ex.Penduduk.TanggalLahir)
			nik := ""
			if ex.Penduduk.NIK != nil {
				nik = *ex.Penduduk.NIK
			}
			result = append(result, models.PendudukRiskResponse{
				ID:          ex.Penduduk.IDKependudukan,
				NIK:         nik,
				NamaLengkap: ex.Penduduk.NamaLengkap,
				Dusun:       ex.Penduduk.Dusun,
				Usia:        usia,
				Risiko:      utils.NormalizeRisk(ex.KategoriRisiko),
			})
		}
	case []models.PemeriksaanRemaja:
		for _, ex := range v {
			if ex.Penduduk == nil {
				continue
			}
			usia := utils.HitungUmur(ex.Penduduk.TanggalLahir)
			nik := ""
			if ex.Penduduk.NIK != nil {
				nik = *ex.Penduduk.NIK
			}
			result = append(result, models.PendudukRiskResponse{
				ID:          ex.Penduduk.IDKependudukan,
				NIK:         nik,
				NamaLengkap: ex.Penduduk.NamaLengkap,
				Dusun:       ex.Penduduk.Dusun,
				Usia:        usia,
				Risiko:      utils.NormalizeRisk(ex.KategoriRisiko),
			})
		}
	case []models.PemeriksaanDewasa:
		for _, ex := range v {
			if ex.Penduduk == nil {
				continue
			}
			usia := utils.HitungUmur(ex.Penduduk.TanggalLahir)
			nik := ""
			if ex.Penduduk.NIK != nil {
				nik = *ex.Penduduk.NIK
			}
			result = append(result, models.PendudukRiskResponse{
				ID:          ex.Penduduk.IDKependudukan,
				NIK:         nik,
				NamaLengkap: ex.Penduduk.NamaLengkap,
				Dusun:       ex.Penduduk.Dusun,
				Usia:        usia,
				Risiko:      utils.NormalizeRisk(ex.KategoriRisiko),
			})
		}
	case []models.PemeriksaanLansia:
		for _, ex := range v {
			if ex.Penduduk == nil {
				continue
			}
			usia := utils.HitungUmur(ex.Penduduk.TanggalLahir)
			nik := ""
			if ex.Penduduk.NIK != nil {
				nik = *ex.Penduduk.NIK
			}
			result = append(result, models.PendudukRiskResponse{
				ID:          ex.Penduduk.IDKependudukan,
				NIK:         nik,
				NamaLengkap: ex.Penduduk.NamaLengkap,
				Dusun:       ex.Penduduk.Dusun,
				Usia:        usia,
				Risiko:      utils.NormalizeRisk(ex.KategoriRisiko),
			})
		}
	default:
		// tipe tidak dikenal, kembalikan slice kosong
	}
	return result
}