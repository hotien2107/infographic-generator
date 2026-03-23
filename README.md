# AI Infographic Generator

AI Infographic Generator là dự án xây dựng nền tảng web dùng AI để chuyển đổi tài liệu, văn bản tự do hoặc dữ liệu có cấu trúc thành infographic trực quan.

Dự án tập trung vào ba năng lực cốt lõi:
- hiểu nội dung đầu vào và xác định các ý chính;
- tổ chức thông tin thành cấu trúc phù hợp để trực quan hóa;
- tạo ra infographic dễ đọc, dễ chia sẻ và có giá trị sử dụng thực tế.

## Sprint 1 codebase

Repository hiện có:
- `backend/`: REST API Go cho Sprint 1.
- `frontend/`: giao diện React tách biệt với backend, build bằng Vite + React + Tailwind theo style `shadcn/ui`.

Backend foundation tại `backend/` bám theo contract ở `contracts/sprint-1-api.yaml`, và frontend tại `frontend/` tiêu thụ trực tiếp các endpoint trong contract đó.

### Thành phần chính
- REST API Go cho các luồng `create project`, `get project`, `upload document`.
- PostgreSQL store để lưu metadata của `projects` và `documents`, tự khởi tạo schema khi service boot.
- MinIO object-storage adapter cho file upload; file lớn sẽ được gửi bằng multipart upload để tối ưu throughput và độ ổn định.
- Frontend React riêng biệt với backend, cung cấp dashboard tạo project, nạp lại trạng thái và upload tài liệu theo Sprint 1.
- Test API happy path và validation quan trọng của Sprint 1 bằng fake dependency để không cần service ngoài khi chạy unit test.

### Chạy backend local

```bash
cd backend
cp .env.example .env
make tidy
make run
```

API mặc định chạy ở `http://localhost:8080`.

### Chạy frontend local

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Frontend mặc định chạy ở `http://localhost:5173` và sẽ gọi backend qua:
- `VITE_API_BASE_URL` nếu bạn cấu hình URL tuyệt đối;
- hoặc proxy dev server tới `http://localhost:8080`.

### Build frontend

```bash
cd frontend
npm run build
```

### Biến môi trường chính

#### Backend
- `APP_ENV`
- `API_PORT`
- `MAX_UPLOAD_SIZE_MB`
- `ALLOWED_FILE_TYPES`
- `POSTGRES_URL`
- `MINIO_ENDPOINT`
- `MINIO_ACCESS_KEY`
- `MINIO_SECRET_KEY`
- `MINIO_BUCKET`
- `MINIO_USE_SSL`
- `MINIO_AUTO_CREATE_BUCKET`
- `MINIO_MULTIPART_THRESHOLD_MB`
- `MINIO_MULTIPART_PART_SIZE_MB`

#### Frontend
- `VITE_API_BASE_URL`
- `VITE_API_PROXY_TARGET`

## Bộ tài liệu sprint

Các tài liệu quản trị sprint và backlog triển khai được đặt tại thư mục `docs/sprints/`, bao gồm backlog tổng, Definition of Done, sprint plan cho từng sprint và sprint board cho Sprint 1.
