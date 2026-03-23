# AI Infographic Generator

AI Infographic Generator là dự án xây dựng nền tảng web dùng AI để chuyển đổi tài liệu, văn bản tự do hoặc dữ liệu có cấu trúc thành infographic trực quan.

## Repo structure

- `backend/`: Go REST API.
- `frontend/`: React + Vite + Tailwind + shadcn/ui.
- `contracts/`: OpenAPI contracts cho từng sprint.
- `docs/sprints/`: sprint plans, boards, DoD và implementation notes.

## Sprint 1 hiện có

- Create project.
- Get project detail.
- Upload document.
- Validation siết chặt theo contract.
- Response envelope thống nhất `data / error / meta`.

## Sprint 2 hiện có

- Project/document processing lifecycle rõ ràng.
- Auto-processing sau upload bằng in-memory worker.
- Manual trigger processing qua API riêng.
- Processing summary + metadata extraction giả lập.
- Frontend hiển thị status badge, lifecycle summary, fail/success state.

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

Frontend mặc định chạy ở `http://localhost:5173` và gọi backend qua:

- `VITE_API_BASE_URL` nếu bạn cấu hình URL tuyệt đối.
- hoặc proxy dev server tới `http://localhost:8080`.

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

## Tài liệu chính

- Sprint 1 contract: `contracts/sprint-1-api.yaml`
- Sprint 2 contract: `contracts/sprint-2-api.yaml`
- Sprint 1 plan: `docs/sprints/sprint-1-plan.md`
- Sprint 1 board: `docs/sprints/sprint-1-board.md`
- Sprint 1 implementation notes: `docs/sprints/sprint-1-implementation-notes.md`
- Sprint 2 plan: `docs/sprints/sprint-2-plan.md`
- Sprint 2 board: `docs/sprints/sprint-2-board.md`
- Definition of Done: `docs/sprints/definition-of-done.md`
