# UNESARENT

Root proyek dengan:

- **Frontend**: React 19 + Vite (JavaScript)
- **Backend**: Go (standard library `net/http`)
- **Auth**: register + login + endpoint `me` menggunakan **JWT (HS256)** dan password **bcrypt**
- **Database**: PostgreSQL disiapkan di `postgres/` (belum terhubung ke backend)

Tampilan sengaja dibuat minimal. Kamu tinggal membuat halaman login/register sendiri,
semua logikanya sudah tersedia lewat `useAuth()`.

---

## Struktur

```
.
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── auth/
│   │   │   ├── jwt.go
│   │   │   └── password.go
│   │   ├── config/
│   │   ├── httpapi/
│   │   └── user/
│   ├── .env.example
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── auth/
│   │   │   ├── AuthContext.jsx
│   │   │   ├── useAuth.js
│   │   │   ├── RequireAuth.jsx
│   │   │   └── tokenStorage.js
│   │   ├── lib/
│   │   │   ├── http.js
│   │   │   └── authApi.js
│   │   ├── App.jsx
│   │   └── main.jsx
│   └── vite.config.js
├── postgres/
│   ├── docker-compose.yml
│   ├── .env.example
│   ├── README.md
│   └── initdb/001_schema.sql
├── Makefile
└── README.md
```

---

## Menjalankan

Prasyarat: Go 1.26+, Node.js 20.19+ (atau 22.12+), dan Docker (opsional, untuk PostgreSQL).

### 1. Backend

```bash
cd backend
cp .env.example .env
go run ./cmd/server
```

Ganti `JWT_SECRET` di `.env` dengan nilai acak minimal 32 karakter.

Server berjalan di `http://localhost:8080`. Cek dengan:

```bash
curl http://localhost:8080/api/health
```

Response: `{"status":"ok"}`

### 2. Frontend

Di terminal lain:

```bash
cd frontend
npm install
npm run dev
```

Buka `http://localhost:5173`.

### Jalankan keduanya sekaligus

```bash
make install
make dev
```

---

## PostgreSQL (belum terhubung)

Setup PostgreSQL ada di folder `postgres/`, berisi Docker Compose, schema awal `users`,
dan contoh environment. Backend saat ini **masih memakai penyimpanan JSON**, jadi database
ini belum dipakai aplikasi.

```bash
cd postgres
cp .env.example .env
docker compose up -d
```

Atau dari root proyek:

```bash
make db-up
make db-down
make db-reset
```

Schema dijalankan otomatis saat container pertama kali dibuat. Detail koneksi, skema,
dan cara menjalankan tanpa Docker ada di `postgres/README.md`.

---

## Endpoint API

| Method | Path                 | Auth   | Deskripsi                          |
| ------ | -------------------- | ------ | ---------------------------------- |
| GET    | `/api/health`        | -      | Health check                       |
| POST   | `/api/auth/register` | -      | Daftar user, langsung dapat token  |
| POST   | `/api/auth/login`    | -      | Login, dapat token                 |
| GET    | `/api/auth/me`       | Bearer | Data user yang sedang login        |

### Register

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Budi Santoso","email":"budi@example.com","password":"rahasia123"}'
```

`201 Created`

```json
{
  "token": "eyJhbGciOi...",
  "token_type": "Bearer",
  "expires_at": "2026-11-15T11:57:02+07:00",
  "user": {
    "id": "a111cd4a97e652933cc3474a66c3ad3c",
    "name": "Budi Santoso",
    "email": "budi@example.com",
    "created_at": "2026-11-14T04:57:02Z"
  }
}
```

### Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"budi@example.com","password":"rahasia123"}'
```

`200 OK` dengan bentuk response yang sama seperti register.

### Me

```bash
curl http://localhost:8080/api/auth/me \
  -H 'Authorization: Bearer <token>'
```

```json
{
  "user": {
    "id": "a111cd4a97e652933cc3474a66c3ad3c",
    "name": "Budi Santoso",
    "email": "budi@example.com",
    "created_at": "2026-11-14T04:57:02Z"
  }
}
```

### Format error

```json
{
  "error": {
    "code": "validation_error",
    "message": "Data yang dikirim tidak valid",
    "fields": {
      "email": "Format email tidak valid",
      "password": "Password minimal 8 karakter"
    }
  }
}
```

