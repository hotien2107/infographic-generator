# AI Infographic Generator

AI Infographic Generator là dự án xây dựng một nền tảng web dùng trí tuệ nhân tạo để biến tài liệu thô, văn bản tự do hoặc dữ liệu có cấu trúc thành infographic hoàn chỉnh. Mục tiêu của hệ thống không chỉ là tạo ra hình ảnh đẹp mà còn phải hiểu đúng nội dung, trích xuất insight quan trọng, tổ chức thông tin hợp lý và hiển thị theo bố cục trực quan để người dùng không chuyên thiết kế vẫn có thể tạo sản phẩm có giá trị sử dụng thực tế.

## 1. Mục tiêu dự án

Dự án hướng tới việc cung cấp một quy trình end-to-end để người dùng:

- tạo project mới và quản lý vòng đời xử lý nội dung;
- tải tài liệu lên hoặc nhập văn bản trực tiếp;
- để hệ thống tự động đọc hiểu, tóm tắt và trích xuất thông tin quan trọng;
- chuyển nội dung đã hiểu thành infographic specification có cấu trúc;
- render bản xem trước infographic, tạo biến thể và tải kết quả xuống.

Ở mức MVP, sản phẩm tập trung vào việc chứng minh khả năng chuyển đổi tài liệu thành infographic trong thời gian ngắn, với trải nghiệm đơn giản, chi phí AI có thể kiểm soát và kiến trúc đủ tốt để mở rộng lên production.

## 2. Bài toán sản phẩm giải quyết

Việc tạo infographic hiện nay thường cần phối hợp giữa người viết nội dung, người phân tích dữ liệu và designer. Quá trình này tốn thời gian, phụ thuộc kỹ năng chuyên môn và khó mở rộng khi khối lượng nội dung lớn. Dự án giải quyết bài toán đó bằng cách tự động hóa ba nhóm công việc cốt lõi:

1. **Hiểu nội dung**: đọc tài liệu, xác định chủ đề, ý chính, số liệu, thực thể và ngữ cảnh.
2. **Tổ chức nội dung**: chia section, xây hierarchy, đề xuất loại biểu diễn phù hợp như timeline, bullet, chart hoặc so sánh.
3. **Trực quan hóa**: áp dụng layout, màu sắc, icon/hình ảnh và xuất thành infographic hoàn chỉnh.

Cách tiếp cận này giúp sản phẩm khác với các công cụ “text-to-image” thuần túy: hệ thống ưu tiên đúng nội dung trước khi tối ưu phần trình bày.

## 3. Đối tượng người dùng mục tiêu

Theo bộ tài liệu dự án, hệ thống nhắm đến nhiều nhóm người dùng có nhu cầu chuyển đổi thông tin sang dạng trực quan nhanh chóng:

- **Sinh viên**: tóm tắt tài liệu học tập dài thành nội dung dễ nhớ.
- **Marketer**: tạo nội dung truyền thông nhanh khi thiếu designer hoặc thời gian gấp.
- **Data analyst**: biến dữ liệu, insight và số liệu thành hình thức trình bày rõ ràng.
- **Founder / startup**: chuẩn bị tài liệu giới thiệu ý tưởng, báo cáo hoặc pitch deck rút gọn.

Điểm chung của các nhóm này là cần tốc độ, tính trực quan và giảm phụ thuộc vào kỹ năng thiết kế thủ công.

## 4. Phạm vi MVP

### In scope

- Tạo và quản lý project.
- Upload tài liệu định dạng PDF, DOCX, TXT hoặc nhập text trực tiếp.
- Validate đầu vào và lưu metadata tài liệu.
- Thực hiện pipeline xử lý gồm extraction, summarization, spec generation và rendering.
- Hiển thị trạng thái xử lý, preview kết quả, regenerate variation và download output.

### Out of scope trong giai đoạn đầu

- Chỉnh sửa layout nâng cao theo ý người dùng.
- Collaboration nhiều người dùng trên cùng thiết kế.
- Marketplace template.
- Brand customization ở mức sâu.

## 5. Luồng nghiệp vụ tổng quát

Luồng vận hành cốt lõi của hệ thống được mô tả xuyên suốt trong các tài liệu hiện có như sau:

1. Người dùng tạo project và cung cấp dữ liệu đầu vào.
2. Hệ thống thực hiện bước ingestion để nhận file/text và lưu trữ metadata.
3. Module extraction chuyển tài liệu thành dữ liệu văn bản có cấu trúc hơn.
4. AI understanding layer tóm tắt và trích xuất insight, số liệu, ngữ cảnh.
5. Structuring layer biến nội dung đã hiểu thành infographic specification độc lập với render.
6. Visualization/rendering layer chọn layout, phong cách và tạo preview infographic.
7. Delivery layer cho phép xem trước, tải xuống hoặc yêu cầu tạo lại phiên bản khác.

Thiết kế pipeline nhiều bước này giúp hệ thống dễ debug, dễ kiểm soát chất lượng từng lớp, dễ thay model và tối ưu chi phí AI.

## 6. Kiến trúc hệ thống đề xuất

Bộ tài liệu thiết kế mô tả kiến trúc nhiều lớp với các thành phần chính:

