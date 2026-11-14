<div align="center">

# UNESARENT Frontend

Antarmuka React dengan logika autentikasi JWT yang siap dihubungkan ke tampilan apa pun.

[![React](https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react&logoColor=black)](https://react.dev/)
[![Vite](https://img.shields.io/badge/Vite-7-646CFF?style=flat-square&logo=vite&logoColor=white)](https://vite.dev/)
[![JavaScript](https://img.shields.io/badge/JavaScript-ES2022-F7DF1E?style=flat-square&logo=javascript&logoColor=black)](https://developer.mozilla.org/docs/Web/JavaScript)

[Dokumentasi utama](../README.md) · [Auth](#autentikasi) · [Menjalankan](#menjalankan)

</div>

---

## Ringkasan

Frontend UNESARENT menyediakan lapisan autentikasi lengkap untuk aplikasi React:
state login, pemulihan sesi, pemanggilan API, penanganan error, dan guard halaman.
Tampilan bawaan sengaja minimal agar kamu dapat membuat UI sendiri tanpa mengubah logic.

Repositori ini adalah bagian dari monorepo [UNESARENT](../README.md).

---

## Fitur

- React 19 dan Vite 7
- `AuthProvider` untuk state autentikasi global
- Hook `useAuth()`
- Register otomatis login
- Pemulihan sesi dari token tersimpan
- Logout client-side
- Fetch wrapper dengan Bearer token otomatis
- `ApiError` dengan `status`, `code`, dan `fields`
- Guard `RequireAuth`
- Proxy development ke backend Go
- Build produksi siap deploy

---

## Struktur

```text
frontend/
├── src/
│   ├── auth/
│   │   ├── AuthContext.jsx
│   │   ├── RequireAuth.jsx
│   │   ├── tokenStorage.js
│   │   └── useAuth.js
│   ├── lib/
│   │   ├── authApi.js
│   │   └── http.js
│   ├── App.jsx
│   ├── index.css
│   └── main.jsx
├── .env.example
├── index.html
├── package.json
└── vite.config.js
```

| File                       | Tanggung jawab                                   |
| -------------------------- | ------------------------------------------------ |
| `auth/AuthContext.jsx`     | State user, token, status, error, dan aksi auth  |
| `auth/useAuth.js`          | Hook untuk mengakses context                     |
| `auth/RequireAuth.jsx`     | Guard halaman tanpa dependency router            |
| `auth/tokenStorage.js`     | Baca, simpan, dan hapus token                    |
| `lib/authApi.js`           | Endpoint register, login, dan me                 |
| `lib/http.js`              | Fetch wrapper, base URL, dan `ApiError`          |
| `App.jsx`                  | Placeholder tampilan                             |
| `main.jsx`                 | Entry point dan pemasangan `AuthProvider`        |

---

## Menjalankan

Prasyarat: Node.js 20.19+ atau 22.12+ dan backend berjalan di `http://localhost:8080`.

```bash
cd frontend
npm install
npm run dev
```

Buka `http://localhost:5173`.

Vite mem-proxy request `/api` ke backend sehingga development tidak membutuhkan
konfigurasi CORS tambahan.

---

## Konfigurasi

File: `.env`

| Variable       | Default | Deskripsi                                       |
| -------------- | ------- | ----------------------------------------------- |
| `VITE_API_URL` | kosong  | Base URL API jika tidak menggunakan proxy Vite  |

Saat development biarkan kosong. Jika frontend dan backend berada di domain berbeda,
isi dengan URL backend tanpa trailing slash:

```env
VITE_API_URL=https://api.example.com
```

Backend juga harus mengizinkan origin frontend melalui `CORS_ORIGINS`.

---

## Autentikasi

### State `useAuth()`

```jsx
import { useAuth } from './auth/useAuth'

const {
  user,
  status,
  error,
  pending,
  isAuthenticated,
  login,
  register,
  logout,
  clearError,
} = useAuth()
```

| Properti          | Tipe               | Deskripsi                                |
| ----------------- | ------------------ | ---------------------------------------- |
| `user`            | `object \| null`   | Data user yang sedang login              |
| `status`          | `string`           | `loading`, `anonymous`, `authenticated`  |
| `error`           | `ApiError \| null` | Error terakhir dari login/register       |
| `pending`         | `boolean`          | Menandakan request auth sedang berjalan  |
| `isAuthenticated` | `boolean`          | `true` jika status authenticated         |
| `login`           | `function`         | `login({ email, password })`             |
| `register`        | `function`         | `register({ name, email, password })`    |
| `logout`          | `function`         | Menghapus token dan data user            |
| `clearError`      | `function`         | Menghapus error terakhir                 |

### Login

```jsx
const { login, pending, error } = useAuth()

async function handleLogin() {
  try {
    const user = await login({
      email: 'budi@example.com',
      password: 'rahasia123',
    })
    console.log(user)
  } catch (err) {
    console.error(err)
  }
}
```

### Register

```jsx
const { register } = useAuth()

await register({
  name: 'Budi Santoso',
  email: 'budi@example.com',
  password: 'rahasia123',
})
```

Register otomatis menyimpan token dan mengisi `user`, sehingga tidak perlu login ulang.

### Logout

```jsx
const { logout } = useAuth()

logout()
```

Logout hanya menghapus sesi di client karena JWT bersifat stateless.

### Guard halaman

```jsx
import { RequireAuth } from './auth/RequireAuth'

export default function App() {
  return (
    <RequireAuth fallback={<LoginPage />}>
      <Dashboard />
    </RequireAuth>
  )
}
```

Jika memakai `react-router-dom`:

```jsx
if (status === 'loading') return <p>Memuat...</p>
if (!isAuthenticated) return <Navigate to="/login" replace />
```

### Pemulihan sesi

Saat halaman dimuat dan token tersedia, `AuthProvider` memanggil `/api/auth/me`.
Jika token valid, status berubah menjadi `authenticated`. Jika token kedaluwarsa atau
tidak valid, token dihapus dan status menjadi `anonymous`.

---

## Penanganan Error

`ApiError` memiliki struktur berikut:

| Properti  | Deskripsi                                             |
| --------- | ----------------------------------------------------- |
| `message` | Pesan yang aman ditampilkan ke user                   |
| `status`  | HTTP status, atau `0` jika gagal koneksi              |
| `code`    | Kode error dari backend                               |
| `fields`  | Error validasi per field, misalnya `{ email: "..." }` |

Contoh:

```jsx
{error && <p>{error.message}</p>}
{error?.fields?.email && <p>{error.fields.email}</p>}
{error?.fields?.password && <p>{error.fields.password}</p>}
```

---

## Request ke Endpoint Lain

```js
import { request } from './lib/http'

const data = await request('/api/endpoint')
```

`request` otomatis menambahkan header `Authorization: Bearer <token>` jika token
tersedia. Untuk request tanpa token, gunakan opsi `auth: false`.

```js
const result = await request('/api/public', {
  method: 'POST',
  auth: false,
  body: { key: 'value' },
})
```

---

## Scripts

| Perintah        | Deskripsi                         |
| --------------- | --------------------------------- |
| `npm run dev`   | Menjalankan development server    |
| `npm run build` | Membuat build produksi ke `dist/` |
| `npm run preview` | Menjalankan preview build       |

Verifikasi build:

```bash
npm run build
```


