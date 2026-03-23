# Sprint 6 Plan – Export & Regenerate

## Mục tiêu
Hoàn thiện vòng đời MVP để người dùng có thể xem trước, tải xuống và yêu cầu tạo lại kết quả.

## Deliverable chính
- Export ảnh/tài liệu đầu ra.
- Regenerate infographic với phiên bản khác.
- Tracking version cho asset render.
- Luồng end-to-end từ input đến export.

## Công việc thực hiện
- Backend: export endpoint, versioning, regenerate job.
- Frontend: nút download/regenerate và hiển thị lịch sử phiên bản cơ bản.
- QA: test tải xuống, regenerate khác bản cũ và xử lý lỗi khi job fail.

## Rủi ro
- Regenerate sinh kết quả trùng lặp.
- Export format không nhất quán giữa trình duyệt/môi trường.
