# Security and Observability Standards

Dokumen ini mendefinisikan baseline security/observability yang terukur untuk gate G6 dan release readiness.

## 1) Security Merge Gates (G6)

| Kontrol | Aturan Minimum (Lulus) | Evidence Wajib |
|---|---|---|
| Secret management | 0 secret findings pada secret scan CI | nama tool scan + hasil CI |
| Dependency hygiene | 0 vulnerability `high`/`critical` yang belum dimitigasi | hasil dependency scan + ticket mitigasi (jika ada medium) |
| Auth & access control | Semua endpoint protected memiliki verifikasi authorization | path middleware/handler + test authz |
| Input/output safety | Semua input eksternal divalidasi di boundary, output UI disanitasi | path validator/sanitizer + test invalid input |
| Error safety | Error response ke client tidak membocorkan detail internal | contoh response + test error mapping |

Aturan tambahan:

- Secret tidak boleh ada di repo, log, fixture, atau screenshot evidence.
- Dependency baru wajib menyertakan justifikasi dan hasil review security.

## 2) Observability Baseline (Release Gate)

### Logging (Wajib)

Log backend harus structured dan minimal memiliki field:

- `timestamp`
- `level`
- `service`
- `operation`
- `request_id` atau `trace_id`

Data sensitif (token, password, API key, PII) wajib di-redact.

### Metrics (Wajib)

Metrik minimum yang harus tersedia:

- API request rate, p95 latency, error rate (4xx/5xx),
- queue depth + failure count,
- webhook processed/failed/duplicate,
- payment checkout initiated/success/failed.

### Alerting Threshold Minimum

- API 5xx rate > 3% selama 5 menit.
- p95 latency endpoint kritikal > 800 ms selama 10 menit.
- webhook failure ratio > 2% selama 10 menit.
- queue backlog > 500 job selama 15 menit.

Threshold di atas adalah baseline. Tim boleh membuat threshold lebih ketat, bukan lebih longgar.

### Tracing

Tracing end-to-end wajib tersedia untuk flow:

- auth,
- search,
- billing checkout + webhook.

## 3) Incident Readiness

- Runbook untuk insiden utama (DB, Redis, webhook, third-party API) harus tersedia dan versioned.
- Perubahan high-risk wajib menyertakan langkah rollback + recovery yang bisa dieksekusi.
- Incident major wajib memiliki postmortem dengan action item terukur (owner + due date).
