package usecases

import (
    "errors"
    "time"
    "monitoring-service/app/middlewares"
    "monitoring-service/app/models"
    "monitoring-service/app/repositories"
    "monitoring-service/app/utils"
)

type PencatatanUsecase interface {
    GetPendudukByKategori(kategori string, desaID *int32, role string) ([]models.PendudukWithPemeriksaan, error)
    ValidasiUmurKategori(pendudukID int32, kategori string) (bool, error)
    TambahPemeriksaanAnak(req *PemeriksaanAnakRequest, desaID *int32, role string) (*models.PemeriksaanAnak, error)
    TambahPemeriksaanRemaja(req *PemeriksaanRemajaRequest, desaID *int32, role string) (*models.PemeriksaanRemaja, error)
    TambahPemeriksaanDewasa(req *PemeriksaanDewasaRequest, desaID *int32, role string) (*models.PemeriksaanDewasa, error)
    TambahPemeriksaanLansia(req *PemeriksaanLansiaRequest, desaID *int32, role string) (*models.PemeriksaanLansia, error)
    GetRiwayatPemeriksaanByPendudukID(pendudukID int32, kategori string, desaID *int32, role string) (interface{}, error)
}

type pencatatanUsecase struct {
    pendudukRepo *repositories.KependudukanRepository
    anakRepo     repositories.PemeriksaanAnakRepository
    remajaRepo   repositories.PemeriksaanRemajaRepository
    dewasaRepo   repositories.PemeriksaanDewasaRepository
    lansiaRepo   repositories.PemeriksaanLansiaRepository
}

func NewPencatatanUsecase(
    pendudukRepo *repositories.KependudukanRepository,
    anakRepo repositories.PemeriksaanAnakRepository,
    remajaRepo repositories.PemeriksaanRemajaRepository,
    dewasaRepo repositories.PemeriksaanDewasaRepository,
    lansiaRepo repositories.PemeriksaanLansiaRepository,
) PencatatanUsecase {
    return &pencatatanUsecase{
        pendudukRepo: pendudukRepo,
        anakRepo:     anakRepo,
        remajaRepo:   remajaRepo,
        dewasaRepo:   dewasaRepo,
        lansiaRepo:   lansiaRepo,
    }
}

// helper hitung umur pada tanggal tertentu (karena utils.HitungUmur hanya untuk sekarang)
func hitungUmurPadaTanggal(tglLahir, tglReferensi time.Time) int {
    years := tglReferensi.Year() - tglLahir.Year()
    if tglReferensi.YearDay() < tglLahir.YearDay() {
        years--
    }
    return years
}

func (u *pencatatanUsecase) GetPendudukByKategori(kategori string, desaID *int32, role string) ([]models.PendudukWithPemeriksaan, error) {
    var minAge, maxAge int
    switch kategori {
    case "anak":
        minAge, maxAge = 6, 9
    case "remaja":
        minAge, maxAge = 10, 18
    case "dewasa":
        minAge, maxAge = 19, 59
    case "lansia":
        minAge, maxAge = 60, 150
    default:
        return nil, errors.New("kategori tidak valid")
    }

    pendudukList, err := u.pendudukRepo.FindByAgeRange(minAge, maxAge, desaID, role)
    if err != nil {
        return nil, err
    }

    result := []models.PendudukWithPemeriksaan{}
    for _, p := range pendudukList {
        umur := utils.HitungUmur(p.TanggalLahir)

        var pemeriksaanTerakhir interface{}
        switch kategori {
        case "anak":
            exam, _ := u.anakRepo.GetLatestByPendudukID(p.IDKependudukan)
            if exam != nil {
                pemeriksaanTerakhir = exam
            }
        case "remaja":
            exam, _ := u.remajaRepo.GetLatestByPendudukID(p.IDKependudukan)
            if exam != nil {
                pemeriksaanTerakhir = exam
            }
        case "dewasa":
            exam, _ := u.dewasaRepo.GetLatestByPendudukID(p.IDKependudukan)
            if exam != nil {
                pemeriksaanTerakhir = exam
            }
        case "lansia":
            exam, _ := u.lansiaRepo.GetLatestByPendudukID(p.IDKependudukan)
            if exam != nil {
                pemeriksaanTerakhir = exam
            }
        }

        masihDalamRentang := umur >= minAge && umur <= maxAge
        wrapper := models.PendudukWithPemeriksaan{
            IDKependudukan:      p.IDKependudukan,
            NIK:                 p.NIK,
            NamaLengkap:         p.NamaLengkap,
            JenisKelamin:        p.JenisKelamin,
            TanggalLahir:        p.TanggalLahir,
            TempatLahir:         p.TempatLahir,
            Dusun:               p.Dusun,
            DesaID:              p.DesaID,
            Desa:                p.Desa,
            UmurSekarang:        umur,
            PemeriksaanTerakhir: pemeriksaanTerakhir,
            DapatDitambahkan:    masihDalamRentang,
        }
        if !masihDalamRentang {
            wrapper.AlasanTidakBisa = "Umur sudah melebihi batas kategori ini"
        }
        result = append(result, wrapper)
    }
    return result, nil
}

