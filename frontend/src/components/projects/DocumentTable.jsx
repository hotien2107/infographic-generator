import { Pencil, Trash2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { formatBytes, formatDate } from '@/lib/format'

import { StatusBadge } from './StatusBadge'

export function DocumentTable({ documents, onRename, onDelete }) {
  return (
    <div className="overflow-hidden rounded-3xl border border-primary/15 bg-white/95">
      <div className="space-y-3 p-3 md:hidden">
        {documents.map((document) => (
          <div key={document.id} className="rounded-2xl border border-primary/10 bg-gradient-to-br from-white to-fuchsia-50/40 p-4 shadow-sm">
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="font-semibold text-slate-900">{document.filename}</p>
                <p className="mt-1 text-xs text-muted-foreground">{document.error_message || document.extracted_text_preview || 'Sẵn sàng cho bước tiếp theo.'}</p>
              </div>
              <StatusBadge status={document.status} type="document" />
            </div>
            <div className="mt-3 grid grid-cols-2 gap-2 text-xs text-slate-600">
              <p>Định dạng: {document.mime_type}</p>
              <p>Kích thước: {formatBytes(document.size_bytes)}</p>
              <p>Tạo: {formatDate(document.created_at)}</p>
              <p>Cập nhật: {formatDate(document.updated_at)}</p>
            </div>
            <div className="mt-3 flex flex-wrap gap-2">
              <Button variant="outline" size="sm" onClick={() => onRename(document)}>
                <Pencil className="h-4 w-4" />Sửa
              </Button>
              <Button variant="outline" size="sm" onClick={() => onDelete(document)}>
                <Trash2 className="h-4 w-4" />Xóa
              </Button>
            </div>
          </div>
        ))}
      </div>

      <div className="hidden overflow-x-auto md:block">
        <table className="min-w-full divide-y divide-border text-left text-sm">
          <thead className="bg-gradient-to-r from-fuchsia-100/80 via-rose-100/80 to-amber-100/80 text-slate-700">
            <tr>
              <th className="px-5 py-4 font-semibold">Tài liệu</th>
              <th className="px-5 py-4 font-semibold">Định dạng</th>
              <th className="px-5 py-4 font-semibold">Kích thước</th>
              <th className="px-5 py-4 font-semibold">Trạng thái</th>
              <th className="px-5 py-4 font-semibold">Ngày tạo</th>
              <th className="px-5 py-4 font-semibold">Cập nhật gần nhất</th>
              <th className="px-5 py-4 font-semibold text-right">Thao tác</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/80">
            {documents.map((document) => (
              <tr key={document.id} className="hover:bg-fuchsia-50/40">
                <td className="px-5 py-4 align-top">
                  <div>
                    <p className="font-medium text-slate-900">{document.filename}</p>
                    <p className="mt-1 text-xs text-muted-foreground">{document.error_message || document.extracted_text_preview || 'Sẵn sàng cho bước tiếp theo.'}</p>
                  </div>
                </td>
                <td className="px-5 py-4 align-top text-slate-700">{document.mime_type}</td>
                <td className="px-5 py-4 align-top text-slate-700">{formatBytes(document.size_bytes)}</td>
                <td className="px-5 py-4 align-top">
                  <StatusBadge status={document.status} type="document" />
                </td>
                <td className="px-5 py-4 align-top text-slate-700">{formatDate(document.created_at)}</td>
                <td className="px-5 py-4 align-top text-slate-700">{formatDate(document.updated_at)}</td>
                <td className="px-5 py-4 align-top">
                  <div className="flex justify-end gap-2">
                    <Button variant="outline" onClick={() => onRename(document)}>
                      <Pencil className="h-4 w-4" />
                      Chỉnh sửa
                    </Button>
                    <Button variant="outline" onClick={() => onDelete(document)}>
                      <Trash2 className="h-4 w-4" />
                      Xóa
                    </Button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
