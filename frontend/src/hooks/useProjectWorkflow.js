import { useMemo, useState } from 'react'

import { projectApi } from '@/lib/api'

function toUserMessage(error, fallback) {
  return error instanceof Error ? error.message : fallback
}

export function useProjectWorkflow() {
  const [project, setProject] = useState(null)
  const [projectId, setProjectId] = useState('')
  const [selectedFile, setSelectedFile] = useState(null)
  const [title, setTitle] = useState('Báo cáo chiến dịch Q2')
  const [inputMode, setInputMode] = useState('file')
  const [notes, setNotes] = useState('Tóm tắt nội dung tài liệu, KPI chính và insight quan trọng để AI dựng infographic.')
  const [errorMessage, setErrorMessage] = useState(null)
  const [successMessage, setSuccessMessage] = useState(null)
  const [isCreating, setIsCreating] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isUploading, setIsUploading] = useState(false)
  const [isTriggering, setIsTriggering] = useState(false)

  const canUpload = useMemo(() => Boolean(project && selectedFile), [project, selectedFile])
  const canTriggerProcessing = useMemo(() => Boolean(project?.documents?.length), [project])

  async function refreshProject(id, successMessageText) {
    const payload = await projectApi.getProject(id)
    setProject(payload.data)
    setProjectId(payload.data.id)
    if (successMessageText) {
      setSuccessMessage(successMessageText)
    }
    return payload.data
  }

  async function handleCreateProject(event) {
    event.preventDefault()
    setIsCreating(true)
    setErrorMessage(null)
    setSuccessMessage(null)

    try {
      const payload = await projectApi.createProject({
        title,
        input_mode: inputMode,
      })
      await refreshProject(payload.data.id, 'Project đã được tạo thành công. Bạn có thể upload tài liệu ngay bây giờ.')
    } catch (error) {
      setErrorMessage(toUserMessage(error, 'Không thể tạo project.'))
    } finally {
      setIsCreating(false)
    }
  }

  async function handleLoadProject() {
    if (!projectId.trim()) {
      setErrorMessage('Vui lòng nhập project ID để tải dữ liệu.')
      setSuccessMessage(null)
      return
    }

    setIsLoading(true)
    setErrorMessage(null)
    setSuccessMessage(null)

    try {
      await refreshProject(projectId.trim(), 'Đã tải lại thông tin project từ backend.')
    } catch (error) {
      setErrorMessage(toUserMessage(error, 'Không thể tải project.'))
    } finally {
      setIsLoading(false)
    }
  }

  async function handleUpload() {
    if (!project || !selectedFile) {
      setErrorMessage('Hãy tạo project và chọn tệp trước khi upload.')
      setSuccessMessage(null)
      return
    }

    setIsUploading(true)
    setErrorMessage(null)
    setSuccessMessage(null)

    try {
      await projectApi.uploadDocument(project.id, selectedFile)
      await refreshProject(project.id, 'Tài liệu đã được upload. Hệ thống sẽ tự động đưa tài liệu vào luồng xử lý.')
    } catch (error) {
      setErrorMessage(toUserMessage(error, 'Không thể upload tài liệu.'))
    } finally {
      setIsUploading(false)
    }
  }

  async function handleTriggerProcessing() {
    if (!project) return

    setIsTriggering(true)
    setErrorMessage(null)
    setSuccessMessage(null)

    try {
      await projectApi.triggerProcessing(project.id)
      await refreshProject(project.id, 'Đã đưa tài liệu mới nhất vào hàng đợi xử lý.')
    } catch (error) {
      setErrorMessage(toUserMessage(error, 'Không thể trigger processing.'))
    } finally {
      setIsTriggering(false)
    }
  }

  return {
    state: {
      project,
      projectId,
      selectedFile,
      title,
      inputMode,
      notes,
      errorMessage,
      successMessage,
      isCreating,
      isLoading,
      isUploading,
      isTriggering,
      canUpload,
      canTriggerProcessing,
    },
    actions: {
      setProjectId,
      setSelectedFile,
      setTitle,
      setInputMode,
      setNotes,
      setErrorMessage,
      setSuccessMessage,
      handleCreateProject,
      handleLoadProject,
      handleUpload,
      handleTriggerProcessing,
      refreshProject,
    },
  }
}
