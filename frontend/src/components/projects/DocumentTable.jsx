import { Pencil, Trash2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { formatBytes, formatDate } from '@/lib/format'

import { StatusBadge } from './StatusBadge'

export function DocumentTable({ documents, onRename, onDelete }) {
  return (
    <div className="overflow-hidden rounded-3xl border border-border bg-white">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-border text-left text-sm">
          <thead className="bg-slate-50/90 text-slate-600">
            <tr>
              <th className="px-5 py-4 font-medium">Tài liệu</th>
              <th className="px-5 py-4 font-medium">Định dạng</th>
              <th className="px-5 py-4 font-medium">Kích thước</th>
              <th className="px-5 py-4 font-medium">Trạng thái</th>
              <th className="px-5 py-4 font-medium">Ngày tạo</th>
              <th className="px-5 py-4 font-medium">Cập nhật gần nhất</th>
              <th className="px-5 py-4 font-medium text-right">Thao tác</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/80">
            {documents.map((document) => (
              <tr key={document.id} className="hover:bg-slate-50/70">
                <td className="px-5 py-4 align-top">
                  <div>
                    <p className="font-medium text-slate-900">{document.filename}</p>
                    <p className="mt-1 text-xs text-muted-foreground">{document.error_message || document.extracted_text_preview || 'Sẵn sàng cho bước tiếp theo.'}</p>
                  </div>
                </td>
                <td className="px-5 py-4 align-top text-slate-700">{document.mime_type}</td>
                <td className="px-5 py-4 align-top text-slate-700">{formatBytes(document.size_bytes)}</td>
                <td className="px-5 py-4 align-top"><StatusBadge status={document.status} type="document" /></td>
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
