package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"monitoring-service/app/models"
	"monitoring-service/app/repositories"
	"time"

	"github.com/diegoholiveira/jsonlogic/v3"
)

type pemeriksaanUsecase struct {
    formRepo     repositories.FormRepository
    periksaRepo  repositories.PemeriksaanRepository
}

type PemeriksaanUsecase interface {
    GetActiveForm(ctx context.Context, kelompok string) (*models.ActiveFormResponse, error)
    SavePemeriksaan(ctx context.Context, req *models.SavePemeriksaanRequest, petugasID *uint) (*models.PemeriksaanResponse, error)
    GetRiwayatPenduduk(ctx context.Context, pendudukID uint, kelompok string) ([]models.RiwayatPemeriksaanResponse, error)
    GetDetailPemeriksaan(ctx context.Context, id uint) (*models.DetailPemeriksaanResponse, error)
	CountPendudukWithExamination(kelompok string, pendudukIDs []int32) (int64, error)
	GetLatestRiskCountByPendudukIDs(kelompok string, pendudukIDs []int32) (map[string]int, error)
}
func NewPemeriksaanUsecase(formRepo repositories.FormRepository, periksaRepo repositories.PemeriksaanRepository) PemeriksaanUsecase {
    return &pemeriksaanUsecase{
        formRepo:    formRepo,
        periksaRepo: periksaRepo,
    }
}
// Helper untuk truthiness
func isTruthy(val interface{}) bool {
    switch v := val.(type) {
    case bool:
        return v
    case string:
        return v != ""
    case float64:
        return v != 0
    case []interface{}:
        return len(v) > 0
    case map[string]interface{}:
        return len(v) > 0
    default:
        return false
    }
}

type RiskResult struct {
    KategoriRisiko string
    Rekomendasi    string
}

// evaluateRisk mengevaluasi aturan risiko berdasarkan jawaban
func (u *pemeriksaanUsecase) evaluateRisk(rules []models.FormAturanRisiko, jawaban map[string]interface{}) (RiskResult, error) {
    if len(rules) == 0 {
        return RiskResult{KategoriRisiko: "Normal", Rekomendasi: "Tidak ada tindakan"}, nil
    }

    dataJSON, err := json.Marshal(jawaban)
    if err != nil {
        return RiskResult{KategoriRisiko: "Normal", Rekomendasi: "Tidak ada tindakan"}, err
    }

    for _, rule := range rules {
        logicReader := bytes.NewReader(rule.Kondisi)
        dataReader := bytes.NewReader(dataJSON)
        var resultBuffer bytes.Buffer

        err := jsonlogic.Apply(logicReader, dataReader, &resultBuffer)
        if err != nil {
            continue
        }

        var result bool
        if err := json.Unmarshal(resultBuffer.Bytes(), &result); err != nil {
            var val interface{}
            if err2 := json.Unmarshal(resultBuffer.Bytes(), &val); err2 == nil {
                result = isTruthy(val)
            } else {
                continue
            }
        }

        if result {
            return RiskResult{
                KategoriRisiko: rule.KategoriRisiko,
                Rekomendasi:    rule.Rekomendasi,
            }, nil
        }
    }

    return RiskResult{KategoriRisiko: "Normal", Rekomendasi: "Tidak ada tindakan"}, nil
}

