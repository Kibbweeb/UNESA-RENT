<div align="center">

# UNESARENT Backend

API autentikasi berbasis Go dengan JWT, bcrypt, dan penyimpanan user yang dapat diganti.

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![JWT](https://img.shields.io/badge/Auth-JWT%20HS256-000000?style=flat-square&logo=jsonwebtokens&logoColor=white)](https://jwt.io/)
[![bcrypt](https://img.shields.io/badge/Password-bcrypt-4B8BBE?style=flat-square)](https://pkg.go.dev/golang.org/x/crypto/bcrypt)

[Dokumentasi utama](../README.md) · [API](#api) · [Konfigurasi](#konfigurasi)

</div>

---

## Ringkasan

Backend UNESARENT menyediakan register, login, dan endpoint profil menggunakan access
token JWT. Seluruh routing memakai standard library `net/http` tanpa framework tambahan.
Password disimpan sebagai hash bcrypt, sedangkan user disimpan di file JSON untuk
kebutuhan development.

Repositori ini adalah bagian dari monorepo [UNESARENT](../README.md).

---

## Fitur

- Register dan login dengan validasi input
- Access token JWT HS256
- Verifikasi signature, issuer, dan masa berlaku token
- Password bcrypt
- Middleware Bearer token
- CORS untuk origin frontend yang dikonfigurasi
- Request logging dan panic recovery
- Body limit 1 MB
- Graceful shutdown
- Penyimpanan JSON atomik dengan permission `0600`
- Interface `user.Store` untuk mempermudah migrasi ke database

---

## Teknologi

| Komponen    | Implementasi                          |
| ----------- | ------------------------------------- |
| Bahasa      | Go 1.26+                              |
| HTTP        | `net/http`                            |
| JWT         | `github.com/golang-jwt/jwt/v5`        |
| Password    | `golang.org/x/crypto/bcrypt`          |
| Logging     | `log/slog`                            |
| Storage     | File JSON dengan mutex                |

---

## Struktur

```text
backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   │   ├── jwt.go
│   │   └── password.go
│   ├── config/
│   │   ├── config.go
│   │   └── env.go
│   ├── httpapi/
│   │   ├── auth_handler.go
│   │   ├── middleware.go
│   │   ├── respond.go
│   │   └── router.go
│   └── user/
│       ├── model.go
│       ├── service.go
│       └── store.go
├── .env.example
└── go.mod
```

| Package              | Tanggung jawab                                        |
| -------------------- | ----------------------------------------------------- |
| `cmd/server`         | Bootstrap aplikasi, HTTP server, graceful shutdown    |
| `internal/config`    | Pembacaan environment dan file `.env`                 |
| `internal/auth`      | Penerbitan/verifikasi JWT dan hashing password        |
| `internal/httpapi`   | Router, handler, middleware, dan format response      |
| `internal/user`      | Model, validasi, business logic, dan penyimpanan user |

---

## Menjalankan

Prasyarat: Go 1.26+.

```bash
cd backend
cp .env.example .env
go run ./cmd/server
```

Ganti `JWT_SECRET` di `.env` dengan string acak minimal 32 karakter:

```bash
openssl rand -base64 48
```

Server berjalan di `http://localhost:8080`.

Health check:

```bash
curl http://localhost:8080/api/health
```

Response:

```json
{
  "status": "ok"
}
```

---

## Konfigurasi

Backend membaca environment variable dan file `.env` di direktori kerja. Environment
variable yang sudah di-set tidak ditimpa oleh `.env`.

| Variable       | Default                     | Wajib | Deskripsi                           |
| -------------- | --------------------------- | ----- | ----------------------------------- |
| `APP_ADDR`     | `:8080`                     | Tidak | Alamat listen HTTP server           |
| `JWT_SECRET`   | secret development          | Ya    | Secret JWT, minimal 32 karakter     |
| `JWT_TTL`      | `24h`                       | Tidak | Masa berlaku access token           |
| `CORS_ORIGINS` | `http://localhost:5173,...` | Tidak | Origin frontend, dipisahkan koma    |
| `STORE_PATH`   | `data/users.json`           | Tidak | Lokasi penyimpanan user             |

`JWT_TTL` memakai format durasi Go, misalnya `30m`, `12h`, atau `24h`. Format seperti
`7d` tidak didukung.

Jika `JWT_SECRET` kosong, server memakai secret development dan menampilkan peringatan.
Konfigurasi tersebut tidak boleh dipakai di production.

---

## API

Dokumentasi request dan response lengkap tersedia di [README utama](../README.md#api-reference).

| Method | Endpoint             | Auth   | Deskripsi                   |
| ------ | -------------------- | ------ | --------------------------- |
| GET    | `/api/health`        | Tidak  | Health check                |
| POST   | `/api/auth/register` | Tidak  | Registrasi user             |
| POST   | `/api/auth/login`    | Tidak  | Login user                  |
| GET    | `/api/auth/me`       | Bearer | Data user yang sedang login |

Contoh login:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"budi@example.com","password":"rahasia123"}'
```

Contoh endpoint terproteksi:

```bash
curl http://localhost:8080/api/auth/me \
  -H 'Authorization: Bearer <token>'
```

---

## Alur Autentikasi

### Register

1. Service memvalidasi nama, email, dan password.
2. Email dinormalisasi menjadi lowercase.
3. Password di-hash dengan bcrypt.
4. User baru disimpan melalui `user.Store`.
5. JWT diterbitkan dan dikembalikan bersama data user.

### Login

1. Email dicari di storage.
2. Password dibandingkan dengan hash bcrypt.
3. Jika email tidak ditemukan, service tetap menjalankan bcrypt comparison terhadap
   dummy hash untuk mengurangi perbedaan waktu respons.
4. JWT diterbitkan jika kredensial valid.

### Request terproteksi

1. Middleware membaca header `Authorization: Bearer <token>`.
2. Token diverifikasi menggunakan secret, issuer, algoritma HS256, dan expiry.
3. User ID dari claim `sub` disimpan di context.
4. Handler mengambil data user terbaru dari storage.

---

## Penyimpanan

Implementasi saat ini adalah `user.MemoryStore`:

- Data disimpan di memori dengan `sync.RWMutex`.
- Data dipersist ke file JSON jika `STORE_PATH` diisi.
- Penulisan file dilakukan secara atomik menggunakan file sementara dan `rename`.
- File dibuat dengan permission `0600`.
- Email disimpan lowercase dan dicek unik.
- Data hilang jika `STORE_PATH` kosong dan server di-restart.

Interface yang tersedia:

```go
type Store interface {
    Create(ctx context.Context, u User) error
    FindByEmail(ctx context.Context, email string) (User, error)
    FindByID(ctx context.Context, id string) (User, error)
}
```

Untuk production, implementasikan interface tersebut dengan PostgreSQL. Setup database
dan schema awal sudah tersedia di [postgres/README.md](../postgres/README.md).

---

## Build dan Verifikasi

```bash
go build ./...
go vet ./...
gofmt -l .
```

Menjalankan server dengan konfigurasi eksplisit:

```bash
JWT_SECRET=$(openssl rand -base64 48) go run ./cmd/server
```