func (u *pencatatanUsecase) ValidasiUmurKategori(pendudukID int32, kategori string) (bool, error) {
    penduduk, err := u.pendudukRepo.FindByID(pendudukID)
    if err != nil {
        return false, err
    }
    umur := utils.HitungUmur(penduduk.TanggalLahir)
    switch kategori {
    case "anak":
        return umur >= 6 && umur <= 12, nil
    case "remaja":
        return umur >= 13 && umur <= 18, nil
    case "dewasa":
        return umur >= 19 && umur <= 59, nil
    case "lansia":
        return umur >= 60, nil
    default:
        return false, nil
    }
}

// ---------- Request DTOs (tidak berubah) ----------
type PemeriksaanAnakRequest struct {
    PendudukID         int32
    TanggalPemeriksaan time.Time
    BeratBadan         *float64
    TinggiBadan        *float64
    IMT                *float64
    StatusGizi         string
    KategoriRisiko     string
    StatusPemantauan   string
    RiwayatPenyakit    string
    CatatanKhusus      string
    PemeriksaID        *int32
}

type PemeriksaanRemajaRequest struct {
    PendudukID         int32
    TanggalPemeriksaan time.Time
    BeratBadan         *float64
    TinggiBadan        *float64
    IMT                *float64
    TekananDarah       string
    KategoriRisiko     string
    StatusPemantauan   string
    RiwayatPenyakit    string
    CatatanKhusus      string
    PemeriksaID        *int32
}

type PemeriksaanDewasaRequest struct {
    PendudukID         int32
    TanggalPemeriksaan time.Time
    BeratBadan         *float64
    TinggiBadan        *float64
    IMT                *float64
    TekananDarah       string
    GulaDarah          *float64
    Kolesterol         *float64
    KategoriRisiko     string
    StatusPemantauan   string
    RiwayatPenyakit    string
    PenyakitKronis     string
    CatatanKhusus      string
    PemeriksaID        *int32
}

type PemeriksaanLansiaRequest struct {
    PendudukID         int32
    TanggalPemeriksaan time.Time
    BeratBadan         *float64
    TinggiBadan        *float64
    IMT                *float64
    TekananDarah       string
    GulaDarah          *float64
    KategoriRisiko     string
    StatusPemantauan   string
    PenyakitKronis     string
    StatusKemandirian  string
    RiwayatJatuh       bool
    CatatanKhusus      string
    PemeriksaID        *int32
}