// GetActiveForm mengambil form aktif beserta pertanyaannya
func (u *pemeriksaanUsecase) GetActiveForm(ctx context.Context, kelompok string) (*models.ActiveFormResponse, error) {
    versi, err := u.formRepo.GetActiveFormVersion(ctx, kelompok)
    if err != nil || versi == nil {
        return nil, errors.New("no active form for this kelompok")
    }

    questions, err := u.formRepo.GetQuestionsByVersion(ctx, versi.ID)
    if err != nil {
        return nil, err
    }

    qResp := make([]models.QuestionResponse, len(questions))
    for i, q := range questions {
        var opsi []string
        json.Unmarshal(q.Opsi, &opsi)
        var validasi map[string]interface{}
        json.Unmarshal(q.AturanValidasi, &validasi)

        qResp[i] = models.QuestionResponse{
            ID:             q.ID,
            FormVersiID:    q.FormVersiID,
            Key:            q.Key,
            Label:          q.Label,
            Tipe:           q.Tipe,
            Opsi:           opsi,
            Satuan:         q.Satuan,
            Wajib:          q.Wajib,
            AturanValidasi: validasi,
            Urutan:         q.Urutan,
        }
    }

    return &models.ActiveFormResponse{
        Versi: models.FormVersionResponse{
            ID:         versi.ID,
            Kelompok:   versi.Kelompok,
            Tahun:      versi.Tahun,
            Nama:       versi.Nama,
            Aktif:      versi.Aktif,
            Keterangan: versi.Keterangan,
        },
        Pertanyaan: qResp,
    }, nil
}
// SavePemeriksaan menyimpan data pemeriksaan dan mengevaluasi risiko
func (u *pemeriksaanUsecase) SavePemeriksaan(ctx context.Context, req *models.SavePemeriksaanRequest, petugasID *uint) (*models.PemeriksaanResponse, error) {
    // 1. Get active version
    versi, err := u.formRepo.GetActiveFormVersion(ctx, req.Kelompok)
    if err != nil || versi == nil {
        return nil, errors.New("no active form for kelompok " + req.Kelompok)
    }

    // 2. Parse date
    tgl, err := time.Parse("2006-01-02", req.Tanggal)
    if err != nil {
        return nil, errors.New("invalid date format, use YYYY-MM-DD")
    }

    // 3. Validate required fields and types
    questions, err := u.formRepo.GetQuestionsByVersion(ctx, versi.ID)
    if err != nil {
        return nil, err
    }

    for _, q := range questions {
        val, exists := req.Data[q.Key]
        if q.Wajib && !exists {
            return nil, fmt.Errorf("field '%s' wajib diisi", q.Label)
        }
        if exists {
            switch q.Tipe {
            case "angka":
                if _, ok := val.(float64); !ok {
                    return nil, fmt.Errorf("field '%s' harus angka", q.Label)
                }
            case "boolean":
                if _, ok := val.(bool); !ok {
                    return nil, fmt.Errorf("field '%s' harus boolean", q.Label)
                }
            case "pilihan":
                var opsiList []string
                json.Unmarshal(q.Opsi, &opsiList)
                found := false
                for _, opt := range opsiList {
                    if fmt.Sprintf("%v", val) == opt {
                        found = true
                        break
                    }
                }
                if !found {
                    return nil, fmt.Errorf("nilai '%s' tidak valid untuk field '%s'", val, q.Label)
                }
            }
        }
    }

    // 4. Hitung derived values (IMT)
    jawaban := req.Data
    if berat, ok := jawaban["berat_badan"].(float64); ok {
        if tinggi, ok := jawaban["tinggi_badan"].(float64); ok && tinggi > 0 {
            imt := berat / ((tinggi / 100) * (tinggi / 100))
            jawaban["imt"] = imt
        }
    }

    // 5. Evaluasi risiko
    riskResult := RiskResult{KategoriRisiko: "Normal", Rekomendasi: "Tidak ada tindakan"}
    rules, err := u.formRepo.GetRiskRulesByVersion(ctx, versi.ID)
    if err == nil && len(rules) > 0 {
        riskResult, err = u.evaluateRisk(rules, jawaban)
        if err != nil {
            riskResult = RiskResult{KategoriRisiko: "Normal", Rekomendasi: "Tidak ada tindakan"}
        }
    }

    // 6. Simpan ke database
    jawabanJSON, err := json.Marshal(jawaban)
    if err != nil {
        return nil, err
    }

    pemeriksaan := &models.Pemeriksaan{
        PendudukID:         req.PendudukID,
        Kelompok:           req.Kelompok,
        TanggalPemeriksaan: tgl,
        FormVersiID:        versi.ID,
        Jawaban:            jawabanJSON,
        KategoriRisiko:     riskResult.KategoriRisiko,
        Rekomendasi:        riskResult.Rekomendasi,
        PetugasID:          petugasID,
    }

    if err := u.periksaRepo.CreatePemeriksaan(ctx, pemeriksaan); err != nil {
        return nil, err
    }

    return &models.PemeriksaanResponse{
        ID:                 pemeriksaan.ID,
        PendudukID:         pemeriksaan.PendudukID,
        Kelompok:           pemeriksaan.Kelompok,
        TanggalPemeriksaan: pemeriksaan.TanggalPemeriksaan,
        FormVersiID:        pemeriksaan.FormVersiID,
        KategoriRisiko:     pemeriksaan.KategoriRisiko,
        Rekomendasi:        pemeriksaan.Rekomendasi,
        PetugasID:          pemeriksaan.PetugasID,
        CreatedAt:          pemeriksaan.CreatedAt,
    }, nil
}