Kode error yang mungkin: `validation_error` (422), `email_taken` (409),
`invalid_credentials` (401), `missing_token` / `invalid_token` / `token_expired` (401),
`invalid_body` (400), `internal_error` (500).

> Catatan: body register/login memakai `DisallowUnknownFields`, jadi kirim hanya field
> yang terdaftar (`name`, `email`, `password`). Jangan kirim `confirmPassword` ke API.

---

## Menghubungkan ke tampilan

### Login

```jsx
import { useState } from 'react'
import { useAuth } from './auth/useAuth'

export default function LoginPage() {
  const { login, pending, error } = useAuth()
  const [form, setForm] = useState({ email: '', password: '' })

  async function handleSubmit(e) {
    e.preventDefault()
    try {
      await login(form)
    } catch (err) {
      console.error(err)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="email"
        value={form.email}
        onChange={(e) => setForm({ ...form, email: e.target.value })}
        placeholder="Email"
        required
      />
      <input
        type="password"
        value={form.password}
        onChange={(e) => setForm({ ...form, password: e.target.value })}
        placeholder="Password"
        required
      />
      {error && <p>{error.message}</p>}
      {error?.fields?.email && <p>{error.fields.email}</p>}
      {error?.fields?.password && <p>{error.fields.password}</p>}
      <button disabled={pending}>{pending ? 'Memproses...' : 'Login'}</button>
    </form>
  )
}
```

### Register

```jsx
const { register, pending, error } = useAuth()

await register({ name: 'Budi', email: 'budi@example.com', password: 'rahasia123' })
```

Register otomatis login (token langsung disimpan oleh context).

### Cek status & logout

```jsx
const { status, user, isAuthenticated, logout } = useAuth()
```

Nilai `status`: `loading`, `anonymous`, atau `authenticated`.

### Guard halaman

```jsx
import { RequireAuth } from './auth/RequireAuth'

<RequireAuth fallback={<LoginPage />}>
  <Dashboard />
</RequireAuth>
```

Kalau memakai `react-router-dom`:

```jsx
if (status === 'loading') return <p>Memuat...</p>
if (!isAuthenticated) return <Navigate to="/login" replace />
```

### Request ke endpoint lain

```js
import { request } from './lib/http'

const data = await request('/api/sesuatu')
```

`request` otomatis menambahkan `Authorization: Bearer <token>` dan melempar `ApiError`
dengan properti `status`, `code`, dan `fields`.
Kalau backend membalas 401 karena token kedaluwarsa, panggil `logout()`.

---

## Environment variable backend

| Variable       | Default                  | Keterangan                              |
| -------------- | ------------------------ | --------------------------------------- |
| `APP_ADDR`     | `:8080`                  | Alamat listen                           |
| `JWT_SECRET`   | secret dev (dengan warning) | **Wajib diganti**, minimal 32 karakter |
| `JWT_TTL`      | `24h`                    | Masa berlaku access token               |
| `CORS_ORIGINS` | `http://localhost:5173,...` | Origin frontend, pisahkan koma       |
| `STORE_PATH`   | `data/users.json`        | File penyimpanan user (dev)             |

Frontend memakai `VITE_API_URL` (opsional). Saat development biarkan kosong karena
Vite mem-proxy `/api` ke `http://localhost:8080`.

---

## Catatan keamanan & production

- Password di-hash bcrypt; hash tidak pernah dikirim ke client.
- Login memakai pesan generik agar tidak membocorkan email mana yang terdaftar.
- Token di client disimpan di `localStorage` (praktis, tetapi rentan XSS). Untuk
  production yang lebih ketat, gunakan access token di memori + refresh token httpOnly.
- Logout bersifat client-side karena JWT stateless. Jika butuh invalidasi token,
  tambahkan refresh token + tabel sesi/denylist.
- Penyimpanan user saat ini file JSON dengan lock (cukup untuk dev). Setup PostgreSQL
  sudah tersedia di `postgres/`; untuk memakainya, implementasikan interface `user.Store`
  lalu sambungkan `DATABASE_URL`.
- Sebelum deploy: set `JWT_SECRET` dari secret manager, aktifkan HTTPS, batasi
  `CORS_ORIGINS`, dan tambahkan rate limiting pada endpoint login/register.