// ---------- Implementasi Tambah Pemeriksaan dengan validasi desa ----------
func (u *pencatatanUsecase) TambahPemeriksaanAnak(req *PemeriksaanAnakRequest, desaID *int32, role string) (*models.PemeriksaanAnak, error) {
    valid, err := u.ValidasiUmurKategori(req.PendudukID, "anak")
    if err != nil {
        return nil, err
    }
    if !valid {
        return nil, errors.New("umur penduduk tidak sesuai untuk kategori anak (6-12 tahun)")
    }

    penduduk, err := u.pendudukRepo.FindByID(req.PendudukID)
    if err != nil {
        return nil, err
    }

    // Validasi desa: jika role tidak memiliki akses penuh, pastikan penduduk berada di desa yang sama
    if !middlewares.HasFullAccess(role) && desaID != nil {
        if penduduk.DesaID == nil || *penduduk.DesaID != *desaID {
            return nil, errors.New("akses ditolak: penduduk bukan dari desa Anda")
        }
    }

    umurSaatPeriksa := hitungUmurPadaTanggal(penduduk.TanggalLahir, req.TanggalPemeriksaan)

    pemeriksaan := &models.PemeriksaanAnak{
        PendudukID:         req.PendudukID,
        TanggalPemeriksaan: req.TanggalPemeriksaan,
        Umur:               int32(umurSaatPeriksa),
        BeratBadan:         req.BeratBadan,
        TinggiBadan:        req.TinggiBadan,
        IMT:                req.IMT,
        StatusGizi:         req.StatusGizi,
        KategoriRisiko:     utils.NormalizeRisk(req.KategoriRisiko),
        StatusPemantauan:   req.StatusPemantauan,
        RiwayatPenyakit:    req.RiwayatPenyakit,
        CatatanKhusus:      req.CatatanKhusus,
        PemeriksaID:        req.PemeriksaID,
    }
    err = u.anakRepo.Create(pemeriksaan)
    return pemeriksaan, err
}

func (u *pencatatanUsecase) TambahPemeriksaanRemaja(req *PemeriksaanRemajaRequest, desaID *int32, role string) (*models.PemeriksaanRemaja, error) {
    valid, err := u.ValidasiUmurKategori(req.PendudukID, "remaja")
    if err != nil {
        return nil, err
    }
    if !valid {
        return nil, errors.New("umur penduduk tidak sesuai untuk kategori remaja (13-18 tahun)")
    }
    penduduk, err := u.pendudukRepo.FindByID(req.PendudukID)
    if err != nil {
        return nil, err
    }

    if !middlewares.HasFullAccess(role) && desaID != nil {
        if penduduk.DesaID == nil || *penduduk.DesaID != *desaID {
            return nil, errors.New("akses ditolak: penduduk bukan dari desa Anda")
        }
    }

    umurSaatPeriksa := hitungUmurPadaTanggal(penduduk.TanggalLahir, req.TanggalPemeriksaan)

    pemeriksaan := &models.PemeriksaanRemaja{
        PendudukID:         req.PendudukID,
        TanggalPemeriksaan: req.TanggalPemeriksaan,
        Umur:               int32(umurSaatPeriksa),
        BeratBadan:         req.BeratBadan,
        TinggiBadan:        req.TinggiBadan,
        IMT:                req.IMT,
        TekananDarah:       req.TekananDarah,
        KategoriRisiko:     utils.NormalizeRisk(req.KategoriRisiko),
        StatusPemantauan:   req.StatusPemantauan,
        RiwayatPenyakit:    req.RiwayatPenyakit,
        CatatanKhusus:      req.CatatanKhusus,
        PemeriksaID:        req.PemeriksaID,
    }
    err = u.remajaRepo.Create(pemeriksaan)
    return pemeriksaan, err
}