- **Web Client**: giao diện người dùng để tạo project, upload dữ liệu, theo dõi trạng thái và xem preview.
- **Backend API**: xử lý request, quản lý project, document, trạng thái job và cung cấp endpoint cho frontend.
- **Auth / Session service**: quản lý người dùng và phiên làm việc nếu có trong MVP.
- **Object Storage**: lưu file upload và output infographic.
- **Queue / Worker / Processing Orchestrator**: điều phối các bước xử lý bất đồng bộ.
- **Document Extractor**: đọc PDF/DOCX/TXT và chuẩn hóa nội dung đầu vào.
- **AI Text Service**: thực hiện summarization, extraction insight, spec generation.
- **AI Image Service / Rendering layer**: tạo hình ảnh hoặc thành phần đồ họa cho infographic.
- **PostgreSQL**: lưu project, document, trạng thái xử lý, metadata và dữ liệu nghiệp vụ.
- **Redis**: hỗ trợ queue, job state hoặc cache.
- **Observability stack**: logging, metrics, theo dõi retry, thời gian xử lý và chi phí.

## 7. Thiết kế AI pipeline

AI pipeline là phần cốt lõi của dự án và được tách thành các vai trò rõ ràng:

- **Reader**: đọc hiểu tài liệu, tóm tắt, nhận diện insight và các dữ liệu quan trọng.
- **Content editor**: tái cấu trúc nội dung thành section và hierarchy phù hợp với infographic.
- **Designer / rendering support**: ánh xạ specification sang layout, style và output cuối cùng.

Các nguyên tắc thiết kế pipeline được nhấn mạnh gồm:

- tách nhỏ tác vụ theo chức năng nhận thức;
- luôn sinh dữ liệu trung gian có schema rõ ràng để validate;
- đặt quality gate ở từng bước thay vì chỉ kiểm tra output cuối;
- ưu tiên output ổn định cho bước trích xuất/cấu trúc;
- tách biệt rõ “hiểu nội dung” và “trình bày nội dung”.

## 8. Yêu cầu chức năng nổi bật

Từ PRD, User Stories và Acceptance Criteria, có thể tóm tắt các yêu cầu chức năng chính như sau:

- upload file hợp lệ và xử lý lỗi đầu vào rõ ràng;
- nhập văn bản trực tiếp và xử lý tương tự tài liệu upload;
- tạo summary có cấu trúc, logic và không mâu thuẫn;
- trích xuất tiêu đề, ý chính, số liệu, mốc thời gian và insight quan trọng;
- sinh infographic với bố cục rõ ràng, thành phần đầy đủ và dễ hiểu;
- preview kết quả trước khi tải xuống;
- hỗ trợ regenerate để tạo phương án hiển thị khác.

## 9. Kế hoạch phát triển theo Agile

Tài liệu quy hoạch dự án đề xuất triển khai theo **Agile / Scrum nhẹ** với sprint dài 2 tuần và lộ trình ban đầu gồm 8 sprint:

1. **Foundation & Project Setup**.
2. **Document Ingestion & Extraction**.
3. **AI Understanding Layer**.
4. **Structuring Layer**.
5. **Rendering Layer**.
6. **Export & Regenerate**.
7. **Hardening & Observability**.
8. **Optimization & Pre-Release**.

Cách chia này phản ánh định hướng xây dựng từng lát cắt giá trị, kiểm chứng sớm giả định sản phẩm, đồng thời liên tục tối ưu chất lượng output AI, độ ổn định kỹ thuật và chi phí vận hành.

## 10. Tài liệu trong thư mục `docs/`

Toàn bộ tài liệu Markdown gốc của dự án đã được gom vào thư mục `docs/` để dễ quản lý. Nhóm tài liệu hiện tại bao gồm:

- `docs/1. Project Plan.md`: tổng quan, mục tiêu, tầm nhìn và phạm vi dự án.
- `docs/2. Product Strategy & Solution Overview.md`: chiến lược sản phẩm, mô hình vận hành và lựa chọn giải pháp.
- `docs/3. Product Requirement Document (PRD).md`: yêu cầu sản phẩm và phạm vi MVP.
- `docs/4. User Stories.md`: tập hợp user stories theo nhóm chức năng.
- `docs/5. Acceptance Criteria.md`: tiêu chí nghiệm thu cho các tính năng chính.
- `docs/6. System design.md`: kiến trúc hệ thống, thành phần và luồng dữ liệu.
- `docs/7. Agile development process.md`: quy trình phát triển Agile áp dụng cho dự án.
- `docs/8. Agile sprint plan.md`: kế hoạch triển khai theo sprint.
- `docs/9. AI pipeline design.md`: thiết kế AI pipeline và nguyên tắc kiểm soát chất lượng.
- `docs/10. AI implementation.md`: chỉ dẫn tổng thể để AI coding agent triển khai dự án end-to-end.

## 11. Định hướng triển khai tiếp theo

Nếu tiếp tục phát triển repository này thành mã nguồn thực tế, các bước nên ưu tiên gồm:

- khởi tạo monorepo hoặc tách frontend/backend rõ ràng;
- thiết kế schema dữ liệu cho user, project, document, job và artifact;
- dựng API cho tạo project, upload file, lấy trạng thái xử lý và download output;
- chuẩn hóa schema JSON cho summary và infographic spec;
- xây worker orchestration cho pipeline nhiều bước với retry/idempotency;
- dựng giao diện web MVP cho upload, preview và quản lý project;
- thêm logging, metrics, cost tracking và test cho từng lớp xử lý.

## 12. Cách sử dụng repository hiện tại

Ở trạng thái hiện tại, repository đóng vai trò **bộ tài liệu nền tảng** để phục vụ discovery, product planning, system design và hướng dẫn triển khai. README này là điểm vào chính để hiểu nhanh dự án; thư mục `docs/` chứa toàn bộ tài liệu chi tiết để nhóm sản phẩm, kỹ thuật hoặc AI coding agent tra cứu theo từng chủ đề.
