import { ArrowLeft, FilePlus2, Pencil } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from '@/router'

import { ConfirmDialog } from '@/components/common/ConfirmDialog'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { LoadingState } from '@/components/common/LoadingState'
import { PageHeader } from '@/components/common/PageHeader'
import { DocumentNameModal } from '@/components/projects/DocumentNameModal'
import { DocumentTable } from '@/components/projects/DocumentTable'
import { DocumentUploadModal } from '@/components/projects/DocumentUploadModal'
import { ProjectFormModal } from '@/components/projects/ProjectFormModal'
import { StatusBadge } from '@/components/projects/StatusBadge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { projectApi } from '@/lib/api'
import { formatDate } from '@/lib/format'
import { toErrorMessage } from '@/lib/http'

export function ProjectDetailPage() {
  const { projectId } = useParams('/projects/:projectId')
  const navigate = useNavigate()
  const [detail, setDetail] = useState(null)
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState(null)
  const [isProjectModalOpen, setIsProjectModalOpen] = useState(false)
  const [isDocumentModalOpen, setIsDocumentModalOpen] = useState(false)
  const [documentToRename, setDocumentToRename] = useState(null)
  const [documentToDelete, setDocumentToDelete] = useState(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [rawTextInput, setRawTextInput] = useState('')
  const [notice, setNotice] = useState(null)

  const project = detail?.project ?? null
  const documents = useMemo(() => detail?.documents ?? [], [detail])

  async function loadProject() {
    if (!projectId) return
    setIsLoading(true)
    setErrorMessage(null)
    try {
      setDetail(await projectApi.getProject(projectId))
    } catch (error) {
      setErrorMessage(toErrorMessage(error, 'Không thể tải chi tiết dự án.'))
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadProject()
  }, [projectId])

  async function handleProjectSubmit(values) {
    if (!project) return
    setIsSubmitting(true)
    try {
      await projectApi.updateProject(project.id, values)
      setIsProjectModalOpen(false)
      await loadProject()
    } catch (error) {
      setErrorMessage(toErrorMessage(error, 'Không thể cập nhật dự án.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleUpload(file) {
    if (!project) return
    setIsSubmitting(true)
    try {
      await projectApi.uploadDocument(project.id, file)
      setIsDocumentModalOpen(false)
      await loadProject()
    } catch (error) {
      setErrorMessage(toErrorMessage(error, 'Không thể thêm tài liệu.'))
    } finally {
      setIsSubmitting(false)
    }
  }


  async function handleSubmitText() {
    if (!project || !rawTextInput.trim()) return
    setIsSubmitting(true)
    setNotice(null)
    try {
      await projectApi.submitText(project.id, rawTextInput.trim())
      setNotice('Đã gửi nội dung, hệ thống đang trích xuất.')
      setRawTextInput('')
      await loadProject()
    } catch (error) {
      setErrorMessage(toErrorMessage(error, 'Không thể gửi nội dung text.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleRenameDocument(filename) {
    if (!project || !documentToRename) return
    setIsSubmitting(true)
    try {
      await projectApi.updateDocument(project.id, documentToRename.id, { filename })
      setDocumentToRename(null)
      await loadProject()
    } catch (error) {
      setErrorMessage(toErrorMessage(error, 'Không thể cập nhật tài liệu.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleDeleteDocument() {
    if (!project || !documentToDelete) return
    setIsDeleting(true)
    try {
      await projectApi.deleteDocument(project.id, documentToDelete.id)
      setDocumentToDelete(null)
      await loadProject()
    } catch (error) {
      setErrorMessage(toErrorMessage(error, 'Không thể xóa tài liệu.'))
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Chi tiết dự án"
        title={project?.title ?? 'Đang tải dự án'}
        description="Xem thông tin tổng quan, thêm tài liệu mới hoặc cập nhật từng tài liệu đang có trong dự án."
        actions={
          <>
            <Button variant="outline" asChild>
              <Link to="/projects">
                <ArrowLeft className="h-4 w-4" />
                Quay lại
              </Link>
            </Button>
            {project ? (
              <Button variant="outline" onClick={() => setIsProjectModalOpen(true)}>
                <Pencil className="h-4 w-4" />
                Chỉnh sửa dự án
              </Button>
            ) : null}
            {project && project.input_mode === 'file' ? (
              <Button onClick={() => setIsDocumentModalOpen(true)}>
                <FilePlus2 className="h-4 w-4" />
                Thêm tài liệu
              </Button>
            ) : null}
          </>
        }
      />

      {isLoading ? <LoadingState label="Đang tải chi tiết dự án..." /> : null}
      {!isLoading && errorMessage ? <ErrorState description={errorMessage} onRetry={loadProject} /> : null}
      {!isLoading && !errorMessage && !project ? (
        <EmptyState
          title="Không tìm thấy dự án"
          description="Dự án bạn đang tìm có thể đã bị xóa hoặc không còn tồn tại."
          action={<Button onClick={() => navigate('/projects')}>Về danh sách dự án</Button>}
        />
      ) : null}

      {!isLoading && !errorMessage && project ? (
        <>
          {notice ? <p className="rounded-xl bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{notice}</p> : null}
          <section className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
            <Card className="rounded-3xl border-white/70 bg-white/90">
              <CardContent className="space-y-5 p-6">
                <div className="flex flex-wrap items-center gap-3">
                  <StatusBadge status={project.status} />
                  <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600">
                    {project.input_mode === 'file' ? 'Tài liệu tải lên' : 'Nội dung nhập tay'}
                  </span>
                </div>
                <div>
                  <h3 className="text-xl font-semibold text-slate-950">Thông tin tổng quan</h3>
                  <p className="mt-2 text-sm leading-6 text-muted-foreground">{project.description || 'Chưa có mô tả cho dự án này.'}</p>
                </div>
                <dl className="grid gap-4 sm:grid-cols-2">
                  <div className="rounded-2xl bg-slate-50 p-4">
                    <dt className="text-sm text-muted-foreground">Ngày tạo</dt>
                    <dd className="mt-1 font-medium text-slate-900">{formatDate(project.created_at)}</dd>
                  </div>
                  <div className="rounded-2xl bg-slate-50 p-4">
                    <dt className="text-sm text-muted-foreground">Cập nhật gần nhất</dt>
                    <dd className="mt-1 font-medium text-slate-900">{formatDate(project.updated_at)}</dd>
                  </div>
                </dl>
              </CardContent>
            </Card>

            <Card className="rounded-3xl border-white/70 bg-white/90">
              <CardContent className="grid gap-4 p-6 sm:grid-cols-2 xl:grid-cols-1">
                {[
                  { label: 'Tổng tài liệu', value: project.processing_summary.total_documents },
                  { label: 'Đã tải lên', value: project.processing_summary.uploaded_documents },
                  { label: 'Đang xử lý', value: project.processing_summary.extracting_documents },
                  { label: 'Hoàn tất', value: project.processing_summary.extracted_documents },
                ].map((item) => (
                  <div key={item.label} className="rounded-2xl bg-slate-50 p-4">
                    <p className="text-sm text-muted-foreground">{item.label}</p>
                    <p className="mt-1 text-3xl font-semibold text-slate-950">{item.value}</p>
                  </div>
                ))}
              </CardContent>
            </Card>
          </section>


          {project.input_mode === 'text' ? (
            <section className="space-y-3">
              <h3 className="text-xl font-semibold text-slate-950">Nhập nội dung trực tiếp</h3>
              <p className="text-sm text-muted-foreground">Dán nội dung thô để hệ thống trích xuất theo pipeline.</p>
              <Textarea value={rawTextInput} onChange={(event) => setRawTextInput(event.target.value)} rows={8} placeholder="Nhập hoặc dán nội dung..." />
              <div>
                <Button onClick={handleSubmitText} disabled={isSubmitting || rawTextInput.trim().length < 10}>
                  {isSubmitting ? 'Đang gửi...' : 'Gửi nội dung để trích xuất'}
                </Button>
              </div>
            </section>
          ) : null}

          <section className="space-y-4">
            <div className="flex items-center justify-between gap-3">
              <div>
                <h3 className="text-xl font-semibold text-slate-950">Danh sách tài liệu</h3>
                <p className="mt-1 text-sm text-muted-foreground">Quản lý các tài liệu đang thuộc dự án này.</p>
              </div>
            </div>

            {documents.length === 0 ? (
              <EmptyState
                title="Dự án chưa có tài liệu"
                description="Thêm tài liệu hoặc nhập text để bắt đầu chuẩn bị nội dung cho infographic của bạn."
                action={<Button onClick={() => setIsDocumentModalOpen(true)}>Thêm tài liệu</Button>}
              />
            ) : (
              <DocumentTable documents={documents} onRename={setDocumentToRename} onDelete={setDocumentToDelete} />
            )}
          </section>
        </>
      ) : null}

      <ProjectFormModal
        open={isProjectModalOpen}
        mode="edit"
        project={project}
        isSubmitting={isSubmitting}
        onClose={() => setIsProjectModalOpen(false)}
        onSubmit={handleProjectSubmit}
      />
      <DocumentUploadModal open={isDocumentModalOpen} isSubmitting={isSubmitting} onClose={() => setIsDocumentModalOpen(false)} onSubmit={handleUpload} />
      <DocumentNameModal open={Boolean(documentToRename)} document={documentToRename} isSubmitting={isSubmitting} onClose={() => setDocumentToRename(null)} onSubmit={handleRenameDocument} />
      <ConfirmDialog
        open={Boolean(documentToDelete)}
        title="Xóa tài liệu"
        description={`Bạn có chắc muốn xóa tài liệu “${documentToDelete?.filename ?? ''}”?`}
        confirmLabel="Xóa tài liệu"
        isSubmitting={isDeleting}
        onClose={() => setDocumentToDelete(null)}
        onConfirm={handleDeleteDocument}
      />
    </div>
  )
}
