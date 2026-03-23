export function toErrorMessage(error, fallback = 'Không thể hoàn tất yêu cầu.') {
  return error instanceof Error ? error.message : fallback
}
