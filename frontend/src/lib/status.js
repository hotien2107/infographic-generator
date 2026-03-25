export const projectStatusLabels = {
  draft: 'Bản nháp',
  uploaded: 'Đã nhận dữ liệu',
  extracting: 'Đang trích xuất',
  extracted: 'Đã trích xuất',
  failed: 'Thất bại',
}

export const documentStatusLabels = {
  uploaded: 'Đã tải lên',
  extracting: 'Đang trích xuất',
  extracted: 'Đã trích xuất',
  failed: 'Thất bại',
}

export function badgeVariantForStatus(status) {
  if (status === 'extracted') return 'success'
  if (status === 'failed') return 'destructive'
  if (status === 'extracting') return 'warning'
  return 'outline'
}
