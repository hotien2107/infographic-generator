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
- In-memory project store để unblock frontend/backend integration sớm.
- Local object-storage adapter để lưu file upload vào thư mục cục bộ khi chưa kết nối S3-compatible storage.
- Test API happy path và validation quan trọng của Sprint 1.

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
- `OBJECT_STORAGE_DIR`
- `MAX_UPLOAD_SIZE_MB`
- `ALLOWED_FILE_TYPES`

## Bộ tài liệu sprint

Các tài liệu quản trị sprint và backlog triển khai được đặt tại thư mục `docs/sprints/`, bao gồm backlog tổng, Definition of Done, sprint plan cho từng sprint và sprint board cho Sprint 1.
