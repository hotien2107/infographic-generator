# Sprint 1 Board – Foundation & Project Setup

## 1. Sprint board

| ID | Loại | Công việc | Owner đề xuất | Trạng thái | Ghi chú |
| --- | --- | --- | --- | --- | --- |
| S1-B01 | Planning | Rà soát PRD, User Stories, Acceptance Criteria | PO / BA | Done | Đầu vào cho backlog và scope |
| S1-B02 | Planning | Tạo product backlog theo 8 sprint | PO / BA | Done | Dùng để planning liên tục |
| S1-B03 | Governance | Tạo Definition of Done | Whole team | Done | Là baseline chất lượng |
| S1-B04 | Analysis | Chuẩn hóa entity user/project/document | BA / Backend | Done | Đầu vào cho API và DB |
| S1-B05 | Contract | Chốt API contract Sprint 1 | Backend / Frontend | Done | `contracts/sprint-1-api.yaml` |
| S1-B06 | DevOps | Xác định baseline env, DB, storage, CI/CD | DevOps | Done | Ghi trong sprint plan |
| S1-B07 | QA | Tạo checklist test create project + upload | QA | Done | Ghi trong sprint plan |
| S1-B08 | Delivery | Tạo tài liệu sprint cho toàn bộ 8 sprint | PO / PM | Done | Phục vụ roadmap execution |

## 2. Cấu trúc công việc theo luồng

### To Do

- Không còn đầu việc P0 chưa được mô tả.

### In Progress

- Dùng board này như baseline; khi có code triển khai thực tế, cập nhật owner/ngày hoàn thành cho từng task.

### Done

- Backlog tổng.
- Definition of Done.
- Sprint 1 plan.
- Sprint plans cho Sprint 2-8.
- Checklist QA Sprint 1.
- Contract API Sprint 1.

## 3. Definition of Ready cho task đưa vào sprint

Một task chỉ nên được kéo vào sprint khi:

- Có mục tiêu rõ ràng và kết quả đầu ra cụ thể.
- Có owner chính chịu trách nhiệm.
- Có dependency đã được nhận diện.
- Có tiêu chí done tối thiểu hoặc checklist kiểm chứng.

## 4. Nhóm công việc tiếp nối sau Sprint 1

- Backend bắt đầu hiện thực hóa `POST /api/v1/projects`.
- Frontend dựng màn hình create project và upload file.
- DevOps chuẩn bị service local cho PostgreSQL/Redis/Object Storage.
- QA chuẩn bị dữ liệu kiểm thử mẫu cho Sprint 2.
