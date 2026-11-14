<div align="center">

# UNESARENT PostgreSQL

Setup database PostgreSQL untuk UNESARENT, disiapkan sebagai layanan mandiri.

[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-4169E1?style=flat-square&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat-square&logo=docker&logoColor=white)](https://docs.docker.com/compose/)

[Dokumentasi utama](../README.md) · [Schema](#schema) · [Integrasi](#integrasi-backend)

</div>

---

> **Status:** database ini belum terhubung ke backend. Backend masih menggunakan
> penyimpanan JSON di `backend/data/users.json`. Folder ini menyiapkan PostgreSQL untuk
> tahap integrasi berikutnya.

---

## Ringkasan

Folder `postgres/` berisi konfigurasi PostgreSQL 17 untuk UNESARENT:

- Docker Compose dengan healthcheck dan named volume
- Schema awal tabel `users`
- Unique index email case-insensitive
- Trigger untuk memperbarui `updated_at`
- Contoh environment dan connection string

Schema dijalankan otomatis saat container dibuat pertama kali.

---

## Struktur

```text
postgres/
├── initdb/
│   └── 001_schema.sql
├── .env.example
├── docker-compose.yml
└── README.md
```

| File                   | Deskripsi                                        |
| ---------------------- | ------------------------------------------------ |
| `docker-compose.yml`   | Service PostgreSQL, port, volume, dan healthcheck |
| `.env.example`         | Kredensial dan `DATABASE_URL`                    |
| `initdb/001_schema.sql` | Schema awal yang dijalankan saat inisialisasi  |

---

## Menjalankan dengan Docker

Prasyarat: Docker dan Docker Compose.

```bash
cd postgres
cp .env.example .env
docker compose up -d
```

Cek status container:

```bash
docker compose ps
```

Melihat log:

```bash
docker compose logs -f postgres
```

Menghentikan container tanpa menghapus data:

```bash
docker compose down
```

Menghapus container beserta data:

```bash
docker compose down -v
```

Dari root proyek:

```bash
make db-up
make db-down
make db-reset
```

---

## Menjalankan dengan PostgreSQL Lokal

Jika sudah memiliki PostgreSQL lokal:

```bash
createdb unesarent
psql -d unesarent -f postgres/initdb/001_schema.sql
```

Pastikan user database memiliki izin untuk membuat tabel, index, function, dan trigger.

---

## Koneksi

Nilai default development:

```text
postgres://unesarent:unesarent_dev@localhost:5432/unesarent?sslmode=disable
```

Nilai tersebut tersedia sebagai `DATABASE_URL` di `.env.example`. Ganti password sebelum
digunakan di luar development.

| Parameter | Default        |
| --------- | -------------- |
| Host      | `localhost`    |
| Port      | `5432`         |
| Database  | `unesarent`    |
| User      | `unesarent`    |
| Password  | `unesarent_dev` |
| SSL mode  | `disable`      |

---

## Schema

### Tabel `users`

| Kolom           | Tipe         | Constraint                              |
| --------------- | ------------ | --------------------------------------- |
| `id`            | UUID         | Primary key, default `gen_random_uuid()` |
| `name`          | VARCHAR(80)  | Not null, 2–80 karakter setelah trim    |
| `email`         | VARCHAR(254) | Not null, unik case-insensitive         |
| `password_hash` | TEXT         | Not null, tidak boleh kosong            |
| `created_at`    | TIMESTAMPTZ  | Default `now()`                         |
| `updated_at`    | TIMESTAMPTZ  | Default `now()`, diperbarui trigger     |

### Index

```sql
CREATE UNIQUE INDEX users_email_lower_key ON users (lower(email));
```

Index ini memastikan `budi@example.com` dan `BUDI@example.com` dianggap email yang sama.

### Trigger

Function `set_updated_at()` dan trigger `users_set_updated_at` memperbarui kolom
`updated_at` setiap kali baris diubah.

### Inisialisasi

`initdb/001_schema.sql` hanya dijalankan oleh image PostgreSQL ketika volume database
masih kosong. Jika schema diubah setelah container pernah berjalan, lakukan reset:

```bash
docker compose down -v
docker compose up -d
```

Untuk perubahan schema pada data yang sudah ada, gunakan migration tool seperti
`golang-migrate`, `goose`, atau `atlas`.

---

## Integrasi Backend

Saat backend akan disambungkan ke PostgreSQL:

1. Tambahkan driver PostgreSQL, misalnya `github.com/jackc/pgx/v5`.
2. Baca `DATABASE_URL` dari environment backend.
3. Implementasikan interface `user.Store` di `backend/internal/user`:

```go
type Store interface {
    Create(ctx context.Context, u User) error
    FindByEmail(ctx context.Context, email string) (User, error)
    FindByID(ctx context.Context, id string) (User, error)
}
```

4. Sesuaikan pembuatan ID. Backend saat ini memakai string hex 32 karakter, sedangkan
   schema memakai UUID.
5. Ganti inisialisasi `MemoryStore` di `cmd/server/main.go` dengan store PostgreSQL.
6. Tambahkan integration test dan migration.
7. Jangan commit file `.env` yang berisi kredensial asli.

---

## Troubleshooting

| Masalah                          | Solusi                                                          |
| -------------------------------- | --------------------------------------------------------------- |
| Port 5432 sudah dipakai          | Ubah `POSTGRES_PORT` di `.env`                                  |
| Schema tidak muncul              | Volume sudah pernah dibuat; jalankan `docker compose down -v`   |
| Container tidak sehat            | Periksa `docker compose logs postgres`                          |
| Password authentication failed   | Samakan kredensial `.env` dengan `DATABASE_URL`                 |
| Data hilang setelah restart      | Jangan gunakan `down -v`; named volume menyimpan data           |
| Perubahan SQL tidak diterapkan   | Init script hanya berjalan pada volume kosong                   |


