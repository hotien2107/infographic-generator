# Sprint 2 Plan – Document Ingestion & Extraction

## Mục tiêu
Thiết lập luồng nhận dữ liệu đầu vào từ file hoặc text và chuyển đổi thành raw content có thể dùng cho pipeline AI.

## Deliverable chính
- API nhập text trực tiếp.
- Extraction service/worker cơ bản cho PDF/TXT.
- Trạng thái project: `uploaded`, `extracting`, `extracted`, `failed`.
- Lưu raw content và metadata extraction.
- UI hiển thị trạng thái extraction.

## Công việc thực hiện
- Backend: tạo endpoint text input, service extraction, persistence cho raw content.
- Data/AI: chuẩn hóa cấu trúc raw content và metadata như `page_count`, `file_type`, `section_headings`.
- Frontend: bổ sung chế độ nhập text trực tiếp và trạng thái xử lý.
- QA: test file hợp lệ, file rỗng, file lỗi encoding và text input.

## Rủi ro
- PDF extraction không ổn định.
- Queue/worker xử lý lỗi ngầm.
- Encoding tiếng Việt bị sai khi trích xuất.
