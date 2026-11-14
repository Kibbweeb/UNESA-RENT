# PostgreSQL

Setup database PostgreSQL untuk UNESARENT. Belum terhubung ke backend: backend masih
memakai penyimpanan JSON di `backend/data/users.json`.

## Isi folder

```
postgres/
├── docker-compose.yml
├── .env.example
└── initdb/
    └── 001_schema.sql
```

## Menjalankan dengan Docker

```bash
cd postgres
cp .env.example .env
docker compose up -d
docker compose ps
```

Schema di `initdb/001_schema.sql` otomatis dijalankan saat volume database pertama kali dibuat.

```bash
docker compose logs -f postgres
docker compose down
docker compose down -v
```

`down` mematikan container tetapi data tetap tersimpan. `down -v` menghapus data.

Dari root proyek juga tersedia:

```bash
make db-up
make db-down
make db-reset
```

## Menjalankan dengan PostgreSQL lokal

```bash
createdb unesarent
psql -d unesarent -f postgres/initdb/001_schema.sql
```

## Koneksi

```text
postgres://unesarent:unesarent_dev@localhost:5432/unesarent?sslmode=disable
```

Nilai tersebut tersedia sebagai `DATABASE_URL` di `postgres/.env.example`. Ganti password
sebelum dipakai di luar development.

## Skema

Tabel `users`:

| Kolom         | Tipe         | Keterangan                          |
| ------------- | ------------ | ----------------------------------- |
| id            | UUID         | Primary key, default `gen_random_uuid()` |
| name          | VARCHAR(80)  | Nama user                           |
| email         | VARCHAR(254) | Unik, case-insensitive lewat index  |
| password_hash | TEXT         | Hash bcrypt                         |
| created_at    | TIMESTAMPTZ  | Waktu dibuat                        |
| updated_at    | TIMESTAMPTZ  | Diperbarui otomatis oleh trigger    |

## Saat nanti dihubungkan ke backend

- Implementasikan interface `user.Store` di `backend/internal/user` dengan PostgreSQL.
- Baca `DATABASE_URL` dari environment backend.
- Backend saat ini memakai ID hex 32 karakter; schema ini memakai UUID sehingga perlu
  penyesuaian saat pembuatan ID atau casting.
- Jangan commit file `.env` yang berisi password asli.
