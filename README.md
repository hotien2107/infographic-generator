# AI Infographic Generator

AI Infographic Generator là dự án xây dựng nền tảng web dùng AI để chuyển đổi tài liệu, văn bản tự do hoặc dữ liệu có cấu trúc thành infographic trực quan.

Dự án tập trung vào ba năng lực cốt lõi:
- hiểu nội dung đầu vào và xác định các ý chính;
- tổ chức thông tin thành cấu trúc phù hợp để trực quan hóa;
- tạo ra infographic dễ đọc, dễ chia sẻ và có giá trị sử dụng thực tế.

## Sprint 1 codebase

Repository hiện có hai phần tách biệt để frontend và backend có thể phát triển/triển khai độc lập:
- `backend/`: REST API Go cho Sprint 1, bám theo contract ở `contracts/sprint-1-api.yaml`.
- `frontend/`: ứng dụng React + TypeScript + Vite, sử dụng các component theo phong cách shadcn/ui để gọi backend và hiển thị workflow tạo project/upload tài liệu.

### Thành phần chính
- REST API Go cho các luồng `create project`, `get project`, `upload document`.
- React frontend riêng biệt cho local development và các môi trường deploy độc lập với backend.
- In-memory project store để unblock frontend/backend integration sớm.
- Local object-storage adapter để lưu file upload vào thư mục cục bộ khi chưa kết nối S3-compatible storage.
- Test API happy path và validation quan trọng của Sprint 1.

## Chạy backend local

```bash
cd backend
cp .env.example .env
make tidy
make run
```

API mặc định chạy ở `http://localhost:8080`.

### Biến môi trường backend chính
- `APP_ENV`
- `API_PORT`
- `OBJECT_STORAGE_DIR`
- `MAX_UPLOAD_SIZE_MB`
- `ALLOWED_FILE_TYPES`
- `FRONTEND_ORIGIN`

`FRONTEND_ORIGIN` mặc định là `http://localhost:5173` để frontend React có thể gọi API từ origin riêng trong local development.

## Chạy frontend local

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Frontend mặc định chạy ở `http://localhost:5173`.

### Biến môi trường frontend chính
- `VITE_API_BASE_URL`: địa chỉ backend API, mặc định là `http://localhost:8080`.

## Tách biệt frontend và backend

- Frontend nằm hoàn toàn trong thư mục `frontend/` và có vòng đời build riêng.
- Backend vẫn giữ vai trò cung cấp contract/API ở `backend/`.
- Backend đã bật CORS cho `FRONTEND_ORIGIN` để local React app gọi trực tiếp vào Go API mà không cần ghép chung server.

## Bộ tài liệu sprint

Các tài liệu quản trị sprint và backlog triển khai được đặt tại thư mục `docs/sprints/`, bao gồm backlog tổng, Definition of Done, sprint plan cho từng sprint và sprint board cho Sprint 1.
