# Product Backlog theo Epic / Feature / Story

Tài liệu này chuyển hóa nội dung từ PRD, User Stories, Acceptance Criteria và Agile sprint plan thành backlog có thể dùng cho planning và tracking.

## 1. Cấu trúc Epic

| Epic | Mục tiêu | User story liên quan | Sprint mục tiêu |
| --- | --- | --- | --- |
| EPIC-01 Foundation & Workspace | Tạo nền tảng dự án, project, upload, môi trường làm việc | US-01, US-13 | Sprint 1 |
| EPIC-02 Ingestion & Extraction | Nhận file/text và trích xuất raw content | US-01, US-02 | Sprint 2 |
| EPIC-03 AI Understanding | Sinh summary có cấu trúc từ raw content | US-03, US-04, US-10 | Sprint 3 |
| EPIC-04 Structuring | Chuyển summary thành infographic spec | US-05, US-10 | Sprint 4 |
| EPIC-05 Rendering & Preview | Tạo preview infographic đầu tiên | US-05, US-06, US-07 | Sprint 5 |
| EPIC-06 Export & Regenerate | Tải xuống và tạo lại infographic | US-08, US-09 | Sprint 6 |
| EPIC-07 Hardening & Observability | Tăng độ ổn định, retry, logging, metrics | US-12, US-13 | Sprint 7 |
| EPIC-08 Optimization & Release Readiness | Tối ưu chất lượng, chi phí và pre-release | US-10, US-11, US-12, US-13 | Sprint 8 |

## 2. Backlog ưu tiên theo sprint

### Sprint 1

| ID | Loại | Mô tả | Kết quả đầu ra |
| --- | --- | --- | --- |
| SP1-01 | Story | Khởi tạo cấu trúc repo và nguyên tắc tổ chức tài liệu/kỹ thuật | Repo có cấu trúc nền tảng rõ ràng |
| SP1-02 | Story | Thiết kế schema ban đầu cho user/project/document | Có schema logic và trạng thái nghiệp vụ |
| SP1-03 | Story | Xác định API contract tạo project, upload document, xem project | Có contract dùng chung FE/BE |
| SP1-04 | Story | Xây dựng backlog, DoD và sprint board nền tảng | Có artifact quản trị sprint |
| SP1-05 | Story | Thiết lập hướng dẫn local development, CI/CD và biến môi trường | Team có thể bắt đầu triển khai đồng nhất |
| SP1-06 | Story | Chuẩn bị test case cho create project và upload | Có checklist kiểm thử Sprint 1 |

### Sprint 2

| ID | Loại | Mô tả | Kết quả đầu ra |
| --- | --- | --- | --- |
| SP2-01 | Story | Nhập text trực tiếp cho project | Hỗ trợ text input |
| SP2-02 | Story | Tạo extraction worker và lưu raw content | Có pipeline extract cơ bản |
| SP2-03 | Story | Bổ sung trạng thái extraction và log | Quan sát được extraction state |
| SP2-04 | Story | Kiểm thử file lỗi/rỗng/không hỗ trợ | Tăng độ tin cậy ingestion |

### Sprint 3

| ID | Loại | Mô tả | Kết quả đầu ra |
| --- | --- | --- | --- |
| SP3-01 | Story | Tích hợp AI text service | Có khả năng tạo summary |
| SP3-02 | Story | Thiết kế summary schema và validator | Output AI được kiểm soát |
| SP3-03 | Story | Theo dõi usage và lỗi AI | Có dữ liệu vận hành bước understanding |

### Sprint 4

| ID | Loại | Mô tả | Kết quả đầu ra |
| --- | --- | --- | --- |
| SP4-01 | Story | Xây dựng infographic spec schema | Có đầu ra trung gian chuẩn hóa |
| SP4-02 | Story | Mapping summary -> section/layout/content blocks | Có pipeline structuring |
| SP4-03 | Story | Review chất lượng spec và fallback logic | Tăng ổn định cho render |

### Sprint 5

| ID | Loại | Mô tả | Kết quả đầu ra |
| --- | --- | --- | --- |
| SP5-01 | Story | Tạo rendering service đầu tiên | Sinh preview được |
| SP5-02 | Story | UI hiển thị preview và trạng thái render | Người dùng thấy kết quả đầu tiên |
| SP5-03 | Story | Lưu asset render vào storage | Quản lý được artifact đầu ra |

### Sprint 6

| ID | Loại | Mô tả | Kết quả đầu ra |
| --- | --- | --- | --- |
| SP6-01 | Story | Export ảnh/tài liệu đầu ra | Có thể tải xuống |
| SP6-02 | Story | Regenerate infographic | Tạo version khác |
| SP6-03 | Story | Hoàn thiện luồng MVP end-to-end | Demo hoàn chỉnh |

### Sprint 7

| ID | Loại | Mô tả | Kết quả đầu ra |
| --- | --- | --- | --- |
| SP7-01 | Story | Retry và idempotency theo từng bước pipeline | Giảm lỗi ngẫu nhiên |
| SP7-02 | Story | Logging/metrics/dashboard cơ bản | Có observability |
| SP7-03 | Story | Error handling và cảnh báo vận hành | Hệ thống ổn định hơn |

### Sprint 8

| ID | Loại | Mô tả | Kết quả đầu ra |
| --- | --- | --- | --- |
| SP8-01 | Story | Tối ưu prompt và chất lượng đầu ra | Nâng chất lượng infographic |
| SP8-02 | Story | Tối ưu chi phí xử lý và cache | Giảm cost per run |
| SP8-03 | Story | Chuẩn bị alpha/beta release checklist | Sẵn sàng pre-release |

## 3. Nguyên tắc refinement

- Story phải gắn rõ acceptance criteria hoặc kết quả có thể kiểm chứng.
- Story đi vào sprint phải đủ nhỏ để hoàn thành trong một sprint 2 tuần.
- Task kỹ thuật cần phản ánh dependency liên quan đến contract, schema, queue, storage hoặc AI provider.
- Mỗi sprint review phải cập nhật lại độ ưu tiên backlog dựa trên học được từ sprint trước.
