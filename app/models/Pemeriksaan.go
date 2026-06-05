package models

import (	
	"time"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)	
type Pemeriksaan struct {
    ID                 uint           `gorm:"primaryKey"`
    PendudukID         uint           `gorm:"not null;index"`
    Kelompok           string         `gorm:"size:20;not null;index"` // anak, remaja, dewasa, lansia
    TanggalPemeriksaan time.Time      `gorm:"not null"`
    FormVersiID        uint           `gorm:"not null;index"`
    Jawaban            datatypes.JSON `gorm:"type:jsonb;not null"`
    KategoriRisiko     string         `gorm:"size:20;not null"` // Tinggi, Sedang, Normal
    PetugasID          *uint          `gorm:"index"`
    CreatedAt          time.Time
    UpdatedAt          time.Time
    DeletedAt          gorm.DeletedAt `gorm:"index"`

    Penduduk *Kependudukan `gorm:"foreignKey:PendudukID"`
    FormVersi *FormVersi   `gorm:"foreignKey:FormVersiID"`
    Petugas   *User        `gorm:"foreignKey:PetugasID"`
}

func (Pemeriksaan) TableName() string {
    return "pemeriksaans"
}


type SavePemeriksaanRequest struct {
    PendudukID uint                   `json:"penduduk_id" validate:"required"`
    Kelompok   string                 `json:"kelompok" validate:"required,oneof=anak remaja dewasa lansia"`
    Tanggal    string                 `json:"tanggal" validate:"required"` // YYYY-MM-DD
    Data       map[string]interface{} `json:"data" validate:"required"`
}

type PemeriksaanResponse struct {
    ID                 uint      `json:"id"`
    PendudukID         uint      `json:"penduduk_id"`
    Kelompok           string    `json:"kelompok"`
    TanggalPemeriksaan time.Time `json:"tanggal_pemeriksaan"`
    FormVersiID        uint      `json:"form_versi_id"`
    KategoriRisiko     string    `json:"kategori_risiko"`
    PetugasID          *uint     `json:"petugas_id"`
    CreatedAt          time.Time `json:"created_at"`
}

type RiskRuleResponse struct {
    ID              uint                   `json:"id"`
    FormVersiID     uint                   `json:"form_versi_id"`
    NamaAturan      string                 `json:"nama_aturan"`
    Kondisi         map[string]interface{} `json:"kondisi"`
    KategoriRisiko  string                 `json:"kategori_risiko"`
    Prioritas       int                    `json:"prioritas"`
}

type ActiveFormResponse struct {
    Versi      FormVersionResponse   `json:"versi"`
    Pertanyaan []QuestionResponse    `json:"pertanyaan"`
}

type RiwayatPemeriksaanResponse struct {
    ID                 uint      `json:"id"`
    TanggalPemeriksaan time.Time `json:"tanggal_pemeriksaan"`
    Kelompok           string    `json:"kelompok"`
    KategoriRisiko     string    `json:"kategori_risiko"`
}

type DetailPemeriksaanResponse struct {
    ID                 uint                   `json:"id"`
    PendudukID         uint                   `json:"penduduk_id"`
    NamaPenduduk       string                 `json:"nama_penduduk"`
    Kelompok           string                 `json:"kelompok"`
    TanggalPemeriksaan time.Time              `json:"tanggal_pemeriksaan"`
    VersiForm          string                 `json:"versi_form"`
    KategoriRisiko     string                 `json:"kategori_risiko"`
    Jawaban            map[string]interface{} `json:"jawaban"`
    PetugasNama        string                 `json:"petugas_nama,omitempty"`
}