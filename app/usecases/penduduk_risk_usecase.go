// usercases/penduduk_risk_usecase.go (lengkap)
package usecases

import (
	"errors"
	"monitoring-service/app/models"
	"monitoring-service/app/repositories"
)

type PendudukRiskUsecase interface {
	GetPendudukByRisk(kategori, risiko string,desaID *int32, role string) ([]models.PendudukRiskResponse, error)
}

type pendudukRiskUsecase struct {
	pemeriksaanRepo repositories.PemeriksaanRepository
}

func NewPendudukRiskUsecase(pemeriksaanRepo repositories.PemeriksaanRepository) PendudukRiskUsecase {
	return &pendudukRiskUsecase{
		pemeriksaanRepo: pemeriksaanRepo,
	}
}


func (u *pendudukRiskUsecase) GetPendudukByRisk(kelompok, risiko string, desaID *int32, role string) ([]models.PendudukRiskResponse, error) {
    validKelompok := map[string]bool{"anak": true, "remaja": true, "dewasa": true, "lansia": true}
    if !validKelompok[kelompok] {
        return nil, errors.New("kelompok tidak valid. Gunakan: anak, remaja, dewasa, lansia")
    }
    if risiko != "" {
        validRisk := map[string]bool{"Tinggi": true, "Sedang": true, "Normal": true}
        if !validRisk[risiko] {
            return nil, errors.New("risiko harus Tinggi, Sedang, atau Normal")
        }
    }
    return u.pemeriksaanRepo.GetPendudukByRisk(kelompok, risiko, desaID, role)
}