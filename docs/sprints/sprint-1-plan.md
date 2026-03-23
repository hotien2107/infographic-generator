# Sprint 1 Plan – Foundation & Project Setup

## 1. Mục tiêu sprint

Thiết lập nền tảng kỹ thuật và quy trình làm việc tối thiểu để đội dự án có thể bắt đầu xây dựng MVP theo một contract thống nhất, có backlog rõ ràng và có khả năng triển khai các luồng cơ bản: tạo project, upload tài liệu và lưu metadata.

## 2. Sprint goal

> Đến cuối Sprint 1, team phải có bộ khung triển khai ban đầu gồm contract nền tảng, backlog/board điều hành, Definition of Done, thiết kế dữ liệu ban đầu và checklist QA cho các luồng create project + upload document.

## 3. Phạm vi cam kết

### 3.1 In-scope

- Rà soát tài liệu sản phẩm và chuyển thành backlog có cấu trúc.
- Chuẩn hóa DoD dùng chung cho team.
- Hoàn thiện sprint board với task theo vai trò.
- Xác định dữ liệu lõi cho user/project/document và các trạng thái đầu tiên.
- Chuẩn hóa contract API cho Sprint 1.
- Xây dựng hướng dẫn thực thi Sprint 1 cho PO/BA/Backend/Frontend/DevOps/QA.
- Chuẩn bị test case/checklist cho luồng tạo project và upload file.

### 3.2 Out-of-scope

- Chưa triển khai extraction worker.
- Chưa tích hợp AI text service hay rendering.
- Chưa cam kết auth production-ready.
- Chưa tối ưu hiệu năng production.

## 4. Dependency đầu vào

- PRD, User Stories, Acceptance Criteria đã được thống nhất ở mức MVP.
- System Design cung cấp hướng kiến trúc nhiều bước.
- API contract Sprint 1 là nguồn tham chiếu chung cho frontend và backend.

## 5. Deliverable bắt buộc

| Mã | Deliverable | Mô tả |
| --- | --- | --- |
| D1 | Product backlog | Backlog phân rã epic/feature/story/task cho 8 sprint |
| D2 | Definition of Done | Chuẩn chất lượng dùng chung cho mọi vai trò |
| D3 | Sprint 1 board | Danh sách task thực thi, owner đề xuất, ưu tiên và trạng thái |
| D4 | API contract Sprint 1 | Contract cho create project, get project, upload document |
| D5 | Sprint 1 QA checklist | Checklist kiểm tra happy path và error path |
| D6 | Hướng dẫn môi trường nền tảng | Danh sách env, CI/CD, database, storage cần có để bắt đầu build |

## 6. Kế hoạch công việc theo vai trò

### Product / BA

- Rà soát lại PRD, User Stories, Acceptance Criteria.
- Mapping các story vào backlog và ưu tiên theo sprint.
- Khóa sprint goal, in-scope/out-of-scope và dependency.
- Làm rõ các enum trạng thái nghiệp vụ cho project/document.

### Backend

- Xác định entity ban đầu: user, project, document.
- Chuẩn hóa workflow trạng thái cho project ở Sprint 1.
- Thống nhất request/response schema với frontend.
- Ghi rõ validation rule cho upload file.

### Frontend

- Chốt dữ liệu cần cho màn hình create project, upload file và project detail.
- Mapping UI state với enum trạng thái trong contract.
- Chuẩn bị plan tích hợp API trong sprint kế tiếp nếu code chưa sẵn sàng.

### DevOps

- Xác định biến môi trường nền tảng.
- Liệt kê yêu cầu cho local development, build pipeline, database và object storage.
- Chuẩn bị tiêu chí để CI kiểm tra contract/tài liệu khi repo bắt đầu có code.

### QA

- Tạo checklist test cho create project.
- Tạo checklist test cho upload file hợp lệ/không hợp lệ.
- Xác định dữ liệu mẫu cần dùng trong sprint kế tiếp.

