import { useEffect, useState } from 'react'
import { Link } from '@/router'

import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { LoadingState } from '@/components/common/LoadingState'
import { PageHeader } from '@/components/common/PageHeader'
import { StatCard } from '@/components/dashboard/StatCard'
import { StatusBadge } from '@/components/projects/StatusBadge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { dashboardApi, projectApi } from '@/lib/api'
import { formatDate } from '@/lib/format'
import { toErrorMessage } from '@/lib/http'

export function DashboardPage() {
  const [summary, setSummary] = useState(null)
  const [projects, setProjects] = useState([])
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState(null)

  async function loadData() {
    setIsLoading(true)
    setErrorMessage(null)
    try {
      const [summaryData, projectList] = await Promise.all([dashboardApi.getSummary(), projectApi.listProjects()])
      setSummary(summaryData)
      setProjects(projectList.slice(0, 5))
    } catch (error) {
      setErrorMessage(toErrorMessage(error, 'Không thể tải trang tổng quan.'))
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Tổng quan"
        title="Theo dõi tiến độ xử lý nội dung"
        description="Nắm nhanh số lượng dự án, tài liệu và những mục cần ưu tiên để đội ngũ xử lý infographic hiệu quả hơn."
        actions={
          <Button asChild>
            <Link to="/projects">Mở danh sách dự án</Link>
          </Button>
        }
      />

      {isLoading ? <LoadingState label="Đang tải số liệu tổng quan..." /> : null}
      {!isLoading && errorMessage ? <ErrorState description={errorMessage} onRetry={loadData} /> : null}
      {!isLoading && !errorMessage && summary ? (
        <>
          <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            <StatCard title="Tổng số dự án" value={summary.total_projects} hint="Tất cả dự án đang được quản lý trên hệ thống." />
            <StatCard title="Tổng số tài liệu" value={summary.total_documents} hint="Bao gồm toàn bộ tài liệu đã được thêm vào các dự án." />
            <StatCard title="Dự án đang xử lý" value={summary.processing_projects} hint="Các dự án đang có nội dung được xử lý." />
            <StatCard title="Dự án hoàn tất" value={summary.completed_projects} hint="Các dự án đã sẵn sàng cho bước tiếp theo." />
            <StatCard title="Dự án cần chú ý" value={summary.attention_projects} hint="Những dự án đang có lỗi hoặc cần kiểm tra thêm." />
            <StatCard title="Bản nháp" value={summary.draft_projects} hint="Các dự án mới tạo nhưng chưa có tài liệu nào." />
          </section>

          <section>
            <Card className="rounded-3xl border-white/70 bg-white/90">
              <CardHeader className="flex flex-row items-center justify-between gap-3">
                <div>
                  <CardTitle className="text-xl">Dự án cập nhật gần đây</CardTitle>
                  <p className="mt-1 text-sm text-muted-foreground">Chọn nhanh một dự án để tiếp tục bổ sung tài liệu hoặc kiểm tra trạng thái.</p>
                </div>
                <Button variant="outline" asChild>
                  <Link to="/projects">Xem tất cả</Link>
                </Button>
              </CardHeader>
              <CardContent>
                {projects.length === 0 ? (
                  <EmptyState
                    title="Chưa có dự án nào"
                    description="Tạo dự án đầu tiên để bắt đầu tập hợp tài liệu và theo dõi tiến độ xử lý nội dung."
                    action={
                      <Button asChild>
                        <Link to="/projects">Tạo dự án</Link>
                      </Button>
                    }
                  />
                ) : (
                  <div className="space-y-3">
                    {projects.map((project) => (
                      <Link
                        key={project.id}
                        to={`/projects/${project.id}`}
                        className="flex flex-col gap-3 rounded-2xl border border-border px-4 py-4 transition-colors hover:bg-slate-50 md:flex-row md:items-center md:justify-between"
                      >
                        <div>
                          <p className="font-medium text-slate-900">{project.title}</p>
                          <p className="mt-1 text-sm text-muted-foreground">{project.description || 'Chưa có mô tả.'}</p>
                        </div>
                        <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
                          <StatusBadge status={project.status} />
                          <span>{project.document_count} tài liệu</span>
                          <span>Cập nhật {formatDate(project.updated_at)}</span>
                        </div>
                      </Link>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </section>
        </>
      ) : null}
    </div>
  )
}
