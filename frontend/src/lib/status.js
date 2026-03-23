export const projectStatusLabels = {
  draft: 'Draft',
  uploaded: 'Uploaded',
  processing: 'Processing',
  processed: 'Processed',
  failed: 'Failed',
}

export const documentStatusLabels = {
  uploaded: 'Uploaded',
  queued: 'Queued',
  processing: 'Processing',
  processed: 'Processed',
  failed: 'Failed',
}

export const stepLabels = {
  waiting_for_upload: 'Waiting for upload',
  uploaded: 'Uploaded',
  queued_for_processing: 'Queued for processing',
  extracting: 'Extracting',
  ready_for_generation: 'Ready for generation',
  failed: 'Failed',
}

export function badgeVariantForStatus(status) {
  if (status === 'processed') return 'success'
  if (status === 'failed') return 'destructive'
  if (status === 'processing' || status === 'queued') return 'warning'
  return 'outline'
}
