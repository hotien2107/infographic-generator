# Sprint 4 Plan – Structuring Layer

## Mục tiêu
Chuyển summary thành infographic specification JSON có cấu trúc rõ ràng, sẵn sàng cho bước rendering.

## Deliverable chính
- Spec schema cho infographic.
- Logic mapping summary -> sections/layout/content blocks.
- Validation và fallback khi thiếu dữ liệu.
- Persistence cho spec JSON.

## Công việc thực hiện
- Backend/AI: thiết kế spec schema, generator và validator.
- Product/Design: review narrative flow, section priority và hierarchy.
- QA: test với nhiều loại tài liệu và các trường hợp summary thiếu dữ liệu.

## Rủi ro
- Spec khó render nếu cấu trúc quá tự do.
- Thiếu thống nhất giữa content hierarchy và layout engine.
