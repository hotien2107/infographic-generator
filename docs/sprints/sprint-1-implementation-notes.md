# Sprint 1 Implementation Notes

## Các điểm đã siết lại sau khi rà soát implementation

- Backend decode JSON theo chế độ strict (`DisallowUnknownFields`) để không chấp nhận field thừa.
- Upload multipart chỉ cho phép `file` và `original_filename`, từ chối field ngoài contract.
- Validation thống nhất cho `title`, `input_mode`, `projectId`, file type, file size, empty file.
- Response envelope được giữ đồng nhất cho cả success/error.
- Contract Sprint 1 được cập nhật để phản ánh chính xác metadata/status đang được trả về ở implementation hiện tại.
- Frontend đã tách API layer và hiển thị message lỗi trực tiếp từ backend để giảm hardcode ở UI.
