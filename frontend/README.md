# Frontend

Frontend được tách riêng khỏi backend và xây dựng bằng React + TypeScript + Vite, với bộ component theo phong cách shadcn/ui.

## Chạy local

```bash
cp .env.example .env
npm install
npm run dev
```

Ứng dụng mặc định chạy ở `http://localhost:5173`.

## Biến môi trường

- `VITE_API_BASE_URL`: URL của backend API, mặc định là `http://localhost:8080`.

## Workflow hiện hỗ trợ

- tạo project mới;
- chọn input mode `file` hoặc `text`;
- upload tài liệu cho project ở `file` mode;
- lấy lại project detail và hiển thị danh sách document từ backend.
