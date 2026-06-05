# API Documentation - Lembar Pemantauan Anak

## Overview
Lembar Pemantauan Anak adalah fitur untuk mencatat pemantauan kesehatan anak berdasarkan tanda-tanda sakit yang terdeteksi. Setiap lembar pemantauan mewakili satu kali pemeriksaan dengan detail gejala yang dicek.

## Base URL
```
/tenaga-kesehatan/lembar-pemantauan
```

## Authentication
Semua endpoint memerlukan:
- JWT Token (Bearer Token)
- Role: Tenaga Kesehatan (Bidan, Dokter, Kader)

## Endpoints

### 1. Create Lembar Pemantauan
**Endpoint:** `POST /tenaga-kesehatan/lembar-pemantauan`

**Request Body:**
```json
{
  "anak_id": 1,
  "rentang_usia_id": 2,
  "periode_waktu": 4,
  "tanggal_periksa": "2024-04-29",
  "nama_pemeriksa": "Ibu Siti / Paraf Kader",
  "detail_gejala": [
    {
      "kategori_tanda_sakit_id": 1,
      "is_terjadi": true
    },
    {
      "kategori_tanda_sakit_id": 2,
      "is_terjadi": false
    },
    {
      "kategori_tanda_sakit_id": 3,
      "is_terjadi": true
    }
  ]
}
```

**Response (Success 200):**
```json
{
  "message": "Lembar pemantauan berhasil dibuat",
  "data": {
    "id": 10,
    "anak_id": 1,
    "rentang_usia_id": 2,
    "periode_waktu": 4,
    "tanggal_periksa": "2024-04-29T00:00:00Z",
    "nama_pemeriksa": "Ibu Siti / Paraf Kader",
    "created_at": "2024-04-29T14:30:00Z",
    "updated_at": "2024-04-29T14:30:00Z",
    "anak": {
      "id": 1,
      "nama_anak": "Budi Santoso",
      "jenis_kelamin": "L",
      "tanggal_lahir": "2024-01-15T00:00:00Z"
    },
    "rentang_usia": {
      "id": 2,
      "nama_rentang": "29 Hari - 3 Bulan",
      "satuan_waktu": "Hari"
    },
    "detail_gejala": [
      {
        "id": 45,
        "lembar_pemantauan_id": 10,
        "kategori_tanda_sakit_id": 1,
        "is_terjadi": true,
        "kategori_tanda_sakit": {
          "id": 1,
          "rentang_usia_id": 2,
          "gejala": "Demam",
          "deskripsi": "Suhu tubuh > 37.5°C",
          "is_active": true
        }
      },
      {
        "id": 46,
        "lembar_pemantauan_id": 10,
        "kategori_tanda_sakit_id": 2,
        "is_terjadi": false,
        "kategori_tanda_sakit": {
          "id": 2,
          "rentang_usia_id": 2,
          "gejala": "Kejang",
          "deskripsi": "Tubuh berkedut tanpa kontrol",
          "is_active": true
        }
      }
    ]
  }
}
```

---

### 2. Get Lembar Pemantauan by ID
**Endpoint:** `GET /tenaga-kesehatan/lembar-pemantauan/:id`

**URL Example:** 
```
GET /tenaga-kesehatan/lembar-pemantauan/10
```

**Response (Success 200):**
```json
{
  "data": {
    "id": 10,
    "anak_id": 1,
    "rentang_usia_id": 2,
    "periode_waktu": 4,
    "tanggal_periksa": "2024-04-29T00:00:00Z",
    "nama_pemeriksa": "Ibu Siti",
    "created_at": "2024-04-29T14:30:00Z",
    "updated_at": "2024-04-29T14:30:00Z",
    "anak": { ... },
    "rentang_usia": { ... },
    "detail_gejala": [ ... ]
  }
}
```

---

### 3. Get Lembar Pemantauan by Anak ID
**Endpoint:** `GET /tenaga-kesehatan/lembar-pemantauan?anak_id=1`

**Query Parameters:**
- `anak_id` (optional): Filter by anak_id. Jika tidak diberikan, akan mengembalikan semua

**Response (Success 200):**
```json
{
  "data": [
    {
      "id": 10,
      "anak_id": 1,
      "rentang_usia_id": 2,
      "periode_waktu": 4,
      "tanggal_periksa": "2024-04-29T00:00:00Z",
      "nama_pemeriksa": "Ibu Siti",
      "created_at": "2024-04-29T14:30:00Z",
      "updated_at": "2024-04-29T14:30:00Z",
      "anak": { ... },
      "rentang_usia": { ... },
      "detail_gejala": [ ... ]
    },
    {
      "id": 9,
      "anak_id": 1,
      "rentang_usia_id": 2,
      "periode_waktu": 3,
      "tanggal_periksa": "2024-04-28T00:00:00Z",
      "nama_pemeriksa": "Bidan Rani",
      "created_at": "2024-04-28T10:15:00Z",
      "updated_at": "2024-04-28T10:15:00Z",
      "anak": { ... },
      "rentang_usia": { ... },
      "detail_gejala": [ ... ]
    }
  ]
}
```

