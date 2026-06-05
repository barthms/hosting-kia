package usecases

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "monitoring-service/app/models"
    "monitoring-service/app/repositories"
    "strings"
    "time"

    "github.com/diegoholiveira/jsonlogic/v3"
)

// ----------------------------------------------------------------------------
// FormUsecase interface dan implementasi (tidak berubah, hanya disalin)
// ----------------------------------------------------------------------------

type FormUsecase interface {
    // Form Versi
    CreateFormVersion(ctx context.Context, req *models.CreateFormVersionRequest) (*models.FormVersionResponse, error)
    ActivateFormVersion(ctx context.Context, id uint) error
    DeactivateFormVersion(ctx context.Context, id uint) error
    DuplicateFormVersion(ctx context.Context, sourceID uint, req *models.DuplicateFormVersionRequest) (*models.FormVersionResponse, error)
    GetFormVersionsByKelompok(ctx context.Context, kelompok string) ([]models.FormVersionResponse, error)
    GetVersionDetail(ctx context.Context, id uint) (*models.FormVersionResponse, []models.QuestionResponse, []models.RiskRuleResponse, error)

    // Pertanyaan
    AddQuestion(ctx context.Context, versiID uint, req *models.AddQuestionRequest) (*models.QuestionResponse, error)
    UpdateQuestion(ctx context.Context, id uint, req *models.UpdateQuestionRequest) error
    DeleteQuestion(ctx context.Context, id uint) error

    // Aturan Risiko
    AddRiskRule(ctx context.Context, versiID uint, req *models.AddRiskRuleRequest) (*models.RiskRuleResponse, error)
    UpdateRiskRule(ctx context.Context, id uint, req *models.UpdateRiskRuleRequest) error
    DeleteRiskRule(ctx context.Context, id uint) error
}

type PemeriksaanUsecase interface {
    GetActiveForm(ctx context.Context, kelompok string) (*models.ActiveFormResponse, error)
    SavePemeriksaan(ctx context.Context, req *models.SavePemeriksaanRequest, petugasID *uint) (*models.PemeriksaanResponse, error)
    GetRiwayatPenduduk(ctx context.Context, pendudukID uint, kelompok string) ([]models.RiwayatPemeriksaanResponse, error)
    GetDetailPemeriksaan(ctx context.Context, id uint) (*models.DetailPemeriksaanResponse, error)
}

type formUsecase struct {
    repo repositories.FormRepository
}

func NewFormUsecase(repo repositories.FormRepository) FormUsecase {
    return &formUsecase{repo: repo}
}

func (u *formUsecase) CreateFormVersion(ctx context.Context, req *models.CreateFormVersionRequest) (*models.FormVersionResponse, error) {
    versi := &models.FormVersi{
        Kelompok:   req.Kelompok,
        Tahun:      req.Tahun,
        Nama:       req.Nama,
        Keterangan: req.Keterangan,
        Aktif:      false,
    }
    if err := u.repo.CreateFormVersion(ctx, versi); err != nil {
        return nil, err
    }
    return &models.FormVersionResponse{
        ID:         versi.ID,
        Kelompok:   versi.Kelompok,
        Tahun:      versi.Tahun,
        Nama:       versi.Nama,
        Aktif:      versi.Aktif,
        Keterangan: versi.Keterangan,
    }, nil
}

func (u *formUsecase) ActivateFormVersion(ctx context.Context, id uint) error {
    versi, err := u.repo.GetFormVersionByID(ctx, id)
    if err != nil || versi == nil {
        return errors.New("version not found")
    }
    if err := u.repo.DeactivateAllVersionsInKelompok(ctx, versi.Kelompok); err != nil {
        return err
    }
    versi.Aktif = true
    return u.repo.UpdateFormVersion(ctx, versi)
}

func (u *formUsecase) DeactivateFormVersion(ctx context.Context, id uint) error {
    versi, err := u.repo.GetFormVersionByID(ctx, id)
    if err != nil || versi == nil {
        return errors.New("version not found")
    }
    active, _ := u.repo.GetActiveFormVersion(ctx, versi.Kelompok)
    if active != nil && active.ID == versi.ID {
        return errors.New("cannot deactivate the only active version")
    }
    versi.Aktif = false
    return u.repo.UpdateFormVersion(ctx, versi)
}

