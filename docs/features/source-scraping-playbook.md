# Source Scraping Playbook (Glints, Kalibrr, JobStreet)

Dokumen ini merangkum konteks implementasi scraping source utama untuk Iteration 1.1, berdasarkan referensi teknis dan hasil percobaan internal.

## 1) Tujuan

1. Menetapkan kontrak scraping per source secara eksplisit.
2. Menangani perbedaan auth/pagination antar source tanpa merusak reliability batch.
3. Menyediakan runbook troubleshooting, terutama untuk kasus JobStreet yang membutuhkan bearer token.

## 2) Snapshot Referensi Saat Ini

| Source | Hasil referensi | Ringkasan |
|---|---|---|
| `glints` | ✅ `status=success`, `item_count=90` | Endpoint GraphQL dapat diakses dengan header yang tepat. |
| `kalibrr` | ✅ `status=success`, `item_count=45` | Endpoint REST dapat diakses tanpa bearer token. |
| `jobstreet` | ⚠️ `status=error`, `item_count=0` | Gagal karena bearer token tidak tersedia/valid. |

Catatan penting: kegagalan JobStreet pada referensi terjadi karena token tidak ditemukan dari browser aktif, bukan karena kegagalan source lain.

## 3) Profil Teknis Per Source

### 3.1 Glints

- Endpoint: GraphQL `POST` (`searchJobsV3`).
- Auth: tidak wajib bearer token pada baseline ini.
- Header penting: `Content-Type`, `Origin`, `Referer`, `x-glints-country-code`.
- Pagination: `page` + `pageSize`.
- Payload kunci: `SearchTerm`, `CountryCode`, `includeExternalJobs`, `LocationIds`.

### 3.2 Kalibrr

- Endpoint: REST `GET`.
- Auth: tidak wajib bearer token pada baseline ini.
- Query penting: `limit`, `offset`, `text`.
- Pagination: `offset = page * limit`.
- Respons utama: daftar jobs (normalisasi wajib karena field raw cukup bervariasi).

### 3.3 JobStreet

- Endpoint: GraphQL `POST` (`JobSearchV6`).
- Auth: **wajib** `Authorization: Bearer <token>`.
- Header tambahan yang dibutuhkan source: `seek-request-brand`, `seek-request-country`, `x-seek-site`.
- Pagination: `page` + `pageSize`.
- Risiko utama: token short-lived/rotasi cepat, sehingga scraping bisa gagal dengan `auth_missing` atau `auth_failed`.

## 4) Canonical Mapping ke Tabel `jobs`

Minimum field yang wajib terisi saat ingest:

- `source`
- `original_job_id`
- `title`
- `company`
- `url`
- `posted_at` (jika tersedia dari source, jika tidak `NULL` diperbolehkan)
- `raw_data` (payload mentah untuk audit/debug)

Aturan penting:

1. Deduplikasi wajib di DB: `UNIQUE(source, original_job_id)`.
2. Field opsional (salary/location/description) boleh kosong, tetapi tidak boleh membuat batch gagal total.
3. Source-specific parser harus mengembalikan format canonical yang sama sebelum proses insert.
4. Jika source tidak memberikan URL detail langsung, adapter wajib membentuk URL canonical dari slug/id; jika gagal, item di-skip dengan reason `missing_url`.

## 5) Strategi Token JobStreet

### 5.1 Environment & Boundary

- Jangan simpan token di repository, fixture, atau log.
- Jangan expose token ke response API internal.
- Masking token wajib di log (`Bearer ********`).

### 5.2 Sumber Token (prioritas)

1. `JOBSTREET_BEARER_TOKEN` dari environment/secret manager.
2. Token manual via operator runbook (untuk dev/staging).
3. Auto-discovery browser hanya untuk **lokal debugging** (non-production, best-effort).

### 5.3 Perilaku Saat Token Bermasalah

- Jika token tidak tersedia: tandai source status `auth_missing`.
- Jika source mengembalikan `401/403`: tandai `auth_failed`, invalidasi token cache, retry terbatas.
- Source lain harus tetap berjalan (partial success diterima).

## 6) Reliability Rules per Source

1. Rate limiting wajib per source (bukan global) agar satu source tidak mengorbankan source lain.
2. Retry hanya untuk error transient (`429`, `5xx`, timeout), bukan untuk error validasi payload.
3. Retries harus bounded (maksimal percobaan dan backoff jelas).
4. Partial success adalah mode default saat ada kegagalan source tunggal.

## 7) Runbook Troubleshooting Cepat

### 7.1 Gejala: JobStreet selalu `item_count=0`

Periksa berurutan:

1. apakah token tersedia di environment,
2. apakah token expired,
3. apakah header source-specific sudah terpasang,
4. apakah ada error `401/403` pada response source.

### 7.2 Gejala: Glints/Kalibrr tiba-tiba kosong

Periksa:

1. perubahan schema payload source,
2. parameter pagination yang berubah,
3. throttling/rate-limit source.

## 8) Security & Compliance Guardrails

1. Pastikan scraping mengikuti Terms of Service source dan kebijakan akses yang berlaku.
2. Jangan bypass challenge/anti-bot dengan teknik yang melanggar kebijakan source.
3. Simpan jejak operasional scraping untuk audit internal (run id, source status, error code).
4. Semua secret/token wajib lewat environment atau secret manager, bukan hardcoded.

## 9) Kaitan Implementasi

- Feature: [`./job-aggregation.md`](./job-aggregation.md)
- Architecture: [`../architecture/scraper-source-adapters.md`](../architecture/scraper-source-adapters.md)
- Flow: [`../flows/scraping-matching-flow.md`](../flows/scraping-matching-flow.md)
- Phase checklist: [`../phases/implementation-checklist.md`](../phases/implementation-checklist.md)
