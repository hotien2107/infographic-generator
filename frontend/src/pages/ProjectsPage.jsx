import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from '@/router'

import { ConfirmDialog } from '@/components/common/ConfirmDialog'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { LoadingState } from '@/components/common/LoadingState'
import { PageHeader } from '@/components/common/PageHeader'
import { ProjectFormModal } from '@/components/projects/ProjectFormModal'
import { ProjectTable } from '@/components/projects/ProjectTable'
import { Button } from '@/components/ui/button'
import { projectApi } from '@/lib/api'
import { toErrorMessage } from '@/lib/http'

export function ProjectsPage() {
  const navigate = useNavigate()
  const [projects, setProjects] = useState([])
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState(null)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [editingProject, setEditingProject] = useState(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [projectToDelete, setProjectToDelete] = useState(null)
  const [isDeleting, setIsDeleting] = useState(false)

  const isEmpty = useMemo(() => !isLoading && !errorMessage && projects.length === 0, [errorMessage, isLoading, projects.length])

  async function loadProjects() {
    setIsLoading(true)
    setErrorMessage(null)
    try {
      setProjects(await projectApi.listProjects())
    } catch (error) {
      setErrorMessage(toErrorMessage(error, 'Không thể tải danh sách dự án.'))
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadProjects()
  }, [])

  async function handleSubmit(values) {
    setIsSubmitting(true)
    try {
      if (editingProject) {
        await projectApi.updateProject(editingProject.id, values)
      } else {
        await projectApi.createProject(values)
      }
      setIsFormOpen(false)
      setEditingProject(null)
      await loadProjects()
    } catch (error) {
      setErrorMessage(toErrorMessage(error, 'Không thể lưu dự án.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleDelete() {
    if (!projectToDelete) return
    setIsDeleting(true)
    try {
      await projectApi.deleteProject(projectToDelete.id)
      setProjectToDelete(null)
      await loadProjects()
    } catch (error) {
      setErrorMessage(toErrorMessage(error, 'Không thể xóa dự án.'))
    } finally {
      setIsDeleting(false)
    }
  }

  function openCreateModal() {
    setEditingProject(null)
    setIsFormOpen(true)
  }

  function openEditModal(project) {
    setEditingProject(project)
    setIsFormOpen(true)
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Dự án"
        title="Quản lý danh sách dự án"
        description="Tạo mới, cập nhật thông tin hoặc truy cập nhanh từng dự án để tiếp tục làm việc với tài liệu liên quan."
        actions={<Button onClick={openCreateModal}>Tạo dự án</Button>}
      />

      {isLoading ? <LoadingState label="Đang tải danh sách dự án..." /> : null}
      {!isLoading && errorMessage ? <ErrorState description={errorMessage} onRetry={loadProjects} /> : null}
      {isEmpty ? (
        <EmptyState
          title="Chưa có dự án nào"
          description="Tạo dự án đầu tiên để bắt đầu sắp xếp tài liệu và theo dõi tiến độ xử lý nội dung của bạn."
          action={<Button onClick={openCreateModal}>Tạo dự án</Button>}
        />
      ) : null}
      {!isLoading && !errorMessage && projects.length > 0 ? (
        <ProjectTable
          projects={projects}
          onView={(project) => navigate(`/projects/${project.id}`)}
          onEdit={openEditModal}
          onDelete={setProjectToDelete}
        />
      ) : null}

      <ProjectFormModal
        open={isFormOpen}
        mode={editingProject ? 'edit' : 'create'}
        project={editingProject}
        isSubmitting={isSubmitting}
        onClose={() => {
          setIsFormOpen(false)
          setEditingProject(null)
        }}
        onSubmit={handleSubmit}
      />

      <ConfirmDialog
        open={Boolean(projectToDelete)}
        title="Xóa dự án"
        description={`Bạn có chắc muốn xóa dự án “${projectToDelete?.title ?? ''}”? Tất cả tài liệu liên quan cũng sẽ bị xóa.`}
        confirmLabel="Xóa dự án"
        isSubmitting={isDeleting}
        onClose={() => setProjectToDelete(null)}
        onConfirm={handleDelete}
      />
    </div>
  )
}
