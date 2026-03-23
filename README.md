# AI Infographic Generator

AI Infographic Generator là dự án xây dựng nền tảng web dùng AI để chuyển đổi tài liệu, văn bản tự do hoặc dữ liệu có cấu trúc thành infographic trực quan.

Dự án tập trung vào ba năng lực cốt lõi:
- hiểu nội dung đầu vào và xác định các ý chính;
- tổ chức thông tin thành cấu trúc phù hợp để trực quan hóa;
- tạo ra infographic dễ đọc, dễ chia sẻ và có giá trị sử dụng thực tế.

## Sprint 1 codebase

Repository hiện đã có backend foundation cho Sprint 1 tại `backend/`, bám theo contract ở `contracts/sprint-1-api.yaml`.

### Thành phần chính
- REST API Go cho các luồng `create project`, `get project`, `upload document`.
- PostgreSQL store để lưu metadata của `projects` và `documents`, tự khởi tạo schema khi service boot.
- MinIO object-storage adapter cho file upload; file lớn sẽ được gửi bằng multipart upload để tối ưu throughput và độ ổn định.
- Test API happy path và validation quan trọng của Sprint 1 bằng fake dependency để không cần service ngoài khi chạy unit test.

### Chạy backend local

```bash
cd backend
cp .env.example .env
make tidy
make run
```

API mặc định chạy ở `http://localhost:8080`.

### Biến môi trường chính
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

## Bộ tài liệu sprint

Các tài liệu quản trị sprint và backlog triển khai được đặt tại thư mục `docs/sprints/`, bao gồm backlog tổng, Definition of Done, sprint plan cho từng sprint và sprint board cho Sprint 1.