func (u *pencatatanUsecase) TambahPemeriksaanDewasa(req *PemeriksaanDewasaRequest, desaID *int32, role string) (*models.PemeriksaanDewasa, error) {
    valid, err := u.ValidasiUmurKategori(req.PendudukID, "dewasa")
    if err != nil {
        return nil, err
    }
    if !valid {
        return nil, errors.New("umur penduduk tidak sesuai untuk kategori dewasa (19-59 tahun)")
    }
    penduduk, err := u.pendudukRepo.FindByID(req.PendudukID)
    if err != nil {
        return nil, err
    }

    if !middlewares.HasFullAccess(role) && desaID != nil {
        if penduduk.DesaID == nil || *penduduk.DesaID != *desaID {
            return nil, errors.New("akses ditolak: penduduk bukan dari desa Anda")
        }
    }

    umurSaatPeriksa := hitungUmurPadaTanggal(penduduk.TanggalLahir, req.TanggalPemeriksaan)

    pemeriksaan := &models.PemeriksaanDewasa{
        PendudukID:         req.PendudukID,
        TanggalPemeriksaan: req.TanggalPemeriksaan,
        Umur:               int32(umurSaatPeriksa),
        BeratBadan:         req.BeratBadan,
        TinggiBadan:        req.TinggiBadan,
        IMT:                req.IMT,
        TekananDarah:       req.TekananDarah,
        GulaDarah:          req.GulaDarah,
        Kolesterol:         req.Kolesterol,
        KategoriRisiko:     utils.NormalizeRisk(req.KategoriRisiko),
        StatusPemantauan:   req.StatusPemantauan,
        RiwayatPenyakit:    req.RiwayatPenyakit,
        PenyakitKronis:     req.PenyakitKronis,
        CatatanKhusus:      req.CatatanKhusus,
        PemeriksaID:        req.PemeriksaID,
    }
    err = u.dewasaRepo.Create(pemeriksaan)
    return pemeriksaan, err
}

func (u *pencatatanUsecase) TambahPemeriksaanLansia(req *PemeriksaanLansiaRequest, desaID *int32, role string) (*models.PemeriksaanLansia, error) {
    valid, err := u.ValidasiUmurKategori(req.PendudukID, "lansia")
    if err != nil {
        return nil, err
    }
    if !valid {
        return nil, errors.New("umur penduduk tidak sesuai untuk kategori lansia (>=60 tahun)")
    }
    penduduk, err := u.pendudukRepo.FindByID(req.PendudukID)
    if err != nil {
        return nil, err
    }

    if !middlewares.HasFullAccess(role) && desaID != nil {
        if penduduk.DesaID == nil || *penduduk.DesaID != *desaID {
            return nil, errors.New("akses ditolak: penduduk bukan dari desa Anda")
        }
    }

    umurSaatPeriksa := hitungUmurPadaTanggal(penduduk.TanggalLahir, req.TanggalPemeriksaan)

    pemeriksaan := &models.PemeriksaanLansia{
        PendudukID:         req.PendudukID,
        TanggalPemeriksaan: req.TanggalPemeriksaan,
        Umur:               int32(umurSaatPeriksa),
        BeratBadan:         req.BeratBadan,
        TinggiBadan:        req.TinggiBadan,
        IMT:                req.IMT,
        TekananDarah:       req.TekananDarah,
        GulaDarah:          req.GulaDarah,
        KategoriRisiko:     utils.NormalizeRisk(req.KategoriRisiko),
        StatusPemantauan:   req.StatusPemantauan,
        PenyakitKronis:     req.PenyakitKronis,
        StatusKemandirian:  req.StatusKemandirian,
        RiwayatJatuh:       req.RiwayatJatuh,
        CatatanKhusus:      req.CatatanKhusus,
        PemeriksaID:        req.PemeriksaID,
    }
    err = u.lansiaRepo.Create(pemeriksaan)
    return pemeriksaan, err
}

func (u *pencatatanUsecase) GetRiwayatPemeriksaanByPendudukID(pendudukID int32, kategori string, desaID *int32, role string) (interface{}, error) {
    // Validasi desa: pastikan penduduk yang diakses berada di desa yang sesuai
    penduduk, err := u.pendudukRepo.FindByID(pendudukID)
    if err != nil {
        return nil, err
    }
    if !middlewares.HasFullAccess(role) && desaID != nil {
        if penduduk.DesaID == nil || *penduduk.DesaID != *desaID {
            return nil, errors.New("akses ditolak: data bukan untuk desa Anda")
        }
    }

    switch kategori {
    case "anak":
        return u.anakRepo.GetAllByPendudukID(pendudukID)
    case "remaja":
        return u.remajaRepo.GetAllByPendudukID(pendudukID)
    case "dewasa":
        return u.dewasaRepo.GetAllByPendudukID(pendudukID)
    case "lansia":
        return u.lansiaRepo.GetAllByPendudukID(pendudukID)
    default:
        return nil, errors.New("kategori tidak valid")
    }
}