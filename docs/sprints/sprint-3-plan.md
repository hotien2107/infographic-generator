# Sprint 3 Plan – AI Understanding Layer

## Mục tiêu
Từ raw content, tạo summary JSON có cấu trúc để phục vụ bước structuring.

## Deliverable chính
- AI text service integration.
- Summary schema + validator.
- Lưu summary JSON và usage event.
- Trạng thái understanding rõ ràng cho project/job.
- Màn hình nội bộ hoặc log để kiểm tra summary.

## Công việc thực hiện
- Backend/AI: gọi AI provider, validate schema, retry/repair output lỗi.
- Product/BA: review chất lượng summary và so khớp expectation nghiệp vụ.
- Frontend: hiển thị trạng thái phân tích nội dung.
- QA: test timeout, rate limit, output thiếu trường.

## Rủi ro
- Output AI thiếu ổn định.
- Schema quá chặt gây fail nhiều.
- Chi phí AI tăng nếu retry không kiểm soát.