---

### 4. Update Lembar Pemantauan
**Endpoint:** `PUT /tenaga-kesehatan/lembar-pemantauan/:id`

**URL Example:**
```
PUT /tenaga-kesehatan/lembar-pemantauan/10
```

**Request Body:**
```json
{
  "anak_id": 1,
  "rentang_usia_id": 2,
  "periode_waktu": 4,
  "tanggal_periksa": "2024-04-29",
  "nama_pemeriksa": "Bidan Rani (Updated)",
  "detail_gejala": [
    {
      "kategori_tanda_sakit_id": 1,
      "is_terjadi": false
    },
    {
      "kategori_tanda_sakit_id": 2,
      "is_terjadi": true
    }
  ]
}
```

**Response (Success 200):**
```json
{
  "message": "Lembar pemantauan berhasil diupdate",
  "data": {
    "id": 10,
    "anak_id": 1,
    "rentang_usia_id": 2,
    "periode_waktu": 4,
    "tanggal_periksa": "2024-04-29T00:00:00Z",
    "nama_pemeriksa": "Bidan Rani (Updated)",
    "created_at": "2024-04-29T14:30:00Z",
    "updated_at": "2024-04-29T15:45:00Z",
    "anak": { ... },
    "rentang_usia": { ... },
    "detail_gejala": [
      {
        "id": 47,
        "lembar_pemantauan_id": 10,
        "kategori_tanda_sakit_id": 1,
        "is_terjadi": false,
        "kategori_tanda_sakit": { ... }
      },
      {
        "id": 48,
        "lembar_pemantauan_id": 10,
        "kategori_tanda_sakit_id": 2,
        "is_terjadi": true,
        "kategori_tanda_sakit": { ... }
      }
    ]
  }
}
```

---

### 5. Delete Lembar Pemantauan
**Endpoint:** `DELETE /tenaga-kesehatan/lembar-pemantauan/:id`

**URL Example:**
```
DELETE /tenaga-kesehatan/lembar-pemantauan/10
```

**Response (Success 200):**
```json
{
  "message": "Lembar pemantauan berhasil dihapus"
}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "anak_id harus lebih dari 0"
}
```

### 401 Unauthorized
```json
{
  "error": "token tidak valid"
}
```

### 403 Forbidden
```json
{
  "error": "role anda tidak memiliki akses ke fitur ini"
}
```

### 404 Not Found
```json
{
  "error": "lembar pemantauan tidak ditemukan"
}
```

### 500 Internal Server Error
```json
{
  "error": "gagal mengambil data lembar pemantauan"
}
```

---

## Data Types

### RentangUsia (Age Range)
- "0-28 Hari" (0-28 Days)
- "29 Hari - 3 Bulan" (29 Days - 3 Months)
- "3 - 6 Bulan" (3 - 6 Months)
- "6 - 12 Bulan" (6 - 12 Months)
- "1 - 2 Tahun" (1 - 2 Years)
- "2 - 6 Tahun" (2 - 6 Years)

### Kategori Tanda Sakit (Symptom Categories)
Contoh gejala untuk "0-28 Hari":
- Demam
- Kejang
- Tali pusat kemerahan
- Tidak menyusu dengan baik

---

## Implementation Notes

1. **Soft Delete**: Semua delete operation adalah soft delete (tidak menghapus dari database, hanya set deleted_at)
2. **Transaction**: Create & Update operation menggunakan transaction untuk ensure consistency
3. **Preloading**: Query otomatis load relasi: Anak, RentangUsia, DetailGejala dengan KategoriTandaSakit
4. **Ordering**: List query di-order by tanggal_periksa DESC (paling baru dulu)
5. **Validation**: Semua request di-validate di usecase layer sebelum database operation

---

## Example CURL Request

### Create
```bash
curl -X POST http://localhost:8000/tenaga-kesehatan/lembar-pemantauan \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "anak_id": 1,
    "rentang_usia_id": 2,
    "periode_waktu": 4,
    "tanggal_periksa": "2024-04-29",
    "nama_pemeriksa": "Ibu Siti",
    "detail_gejala": [
      {"kategori_tanda_sakit_id": 1, "is_terjadi": true},
      {"kategori_tanda_sakit_id": 2, "is_terjadi": false}
    ]
  }'
```

### Get by Anak ID
```bash
curl -X GET "http://localhost:8000/tenaga-kesehatan/lembar-pemantauan?anak_id=1" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Update
```bash
curl -X PUT http://localhost:8000/tenaga-kesehatan/lembar-pemantauan/10 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "anak_id": 1,
    "rentang_usia_id": 2,
    "periode_waktu": 4,
    "tanggal_periksa": "2024-04-29",
    "nama_pemeriksa": "Bidan Rani (Updated)",
    "detail_gejala": [
      {"kategori_tanda_sakit_id": 1, "is_terjadi": false},
      {"kategori_tanda_sakit_id": 2, "is_terjadi": true}
    ]
  }'
```

### Delete
```bash
curl -X DELETE http://localhost:8000/tenaga-kesehatan/lembar-pemantauan/10 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```
