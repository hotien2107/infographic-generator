# Infographic Project Workspace

Ứng dụng giúp đội ngũ nội dung quản lý dự án, tài liệu đầu vào và theo dõi tiến độ chuẩn bị dữ liệu cho infographic.

## Cấu trúc repo

- `backend/`: Go REST API.
- `frontend/`: React + Vite + Tailwind.
- `contracts/`: OpenAPI contract cho API hiện tại.
- `docs/`: tài liệu sản phẩm và kế hoạch triển khai.

## Tính năng hiện có

- Trang tổng quan với số liệu dự án và tài liệu.
- Danh sách dự án riêng, hỗ trợ tạo, sửa, xóa.
- Trang chi tiết dự án với thông tin tổng quan và danh sách tài liệu.
- Thêm, đổi tên, xóa tài liệu ngay trong màn hình chi tiết dự án.
- Hỗ trợ nhập text trực tiếp hoặc upload PDF/TXT và theo dõi pipeline extraction (uploaded/extracting/extracted/failed).
- Backend tự đọc cấu hình từ file `.env` nếu file tồn tại; biến nào không có thì mới dùng giá trị mặc định trong code.

## Chạy backend local

```bash
cd backend
cp .env.example .env
make tidy
make run
```

Backend sẽ tự nạp file `.env` trong thư mục `backend/` khi khởi động.

### Biến môi trường backend chính

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
- `AUTO_PROCESS_DOCUMENTS`
- `PROCESSING_QUEUE_BUFFER`
- `PROCESSING_STEP_DELAY_MS`
- `PROCESSING_FAIL_PATTERN`

## Chạy frontend local

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Frontend mặc định chạy ở `http://localhost:5173`.

## API chính

- `GET /api/v1/dashboard/summary`
- `GET /api/v1/projects`
- `POST /api/v1/projects`
- `GET /api/v1/projects/{projectId}`
- `PATCH /api/v1/projects/{projectId}`
- `DELETE /api/v1/projects/{projectId}`
- `POST /api/v1/projects/{projectId}/documents`
- `GET /api/v1/projects/{projectId}/documents`
- `PATCH /api/v1/projects/{projectId}/documents/{documentId}`
- `DELETE /api/v1/projects/{projectId}/documents/{documentId}`
- `POST /api/v1/projects/{projectId}/processing`
- `POST /api/v1/projects/{projectId}/text`

## Test

### Backend

```bash
cd backend
go test ./...
```

### Frontend

```bash
cd frontend
npm install
npm run test
npm run build
```

## Contract

- API hiện tại: `contracts/sprint-2-api.yaml`