func (u *formUsecase) DuplicateFormVersion(ctx context.Context, sourceID uint, req *models.DuplicateFormVersionRequest) (*models.FormVersionResponse, error) {
    source, err := u.repo.GetFormVersionByID(ctx, sourceID)
    if err != nil || source == nil {
        return nil, errors.New("source version not found")
    }
    newVersi := &models.FormVersi{
        Kelompok:   source.Kelompok,
        Tahun:      req.TahunBaru,
        Nama:       req.NamaBaru,
        Keterangan: req.Keterangan,
        Aktif:      false,
    }
    if err := u.repo.CreateFormVersion(ctx, newVersi); err != nil {
        return nil, err
    }
    // Copy questions
    questions, _ := u.repo.GetQuestionsByVersion(ctx, sourceID)
    for _, q := range questions {
        newQ := models.FormPertanyaan{
            FormVersiID:    newVersi.ID,
            Key:            q.Key,
            Label:          q.Label,
            Tipe:           q.Tipe,
            Opsi:           q.Opsi,
            Satuan:         q.Satuan,
            Wajib:          q.Wajib,
            AturanValidasi: q.AturanValidasi,
            Urutan:         q.Urutan,
        }
        if err := u.repo.CreateQuestion(ctx, &newQ); err != nil {
            return nil, err
        }
    }
    // Copy risk rules
    rules, _ := u.repo.GetRiskRulesByVersion(ctx, sourceID)
    for _, r := range rules {
        newRule := models.FormAturanRisiko{
            FormVersiID:    newVersi.ID,
            NamaAturan:     r.NamaAturan,
            Kondisi:        r.Kondisi,
            KategoriRisiko: r.KategoriRisiko,
            Prioritas:      r.Prioritas,
        }
        if err := u.repo.CreateRiskRule(ctx, &newRule); err != nil {
            return nil, err
        }
    }
    return &models.FormVersionResponse{
        ID:         newVersi.ID,
        Kelompok:   newVersi.Kelompok,
        Tahun:      newVersi.Tahun,
        Nama:       newVersi.Nama,
        Aktif:      newVersi.Aktif,
        Keterangan: newVersi.Keterangan,
    }, nil
}

func (u *formUsecase) GetFormVersionsByKelompok(ctx context.Context, kelompok string) ([]models.FormVersionResponse, error) {
    list, err := u.repo.GetFormVersionsByKelompok(ctx, kelompok)
    if err != nil {
        return nil, err
    }
    res := make([]models.FormVersionResponse, len(list))
    for i, v := range list {
        res[i] = models.FormVersionResponse{
            ID:         v.ID,
            Kelompok:   v.Kelompok,
            Tahun:      v.Tahun,
            Nama:       v.Nama,
            Aktif:      v.Aktif,
            Keterangan: v.Keterangan,
        }
    }
    return res, nil
}

func (u *formUsecase) GetVersionDetail(ctx context.Context, id uint) (*models.FormVersionResponse, []models.QuestionResponse, []models.RiskRuleResponse, error) {
    versi, err := u.repo.GetFormVersionByID(ctx, id)
    if err != nil || versi == nil {
        return nil, nil, nil, errors.New("version not found")
    }
    questions, _ := u.repo.GetQuestionsByVersion(ctx, id)
    rules, _ := u.repo.GetRiskRulesByVersion(ctx, id)
    versiResp := &models.FormVersionResponse{
        ID:         versi.ID,
        Kelompok:   versi.Kelompok,
        Tahun:      versi.Tahun,
        Nama:       versi.Nama,
        Aktif:      versi.Aktif,
        Keterangan: versi.Keterangan,
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
    rResp := make([]models.RiskRuleResponse, len(rules))
    for i, r := range rules {
        var kondisi map[string]interface{}
        json.Unmarshal(r.Kondisi, &kondisi)
        rResp[i] = models.RiskRuleResponse{
            ID:              r.ID,
            FormVersiID:     r.FormVersiID,
            NamaAturan:      r.NamaAturan,
            Kondisi:         kondisi,
            KategoriRisiko:  r.KategoriRisiko,
            Prioritas:       r.Prioritas,
        }
    }
    return versiResp, qResp, rResp, nil
}

func (u *formUsecase) AddQuestion(ctx context.Context, versiID uint, req *models.AddQuestionRequest) (*models.QuestionResponse, error) {
    exists, err := u.repo.CheckQuestionKeyExists(ctx, versiID, req.Key)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, fmt.Errorf("question key '%s' already exists", req.Key)
    }
    opsiJSON, _ := json.Marshal(req.Opsi)
    validasiJSON, _ := json.Marshal(req.Validasi)
    q := &models.FormPertanyaan{
        FormVersiID:    versiID,
        Key:            req.Key,
        Label:          req.Label,
        Tipe:           req.Tipe,
        Opsi:           opsiJSON,
        Satuan:         req.Satuan,
        Wajib:          req.Wajib,
        AturanValidasi: validasiJSON,
        Urutan:         req.Urutan,
    }
    if err := u.repo.CreateQuestion(ctx, q); err != nil {
        return nil, err
    }
    return &models.QuestionResponse{
        ID:             q.ID,
        FormVersiID:    q.FormVersiID,
        Key:            q.Key,
        Label:          q.Label,
        Tipe:           q.Tipe,
        Opsi:           req.Opsi,
        Satuan:         q.Satuan,
        Wajib:          q.Wajib,
        AturanValidasi: req.Validasi,
        Urutan:         q.Urutan,
    }, nil
}

