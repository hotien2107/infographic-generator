export const projectStatusLabels = {
  draft: 'Bản nháp',
  uploaded: 'Đã thêm tài liệu',
  processing: 'Đang xử lý',
  processed: 'Hoàn tất',
  failed: 'Cần chú ý',
}

export const documentStatusLabels = {
  uploaded: 'Đã tải lên',
  queued: 'Đang chờ xử lý',
  processing: 'Đang xử lý',
  processed: 'Hoàn tất',
  failed: 'Cần kiểm tra',
}

export function badgeVariantForStatus(status) {
  if (status === 'processed') return 'success'
  if (status === 'failed') return 'destructive'
  if (status === 'processing' || status === 'queued') return 'warning'
  return 'outline'
}
