# Definition of Done

Definition of Done (DoD) này áp dụng cho toàn bộ dự án **AI Infographic Generator** nhằm đảm bảo mọi hạng mục được hoàn thành theo cùng một chuẩn chất lượng tối thiểu.

## 1. Yêu cầu nghiệp vụ

- Story đáp ứng đúng phạm vi đã thống nhất trong sprint plan.
- Acceptance criteria liên quan đã được kiểm tra và không còn điểm mơ hồ trọng yếu.
- Các ràng buộc nghiệp vụ, trạng thái và luồng lỗi đã được mô tả rõ trong tài liệu hoặc contract.

## 2. Yêu cầu kỹ thuật

- Mã nguồn hoặc tài liệu kỹ thuật được cập nhật đúng vị trí trong repository.
- Không tạo ra mâu thuẫn với kiến trúc hệ thống, API contract hoặc naming convention đang dùng.
- Cấu hình môi trường, biến môi trường và dependency mới được mô tả đầy đủ nếu có phát sinh.
- Các schema, enum trạng thái và giao diện trao đổi dữ liệu được đồng bộ giữa các thành phần liên quan.

## 3. Kiểm thử

- Có ít nhất một hình thức kiểm tra phù hợp: test tự động, contract validation, checklist QA hoặc test case thủ công.
- Luồng thành công và luồng lỗi chính đã được kiểm tra.
- Không còn lỗi blocker hoặc critical đã biết liên quan trực tiếp đến hạng mục bàn giao.

## 4. Tài liệu

- Tài liệu người dùng nội bộ hoặc tài liệu vận hành được cập nhật nếu thay đổi ảnh hưởng cách triển khai hoặc sử dụng.
- Các quyết định kỹ thuật quan trọng được lưu lại trong tài liệu sprint, backlog hoặc architecture note.
- Các dependency hoặc rủi ro tồn đọng được ghi nhận minh bạch.

## 5. Review và sẵn sàng bàn giao

- Hạng mục đã được review bởi ít nhất một thành viên liên quan (PO/BA/dev/QA tùy loại việc).
- Không còn TODO mơ hồ trong phần phạm vi đã commit là hoàn tất.
- Có thể demo được hoặc giải thích rõ kết quả đầu ra bằng artifact cụ thể.

## 6. Checklist nhanh trước khi đóng task

- [ ] Scope đúng với story/task đã giao
- [ ] Acceptance criteria liên quan đã được đối chiếu
- [ ] Contract/schema/trạng thái đã đồng bộ
- [ ] Có bằng chứng kiểm thử
- [ ] Tài liệu đã cập nhật
- [ ] Đã ghi nhận risk hoặc follow-up nếu còn tồn đọng