func (u *formUsecase) UpdateQuestion(ctx context.Context, id uint, req *models.UpdateQuestionRequest) error {
    q, err := u.repo.GetQuestionByID(ctx, id)
    if err != nil || q == nil {
        return errors.New("question not found")
    }
    if req.Label != "" {
        q.Label = req.Label
    }
    if req.Tipe != "" {
        q.Tipe = req.Tipe
    }
    if req.Opsi != nil {
        opsiJSON, _ := json.Marshal(req.Opsi)
        q.Opsi = opsiJSON
    }
    if req.Satuan != "" {
        q.Satuan = req.Satuan
    }
    if req.Wajib != nil {
        q.Wajib = *req.Wajib
    }
    if req.Validasi != nil {
        valJSON, _ := json.Marshal(req.Validasi)
        q.AturanValidasi = valJSON
    }
    if req.Urutan != 0 {
        q.Urutan = req.Urutan
    }
    return u.repo.UpdateQuestion(ctx, q)
}

func (u *formUsecase) DeleteQuestion(ctx context.Context, id uint) error {
    return u.repo.DeleteQuestion(ctx, id)
}

func (u *formUsecase) AddRiskRule(ctx context.Context, versiID uint, req *models.AddRiskRuleRequest) (*models.RiskRuleResponse, error) {
    kondisiJSON, _ := json.Marshal(req.Kondisi)
    rule := &models.FormAturanRisiko{
        FormVersiID:    versiID,
        NamaAturan:     req.NamaAturan,
        Kondisi:        kondisiJSON,
        KategoriRisiko: req.KategoriRisiko,
        Prioritas:      req.Prioritas,
    }
    if err := u.repo.CreateRiskRule(ctx, rule); err != nil {
        return nil, err
    }
    return &models.RiskRuleResponse{
        ID:              rule.ID,
        FormVersiID:     rule.FormVersiID,
        NamaAturan:      rule.NamaAturan,
        Kondisi:         req.Kondisi,
        KategoriRisiko:  rule.KategoriRisiko,
        Prioritas:       rule.Prioritas,
    }, nil
}

func (u *formUsecase) UpdateRiskRule(ctx context.Context, id uint, req *models.UpdateRiskRuleRequest) error {
    rule, err := u.repo.GetRiskRuleByID(ctx, id)
    if err != nil || rule == nil {
        return errors.New("rule not found")
    }
    if req.NamaAturan != "" {
        rule.NamaAturan = req.NamaAturan
    }
    if req.Kondisi != nil {
        kondisiJSON, _ := json.Marshal(req.Kondisi)
        rule.Kondisi = kondisiJSON
    }
    if req.KategoriRisiko != "" {
        rule.KategoriRisiko = req.KategoriRisiko
    }
    if req.Prioritas != 0 {
        rule.Prioritas = req.Prioritas
    }
    return u.repo.UpdateRiskRule(ctx, rule)
}

func (u *formUsecase) DeleteRiskRule(ctx context.Context, id uint) error {
    return u.repo.DeleteRiskRule(ctx, id)
}

// ----------------------------------------------------------------------------
// PemeriksaanUsecase (implementasi yang diperbaiki)
// ----------------------------------------------------------------------------

type pemeriksaanUsecase struct {
    formRepo     repositories.FormRepository
    periksaRepo  repositories.PemeriksaanRepository
}

