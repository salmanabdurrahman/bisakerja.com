# Go Coding Standards

Standar ini mengacu ke idiom Go dan struktur `golang-standards/project-layout` yang dipakai di `apps/api`.

## 1) Struktur Paket

- `cmd/` hanya untuk bootstrap executable.
- `internal/` untuk logic aplikasi privat.
- `pkg/` hanya untuk util reusable lintas domain.
- Hindari business logic di `handler`/`main`.

## 2) Prinsip Desain

1. **Explicit over magic**: alur harus terbaca langsung.
2. **Small interfaces**: definisikan interface di sisi consumer.
3. **Dependency injection**: inject dependency via constructor.
4. **Context first**: semua operasi IO menerima `context.Context`.

## 3) Naming dan API Design

- Gunakan nama package pendek dan jelas (`billing`, `jobs`, `auth`).
- Hindari `util`, `common`, `helper` yang terlalu generik.
- Ekspor simbol hanya jika dibutuhkan lintas package.
- Function sebaiknya fokus satu tanggung jawab.

## 4) Error Handling

- Jangan swallow error.
- Wrap error dengan konteks (`fmt.Errorf(\"...: %w\", err)`).
- Gunakan sentinel error hanya untuk kasus yang stabil dan terdefinisi.
- Mapping ke HTTP error code harus konsisten (lihat `docs/api/errors.md`).

## 5) Context & Timeout

- Semua call ke DB, Redis, HTTP eksternal (termasuk Mayar) wajib pakai timeout.
- Jangan simpan `context.Context` dalam struct.
- Batalkan goroutine bila context selesai.

## 6) Logging

- Gunakan structured logging.
- Wajib ada `request_id` untuk request flow.
- Jangan log secret/token/password/PII sensitif.
- Error log wajib menyertakan service, operation, dan error class.

## 7) Concurrency

- Gunakan goroutine hanya jika ada benefit yang jelas.
- Selalu pertimbangkan leak, race condition, deadlock.
- Gunakan worker pool/channel dengan batas jelas.
- Proteksi shared state dengan mutex atau desain immutable message passing.

## 8) HTTP Handler Standards

- Handler hanya untuk parsing request + call usecase + map response.
- Validation dilakukan eksplisit sebelum masuk usecase.
- Response format wajib mengikuti envelope yang sudah ditetapkan.

## 9) Database & Repository

- Query harus jelas ownership-nya per domain.
- Gunakan transaksi DB untuk update multi-entitas kritikal (contoh billing webhook).
- Pastikan operasi idempotent untuk ingestion/webhook.

## 10) Linting Minimum

Direkomendasikan menggunakan `golangci-lint` dengan rules minimum:

- `govet`, `staticcheck`, `ineffassign`, `errcheck`
- `gocritic`, `revive` (style/readability)
- `gosec` (security static checks)

## 11) Anti-Pattern yang Harus Dihindari

- Function >100 baris tanpa alasan kuat.
- `panic` untuk alur bisnis normal.
- Global mutable state tanpa kontrol.
- Menambah dependency eksternal tanpa justifikasi.
