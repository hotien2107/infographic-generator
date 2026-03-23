# Sprint 2 Board – Document Ingestion & Processing Foundation

## Sprint board

| ID | Loại | Công việc | Owner đề xuất | Trạng thái | Ghi chú |
| --- | --- | --- | --- | --- | --- |
| S2-B01 | Analysis | Chốt state machine project/document/current_step | Tech Lead / Backend / Frontend | Done | Đồng bộ contract + UI |
| S2-B02 | Backend | Refactor service/store cho upload + processing lifecycle | Backend | Done | Giữ flow Sprint 1 không bị gãy |
| S2-B03 | Backend | Thêm in-memory worker và endpoint trigger processing | Backend | Done | Dễ thay bằng queue thật ở sprint sau |
| S2-B04 | Frontend | Tách API layer, hooks và component workflow | Frontend | Done | Giảm App.jsx monolith |
| S2-B05 | Frontend | Hiển thị processing summary, status badge, error state | Frontend | Done | UI phản ánh auto/manual processing |
| S2-B06 | QA | Viết test cho validation, upload, processing success/failure | QA / Dev | Done | Bao phủ backend + frontend |
| S2-B07 | Docs | Cập nhật README, contract Sprint 1/2, sprint docs | Tech Lead | Done | Ghi rõ scope và follow-up |

## Follow-up cho Sprint 3+

- Thay in-memory worker bằng queue bền vững.
- Tách extraction pipeline theo interface rõ ràng cho AI service.
- Bổ sung observability tốt hơn cho từng task processing.
