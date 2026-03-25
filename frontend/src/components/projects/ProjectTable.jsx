import { Eye, Pencil, Trash2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { formatDate } from '@/lib/format'

import { StatusBadge } from './StatusBadge'

export function ProjectTable({ projects, onView, onEdit, onDelete }) {
  return (
    <div className="overflow-hidden rounded-3xl border border-primary/15 bg-white/95">
      <div className="space-y-3 p-3 md:hidden">
        {projects.map((project) => (
          <div key={project.id} className="rounded-2xl border border-primary/10 bg-gradient-to-br from-white to-indigo-50/50 p-4 shadow-sm">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="font-semibold text-slate-900">{project.title}</p>
                <p className="mt-1 text-xs leading-5 text-muted-foreground">{project.description || 'Chưa có mô tả.'}</p>
              </div>
              <StatusBadge status={project.status} />
            </div>
            <div className="mt-3 grid grid-cols-2 gap-2 text-xs text-slate-600">
              <p>Loại: {project.input_mode === 'file' ? 'Tài liệu tải lên' : 'Nội dung nhập tay'}</p>
              <p>Tài liệu: {project.document_count}</p>
              <p>Tạo: {formatDate(project.created_at)}</p>
              <p>Cập nhật: {formatDate(project.updated_at)}</p>
            </div>
            <div className="mt-3 flex flex-wrap gap-2">
              <Button variant="ghost" size="sm" onClick={() => onView(project)}>
                <Eye className="h-4 w-4" />Xem
              </Button>
              <Button variant="outline" size="sm" onClick={() => onEdit(project)}>
                <Pencil className="h-4 w-4" />Sửa
              </Button>
              <Button variant="outline" size="sm" onClick={() => onDelete(project)}>
                <Trash2 className="h-4 w-4" />Xóa
              </Button>
            </div>
          </div>
        ))}
      </div>

      <div className="hidden overflow-x-auto md:block">
        <table className="min-w-full divide-y divide-border text-left text-sm">
          <thead className="bg-gradient-to-r from-indigo-100/80 via-violet-100/80 to-sky-100/80 text-slate-700">
            <tr>
              <th className="px-5 py-4 font-semibold">Tên dự án</th>
              <th className="px-5 py-4 font-semibold">Loại đầu vào</th>
              <th className="px-5 py-4 font-semibold">Trạng thái</th>
              <th className="px-5 py-4 font-semibold">Tài liệu</th>
              <th className="px-5 py-4 font-semibold">Ngày tạo</th>
              <th className="px-5 py-4 font-semibold">Cập nhật gần nhất</th>
              <th className="px-5 py-4 font-semibold text-right">Thao tác</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/80">
            {projects.map((project) => (
              <tr key={project.id} className="hover:bg-indigo-50/50">
                <td className="px-5 py-4 align-top">
                  <div>
                    <p className="font-medium text-slate-900">{project.title}</p>
                    <p className="mt-1 max-w-sm text-xs leading-5 text-muted-foreground">{project.description || 'Chưa có mô tả.'}</p>
                  </div>
                </td>
                <td className="px-5 py-4 align-top capitalize text-slate-700">{project.input_mode === 'file' ? 'Tài liệu tải lên' : 'Nội dung nhập tay'}</td>
                <td className="px-5 py-4 align-top">
                  <StatusBadge status={project.status} />
                </td>
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
