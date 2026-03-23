# Sprint 7 Plan – Hardening & Observability

## Mục tiêu
Tăng độ ổn định vận hành và khả năng theo dõi pipeline trước khi bước vào giai đoạn tối ưu phát hành.

## Deliverable chính
- Retry/idempotency theo step.
- Logging tập trung và metrics cơ bản.
- Dashboard/alert cho lỗi pipeline quan trọng.
- Error handling nhất quán cho API và worker.

## Công việc thực hiện
- Backend/Platform: retry policy, deduplication key, tracing và metrics.
- DevOps: dashboard/alert tối thiểu cho latency, failure rate, queue depth.
- QA: regression test và test khôi phục sau lỗi.

## Rủi ro
- Retry sai gây duplicate processing.
- Observability thiếu ngữ cảnh nên khó debug.