## 7. Danh sách task Sprint 1

| ID | Vai trò | Task | Ưu tiên | Estimate |
| --- | --- | --- | --- | --- |
| S1-T01 | PO/BA | Tổng hợp backlog từ PRD, User Stories, AC | P0 | 2 SP |
| S1-T02 | PO/BA | Hoàn thiện Sprint 1 plan và sprint board | P0 | 2 SP |
| S1-T03 | BA/Tech Lead | Chuẩn hóa trạng thái project/document ban đầu | P0 | 1 SP |
| S1-T04 | Backend | Mô tả schema logic user/project/document | P0 | 2 SP |
| S1-T05 | Backend/Frontend | Chốt API contract cho create project/upload/get detail | P0 | 3 SP |
| S1-T06 | DevOps | Liệt kê env, CI/CD, DB, storage baseline | P1 | 2 SP |
| S1-T07 | QA | Viết checklist/test case Sprint 1 | P0 | 2 SP |
| S1-T08 | Whole team | Review DoD và thống nhất tiêu chí Done | P0 | 1 SP |

## 8. Rủi ro và hướng xử lý

| Rủi ro | Tác động | Giảm thiểu |
| --- | --- | --- |
| Contract API thay đổi liên tục | FE/BE lệch nhau, chậm tích hợp | Chốt contract sớm và version hóa |
| Chưa rõ validation file | Dễ phát sinh lỗi upload | Ghi rõ file type, dung lượng, error code |
| Thiếu thống nhất trạng thái project | UI và backend khó đồng bộ | Dùng enum duy nhất trong contract và backlog |
| Scope Sprint 1 bị kéo sang extraction | Trễ nền tảng, giảm chất lượng planning | Khóa out-of-scope và review change request |

## 9. Sprint 1 QA checklist

### Create project

- [ ] Tạo project với `title` hợp lệ và `input_mode=file` thành công.
- [ ] Tạo project với `input_mode=text` thành công.
- [ ] Từ chối `title` dưới 3 ký tự.
- [ ] Từ chối request thiếu `input_mode`.
- [ ] Trả về đúng trạng thái ban đầu `draft` và `waiting_for_upload`.

### Upload document

- [ ] Upload file hợp lệ vào project tồn tại trả về `202`.
- [ ] Sau upload thành công, project/document phản ánh trạng thái `uploaded`.
- [ ] Từ chối file sai định dạng với mã lỗi phù hợp.
- [ ] Từ chối file vượt dung lượng cho phép.
- [ ] Trả về `404` khi project không tồn tại.

## 10. Hướng dẫn môi trường nền tảng

### 10.1 Biến môi trường dự kiến

- `APP_ENV`
- `API_PORT`
- `DATABASE_URL`
- `REDIS_URL`
- `OBJECT_STORAGE_ENDPOINT`
- `OBJECT_STORAGE_BUCKET`
- `OBJECT_STORAGE_ACCESS_KEY`
- `OBJECT_STORAGE_SECRET_KEY`
- `MAX_UPLOAD_SIZE_MB`
- `ALLOWED_FILE_TYPES`

### 10.2 Baseline hạ tầng

- Database: PostgreSQL cho project/document metadata.
- Queue/cache: Redis cho các sprint sau, có thể khai báo sẵn từ Sprint 1.
- Object storage: S3-compatible storage cho file upload.
- CI/CD: tối thiểu có bước validate tài liệu, contract và build khi code được thêm vào.

## 11. Exit criteria

Sprint 1 được xem là hoàn thành khi:

- Backlog 8 sprint đã được tạo và đủ dùng cho planning.
- Có DoD dùng chung và được team thống nhất.
- Có Sprint 1 board với task/action rõ ràng.
- Sprint 1 API contract phản ánh đúng create project/upload/get project.
- Có QA checklist cho happy path và error path chính.
- Có hướng dẫn nền tảng để bắt đầu triển khai code ở sprint sau.
