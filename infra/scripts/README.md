# Infrastructure Script

### `init_minio.sh`
Digunakan untuk inisialisasi MinIO.
- Membuat bucket berdasarkan environment variable `MINIO_BUCKET`
- Mengatur bucket policy
    - Public read untuk path `/games/*`
- lifecycle rule dan versioning

---

### `backup_postgres.sh`
Script backup database PostgreSQL.
- Backup harian menggunakan `pg_dump`
- Retention default 7 hari
- Output disimpan ke folder backup di VM

---

### `backup_minio.sh`
Backup data MinIO ke disk lokal.
- Mirror bucket MinIO ke disk menggunakan `mc mirror`
- Retention berbasis tanggal / jumlah backup

---

### `restore_postgres.sh`
Script untuk restore database PostgreSQL.
- Restore database dari file dump hasil backup

---

### `restore_minio.sh`
Script untuk restore data MinIO.
- Restore bucket MinIO dari hasil mirror di disk