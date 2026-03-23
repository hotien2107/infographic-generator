import { Eye, Pencil, Trash2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { formatDate } from '@/lib/format'

import { StatusBadge } from './StatusBadge'

export function ProjectTable({ projects, onView, onEdit, onDelete }) {
  return (
    <div className="overflow-hidden rounded-3xl border border-border bg-white">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-border text-left text-sm">
          <thead className="bg-slate-50/90 text-slate-600">
            <tr>
              <th className="px-5 py-4 font-medium">Tên dự án</th>
              <th className="px-5 py-4 font-medium">Loại đầu vào</th>
              <th className="px-5 py-4 font-medium">Trạng thái</th>
              <th className="px-5 py-4 font-medium">Tài liệu</th>
              <th className="px-5 py-4 font-medium">Ngày tạo</th>
              <th className="px-5 py-4 font-medium">Cập nhật gần nhất</th>
              <th className="px-5 py-4 font-medium text-right">Thao tác</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/80">
            {projects.map((project) => (
              <tr key={project.id} className="hover:bg-slate-50/70">
                <td className="px-5 py-4 align-top">
                  <div>
                    <p className="font-medium text-slate-900">{project.title}</p>
                    <p className="mt-1 max-w-sm text-xs leading-5 text-muted-foreground">{project.description || 'Chưa có mô tả.'}</p>
                  </div>
                </td>
                <td className="px-5 py-4 align-top capitalize text-slate-700">{project.input_mode === 'file' ? 'Tài liệu tải lên' : 'Nội dung nhập tay'}</td>
                <td className="px-5 py-4 align-top"><StatusBadge status={project.status} /></td>
                <td className="px-5 py-4 align-top text-slate-700">{project.document_count}</td>
                <td className="px-5 py-4 align-top text-slate-700">{formatDate(project.created_at)}</td>
                <td className="px-5 py-4 align-top text-slate-700">{formatDate(project.updated_at)}</td>
                <td className="px-5 py-4 align-top">
                  <div className="flex justify-end gap-2">
                    <Button variant="ghost" onClick={() => onView(project)}>
                      <Eye className="h-4 w-4" />
                      Xem chi tiết
                    </Button>
                    <Button variant="outline" onClick={() => onEdit(project)}>
                      <Pencil className="h-4 w-4" />
                      Chỉnh sửa
                    </Button>
                    <Button variant="outline" onClick={() => onDelete(project)}>
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