func NewPemeriksaanUsecase(formRepo repositories.FormRepository, periksaRepo repositories.PemeriksaanRepository) PemeriksaanUsecase {
    return &pemeriksaanUsecase{
        formRepo:    formRepo,
        periksaRepo: periksaRepo,
    }
}

// evaluateRisk mengevaluasi aturan JSON logic terhadap data jawaban.
// Mengembalikan kategori risiko pertama yang cocok, atau "Normal" jika tidak ada.
func (u *pemeriksaanUsecase) evaluateRisk(rules []models.FormAturanRisiko, jawaban map[string]interface{}) (string, error) {
    // Marshal data jawaban ke JSON sekali untuk digunakan berulang
    dataJSON, err := json.Marshal(jawaban)
    if err != nil {
        return "Normal", err
    }
    dataReader := strings.NewReader(string(dataJSON))

    for _, rule := range rules {
        // Marshal kondisi rule ke JSON
        kondisiJSON, err := json.Marshal(rule.Kondisi)
        if err != nil {
            continue // skip aturan yang tidak valid
        }
        logicReader := strings.NewReader(string(kondisiJSON))

        // Buffer untuk hasil evaluasi
        var resultBuffer bytes.Buffer

        // Apply JSON logic (menggunakan io.Reader/Writer)
        err = jsonlogic.Apply(logicReader, dataReader, &resultBuffer)
        if err != nil {
            continue // gagal evaluasi, lanjut ke aturan berikutnya
        }

        // Hasil evaluasi adalah JSON boolean, parse
        var result bool
        if err := json.Unmarshal(resultBuffer.Bytes(), &result); err != nil {
            // Jika tidak bisa di-unmarshal ke boolean, coba truthiness
            var val interface{}
            if err2 := json.Unmarshal(resultBuffer.Bytes(), &val); err2 == nil {
                result = isTruthy(val)
            } else {
                continue
            }
        }

        if result {
            return rule.KategoriRisiko, nil
        }

        // Reset dataReader untuk aturan berikutnya (karena sudah terbaca habis)
        dataReader = strings.NewReader(string(dataJSON))
    }
    return "Normal", nil
}

// isTruthy helper untuk menentukan truthiness dari nilai JSON
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
    // 3. Validate required fields and types based on questions
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
    // 4. Calculate derived values (e.g., IMT)
    jawaban := req.Data
    if berat, ok := jawaban["berat_badan"].(float64); ok {
        if tinggi, ok := jawaban["tinggi_badan"].(float64); ok && tinggi > 0 {
            imt := berat / ((tinggi / 100) * (tinggi / 100))
            jawaban["imt"] = imt
        }
    }
    // 5. Evaluate risk rules (menggunakan helper)
    kategoriRisiko := "Normal"
    rules, err := u.formRepo.GetRiskRulesByVersion(ctx, versi.ID)
    if err != nil {
        // Jika gagal mengambil aturan, tetap lanjut dengan default
        kategoriRisiko = "Normal"
    } else {
        kategori, err := u.evaluateRisk(rules, jawaban)
        if err != nil {
            kategoriRisiko = "Normal"
        } else {
            kategoriRisiko = kategori
        }
    }
    // 6. Save
    jawabanJSON, _ := json.Marshal(jawaban)
    pemeriksaan := &models.Pemeriksaan{
        PendudukID:         req.PendudukID,
        Kelompok:           req.Kelompok,
        TanggalPemeriksaan: tgl,
        FormVersiID:        versi.ID,
        Jawaban:            jawabanJSON,
        KategoriRisiko:     kategoriRisiko,
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
        PetugasID:          pemeriksaan.PetugasID,
        CreatedAt:          pemeriksaan.CreatedAt,
    }, nil
}

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
        }
    }
    return res, nil
}

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
    // var petugasNama string
    // if p.PetugasID != nil {
    //     user, _ := u.periksaRepo.GetUserByID(ctx, *p.PetugasID)
    //     if user != nil {
    //         petugasNama = user.Nama
    //     }
    // }
    // Get version info
    versi, _ := u.formRepo.GetFormVersionByID(ctx, p.FormVersiID)
    namaVersi := ""
    if versi != nil {
        namaVersi = versi.Nama
    }
    // Parse jawaban
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
        Jawaban:            jawabanMap,
        // PetugasNama:        petugasNama,
    }, nil
}