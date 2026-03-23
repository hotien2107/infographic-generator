import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { formatBytes, formatDate } from '@/lib/format'

import { StatusBadge } from './StatusBadge'

export function DocumentList({ documents = [] }) {
  return (
    <Card className="bg-white/95">
      <CardHeader>
        <CardTitle className="text-2xl">Danh sách tài liệu</CardTitle>
        <CardDescription>Hiển thị dữ liệu document ingestion + processing trả về từ backend.</CardDescription>
      </CardHeader>
      <CardContent>
        {documents.length ? (
          <div className="space-y-3">
            {documents.map((document) => (
              <div key={document.id} className="rounded-xl border border-border bg-slate-50 p-4">
                <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
                  <div className="space-y-2">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="font-medium text-foreground">{document.filename}</p>
                      <StatusBadge kind="document" value={document.status} />
                    </div>
                    <p className="text-sm text-muted-foreground">{document.mime_type} · {formatBytes(document.size_bytes)}</p>
                    <p className="text-sm text-muted-foreground">Storage key: <span className="break-all font-medium text-foreground">{document.storage_key}</span></p>
                  </div>
                  <div className="space-y-1 text-sm text-muted-foreground md:text-right">
                    <p>Created: <span className="font-medium text-foreground">{formatDate(document.created_at)}</span></p>
                    <p>Started: <span className="font-medium text-foreground">{formatDate(document.processing_started_at)}</span></p>
                    <p>Finished: <span className="font-medium text-foreground">{formatDate(document.processing_finished_at)}</span></p>
                  </div>
                </div>
                {document.extracted_text_preview ? (
                  <div className="mt-4 rounded-lg border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-900">
                    <p className="font-medium">Extracted preview</p>
                    <p className="mt-1">{document.extracted_text_preview}</p>
                  </div>
                ) : null}
                {document.error_message ? (
                  <div className="mt-4 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800">
                    <p className="font-medium">Processing error</p>
                    <p className="mt-1">{document.error_message}</p>
                  </div>
                ) : null}
              </div>
            ))}
          </div>
        ) : (
          <div className="rounded-xl border border-dashed border-border p-10 text-center text-sm text-muted-foreground">
            Chưa có tài liệu nào. Chọn file rồi upload để khởi động luồng ingestion.
          </div>
        )}
      </CardContent>
    </Card>
  )
}
