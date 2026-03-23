# Sprint 2 Plan – Document Ingestion & Processing Foundation

## 1. Scope

Sprint 2 mở rộng từ trạng thái “project đã upload document” để tạo nền tảng cho pipeline AI ở các sprint sau. Trọng tâm là lifecycle ingestion/processing, không phải extraction AI thật.

## 2. In-scope

- Mở rộng state machine cho `project` và `document`.
- Tự động enqueue document sau upload và hỗ trợ trigger thủ công.
- Giả lập background processing bằng in-memory worker/goroutine.
- Lưu metadata tối thiểu cho processing.
- Mở rộng project detail để frontend theo dõi trạng thái.
- Bổ sung test backend/frontend cho các trạng thái chính.
- Cập nhật contract và README.

## 3. Out-of-scope

- OCR/PDF parser thật.
- Queue production như Redis/Kafka/SQS.
- AI summarization/generation/rendering.
- Auth/authorization production-ready.

## 4. Deliverables

| Mã | Deliverable | Mô tả |
| --- | --- | --- |
| S2-D1 | Backend processing lifecycle | Worker giả lập + transition state rõ ràng |
| S2-D2 | Persistence metadata | Lưu `processing_started_at`, `processing_finished_at`, `error_message`, `extracted_text_preview` |
| S2-D3 | Sprint 2 API contract | OpenAPI cho upload/get detail/trigger processing |
| S2-D4 | Frontend processing UI | Status badges, summary card, trigger processing, refresh state |
| S2-D5 | Test coverage | Bao phủ happy path + failure path của processing |

## 5. State machine

### 5.1 Project status

- `draft`
- `uploaded`
- `processing`
- `processed`
- `failed`

### 5.2 Document status

- `uploaded`
- `queued`
- `processing`
- `processed`
- `failed`

### 5.3 Current step

- `waiting_for_upload`
- `uploaded`
- `queued_for_processing`
- `extracting`
- `ready_for_generation`
- `failed`

## 6. API

- `POST /api/v1/projects`
- `GET /api/v1/projects/{projectId}`
- `POST /api/v1/projects/{projectId}/documents`
- `POST /api/v1/projects/{projectId}/processing`

Toàn bộ response dùng envelope nhất quán:

```json
{
  "data": {},
  "error": null,
  "meta": {
    "request_id": "uuid",
    "timestamp": "date-time"
  }
}
```

## 7. Test checklist

- [x] Upload xong trả về `202`.
- [x] Project/detail phản ánh `processing_summary`.
- [x] Auto-processing chuyển `uploaded -> queued -> processing -> processed`.
- [x] Trigger thủ công có thể đẩy document mới nhất vào queue.
- [x] Worker giả lập có thể tạo failure state theo filename pattern.
- [x] Frontend render được badge/trạng thái processing chính.

## 8. Risk / follow-up

- Worker hiện dùng in-memory queue nên không survive process restart.
- Failure simulation dựa trên filename pattern, chỉ dùng cho development/test.
- Sprint 3 nên thay worker bằng queue abstraction có retry/dead-letter rõ ràng.
