# Flows Documentation Hub

Dokumen flow dipisah per proses bisnis agar mudah dibaca dan dipelihara.

- [Auth & Subscription Flow](./auth-subscription-flow.md)
- [Scraping & Matching Flow](./scraping-matching-flow.md)
- [Search Serving Flow](./search-serving-flow.md)
- [Admin Operations Flow](./admin-ops-flow.md)
- [Growth Engagement Flow](./growth-engagement-flow.md)

Dokumen ringkas lama tetap tersedia di [`system_flows.md`](./system_flows.md).

## Catatan Konvensi

- Path endpoint pada flow mengikuti format tanpa prefix base URL (contoh: `/billing/status`).
- URL final didapat dari base URL API (`/api/v1`) di [`../api/README.md`](../api/README.md).