// GetRiwayatPenduduk mengambil riwayat pemeriksaan
func (u *pemeriksaanUsecase) GetRiwayatPenduduk(ctx context.Context, pendudukID uint, kelompok string) ([]models.RiwayatPemeriksaanResponse, error) {
    list, err := u.periksaRepo.GetRiwayatByPenduduk(ctx, pendudukID, kelompok)
    if err != nil {
        return nil, err
    }
    res := make([]models.RiwayatPemeriksaanResponse, len(list))
    for i, p := range list {
        res[i] = models.RiwayatPemeriksaanResponse{
            ID:                 p.ID,
            TanggalPemeriksaan: p.TanggalPemeriksaan,
            Kelompok:           p.Kelompok,
            KategoriRisiko:     p.KategoriRisiko,
            Rekomendasi:        p.Rekomendasi,
        }
    }
    return res, nil
}

// GetDetailPemeriksaan mengambil detail pemeriksaan
func (u *pemeriksaanUsecase) GetDetailPemeriksaan(ctx context.Context, id uint) (*models.DetailPemeriksaanResponse, error) {
    p, err := u.periksaRepo.GetPemeriksaanByID(ctx, id)
    if err != nil || p == nil {
        return nil, errors.New("pemeriksaan not found")
    }

    penduduk, _ := u.periksaRepo.GetPendudukByID(ctx, p.PendudukID)
    var namaPenduduk string
    if penduduk != nil {
        namaPenduduk = penduduk.NamaLengkap
    }

    versi, _ := u.formRepo.GetFormVersionByID(ctx, p.FormVersiID)
    namaVersi := ""
    if versi != nil {
        namaVersi = versi.Nama
    }

    var jawabanMap map[string]interface{}
    json.Unmarshal(p.Jawaban, &jawabanMap)

    return &models.DetailPemeriksaanResponse{
        ID:                 p.ID,
        PendudukID:         p.PendudukID,
        NamaPenduduk:       namaPenduduk,
        Kelompok:           p.Kelompok,
        TanggalPemeriksaan: p.TanggalPemeriksaan,
        VersiForm:          namaVersi,
        KategoriRisiko:     p.KategoriRisiko,
        Rekomendasi:        p.Rekomendasi,
        Jawaban:            jawabanMap,
    }, nil
}


func (u *pemeriksaanUsecase) CountPendudukWithExamination(kelompok string, pendudukIDs []int32) (int64, error) {
	return u.periksaRepo.CountDistinctPendudukByKelompokAndIDs(kelompok, pendudukIDs)
}

func (u *pemeriksaanUsecase) GetLatestRiskCountByPendudukIDs(kelompok string, pendudukIDs []int32) (map[string]int, error) {
	return u.periksaRepo.GetLatestRiskCountByPendudukIDs(kelompok, pendudukIDs)
}