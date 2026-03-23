import { AlertCircle, CheckCircle2, FileText, Loader2, RefreshCw, Sparkles, Upload } from 'lucide-react'
import { useMemo, useState } from 'react'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { Textarea } from '@/components/ui/textarea'

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, '') ?? ''

async function apiRequest(path, init) {
  const response = await fetch(`${apiBaseUrl}${path}`, init)
  const payload = await response.json()

  if (!response.ok || payload.error) {
    throw new Error(payload.error?.message ?? 'Unexpected API error')
  }

  return payload
}

function formatBytes(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(value) {
  return new Intl.DateTimeFormat('vi-VN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

function App() {
  const [title, setTitle] = useState('Báo cáo chiến dịch Q2')
  const [inputMode, setInputMode] = useState('file')
  const [notes, setNotes] = useState('Tóm tắt nội dung tài liệu, KPI chính và insight quan trọng để AI dựng infographic.')
  const [project, setProject] = useState(null)
  const [projectId, setProjectId] = useState('')
  const [selectedFile, setSelectedFile] = useState(null)
  const [errorMessage, setErrorMessage] = useState(null)
  const [successMessage, setSuccessMessage] = useState(null)
  const [creatingProject, setCreatingProject] = useState(false)
  const [loadingProject, setLoadingProject] = useState(false)
  const [uploadingDocument, setUploadingDocument] = useState(false)

  const canUpload = useMemo(() => Boolean(project && selectedFile), [project, selectedFile])

  async function handleCreateProject(event) {
    event.preventDefault()
    setCreatingProject(true)
    setErrorMessage(null)
    setSuccessMessage(null)

    try {
      const payload = await apiRequest('/api/v1/projects', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          title,
          input_mode: inputMode,
        }),
      })

      const detailPayload = await apiRequest(`/api/v1/projects/${payload.data.id}`)
      setProject(detailPayload.data)
      setProjectId(detailPayload.data.id)
      setSuccessMessage('Project đã được tạo thành công. Bạn có thể upload tài liệu ngay bây giờ.')
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : 'Không thể tạo project.')
    } finally {
      setCreatingProject(false)
    }
  }

  async function handleLoadProject() {
    if (!projectId.trim()) {
      setErrorMessage('Vui lòng nhập project ID để tải dữ liệu.')
      return
    }

    setLoadingProject(true)
    setErrorMessage(null)
    setSuccessMessage(null)

    try {
      const payload = await apiRequest(`/api/v1/projects/${projectId.trim()}`)
      setProject(payload.data)
      setSuccessMessage('Đã tải lại thông tin project từ backend.')
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : 'Không thể tải project.')
    } finally {
      setLoadingProject(false)
    }
  }

  async function handleUpload() {
    if (!project || !selectedFile) {
      setErrorMessage('Hãy tạo project và chọn tệp trước khi upload.')
      return
    }

    setUploadingDocument(true)
    setErrorMessage(null)
    setSuccessMessage(null)

    try {
      const formData = new FormData()
      formData.append('file', selectedFile)
      formData.append('original_filename', selectedFile.name)

      await apiRequest(`/api/v1/projects/${project.id}/documents`, {
        method: 'POST',
        body: formData,
      })

      const detailPayload = await apiRequest(`/api/v1/projects/${project.id}`)
      setProject(detailPayload.data)
      setSuccessMessage('Tài liệu đã được upload. Project đã sẵn sàng cho bước xử lý tiếp theo.')
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : 'Không thể upload tài liệu.')
    } finally {
      setUploadingDocument(false)
    }
  }

  function handleFileChange(event) {
    const nextFile = event.target.files?.[0] ?? null
    setSelectedFile(nextFile)
  }

  return (
    <main className="min-h-screen px-4 py-10 text-foreground sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-7xl flex-col gap-6">
        <section className="grid gap-6 lg:grid-cols-[1.3fr_0.7fr]">
          <Card className="overflow-hidden bg-white/95">
            <CardHeader className="space-y-4 bg-slate-950 text-slate-50">
              <Badge variant="secondary" className="w-fit bg-white/10 text-white">
                Sprint 1 Frontend · React + shadcn/ui
              </Badge>
              <div className="space-y-3">
                <CardTitle className="text-4xl font-bold leading-tight">AI Infographic Generator</CardTitle>
                <CardDescription className="max-w-2xl text-slate-200">
                  Giao diện frontend tách biệt với backend Go, giúp tạo project, tải trạng thái hiện tại và upload tài liệu đầu vào đúng theo contract Sprint 1.
                </CardDescription>
              </div>
              <div className="grid gap-3 sm:grid-cols-3">
                <div className="rounded-lg border border-white/10 bg-white/5 p-4">
                  <div className="mb-2 flex items-center gap-2 text-sm font-medium"><Sparkles className="h-4 w-4" />Create project</div>
                  <p className="text-sm text-slate-300">Khởi tạo flow file/text để AI pipeline có context ngay từ bước đầu.</p>
                </div>
                <div className="rounded-lg border border-white/10 bg-white/5 p-4">
                  <div className="mb-2 flex items-center gap-2 text-sm font-medium"><Upload className="h-4 w-4" />Upload document</div>
                  <p className="text-sm text-slate-300">Upload PDF, DOCX hoặc TXT và đồng bộ trạng thái project tức thì.</p>
                </div>
                <div className="rounded-lg border border-white/10 bg-white/5 p-4">
                  <div className="mb-2 flex items-center gap-2 text-sm font-medium"><RefreshCw className="h-4 w-4" />Refresh detail</div>
                  <p className="text-sm text-slate-300">Tải lại project hiện có từ backend để kiểm tra tiến trình xử lý.</p>
                </div>
              </div>
            </CardHeader>
            <CardContent className="grid gap-6 p-6 lg:grid-cols-[1.1fr_0.9fr]">
              <form className="space-y-5" onSubmit={handleCreateProject}>
                <div className="space-y-2">
                  <Label htmlFor="title">Tên project</Label>
                  <Input id="title" value={title} onChange={(event) => setTitle(event.target.value)} placeholder="Ví dụ: Báo cáo tăng trưởng Q2" />
                </div>
                <div className="space-y-3">
                  <Label>Input mode</Label>
                  <div className="grid gap-3 sm:grid-cols-2">
                    {['file', 'text'].map((mode) => (
                      <button
                        key={mode}
                        type="button"
                        onClick={() => setInputMode(mode)}
                        className={`rounded-lg border p-4 text-left transition ${
                          inputMode === mode ? 'border-blue-600 bg-blue-50 shadow-sm' : 'border-border bg-white hover:border-blue-200'
                        }`}
                      >
                        <div className="mb-2 flex items-center gap-2 font-medium capitalize">
                          {mode === 'file' ? <Upload className="h-4 w-4" /> : <FileText className="h-4 w-4" />}
                          {mode}
                        </div>
                        <p className="text-sm text-muted-foreground">
                          {mode === 'file'
                            ? 'Dành cho tài liệu gốc như PDF, DOCX hoặc TXT để frontend upload trực tiếp.'
                            : 'Chuẩn bị sẵn cho luồng nhập text trực tiếp ở các sprint tiếp theo.'}
                        </p>
                      </button>
                    ))}
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="notes">Ghi chú cho đội nội dung</Label>
                  <Textarea id="notes" value={notes} onChange={(event) => setNotes(event.target.value)} placeholder="Mô tả scope infographic mong muốn..." />
                  <p className="text-xs text-muted-foreground">Trường này hiện chỉ phục vụ UI note, không gửi sang backend Sprint 1.</p>
                </div>
                <Button type="submit" className="w-full sm:w-auto" disabled={creatingProject}>
                  {creatingProject ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                  Tạo project
                </Button>
              </form>

              <div className="space-y-5 rounded-xl border border-dashed border-border bg-slate-50/80 p-5">
                <div>
                  <h2 className="text-lg font-semibold">Kết nối backend</h2>
                  <p className="mt-1 text-sm text-muted-foreground">
                    Frontend sẽ gọi API thông qua <code className="rounded bg-slate-100 px-1 py-0.5">VITE_API_BASE_URL</code> hoặc proxy của Vite tại <code className="rounded bg-slate-100 px-1 py-0.5">http://localhost:8080</code>.
                  </p>
                </div>
                <Separator />
                <div className="space-y-3">
                  <Label htmlFor="project-id">Tải project theo ID</Label>
                  <div className="flex flex-col gap-3 sm:flex-row">
                    <Input id="project-id" value={projectId} onChange={(event) => setProjectId(event.target.value)} placeholder="Dán UUID project để reload" />
                    <Button type="button" variant="outline" onClick={handleLoadProject} disabled={loadingProject}>
                      {loadingProject ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
                      Tải lại
                    </Button>
                  </div>
                </div>
                <div className="space-y-3">
                  <Label htmlFor="document">Upload tài liệu</Label>
                  <Input id="document" type="file" accept=".pdf,.docx,.txt" onChange={handleFileChange} />
                  <Button type="button" onClick={handleUpload} disabled={!canUpload || uploadingDocument}>
                    {uploadingDocument ? <Loader2 className="h-4 w-4 animate-spin" /> : <Upload className="h-4 w-4" />}
                    Upload file
                  </Button>
                  {inputMode === 'text' ? (
                    <p className="text-xs text-amber-700">Project mode đang là <strong>text</strong>, nhưng backend Sprint 1 hiện mới có endpoint upload document. UI vẫn cho phép đổi mode để chuẩn bị cho sprint sau.</p>
                  ) : null}
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="bg-white/95">
            <CardHeader>
              <CardTitle className="text-2xl">Snapshot trạng thái</CardTitle>
              <CardDescription>Quan sát nhanh tiến độ của project và bộ tài liệu đã upload.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-3 sm:grid-cols-2">
                <div className="rounded-lg bg-slate-50 p-4">
                  <p className="text-sm text-muted-foreground">Project status</p>
                  <p className="mt-2 text-2xl font-semibold">{project?.status ?? '—'}</p>
                </div>
                <div className="rounded-lg bg-slate-50 p-4">
                  <p className="text-sm text-muted-foreground">Current step</p>
                  <p className="mt-2 text-2xl font-semibold">{project?.current_step ?? '—'}</p>
                </div>
              </div>
              <div className="rounded-lg border border-border bg-slate-50/80 p-4">
                <p className="text-sm text-muted-foreground">Documents attached</p>
                <p className="mt-2 text-3xl font-semibold">{project?.documents.length ?? 0}</p>
              </div>
              <div className="rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
                {project ? (
                  <div className="space-y-2">
                    <p><span className="font-medium text-foreground">Project ID:</span> {project.id}</p>
                    <p><span className="font-medium text-foreground">Tạo lúc:</span> {formatDate(project.created_at)}</p>
                    <p><span className="font-medium text-foreground">Cập nhật:</span> {formatDate(project.updated_at)}</p>
                  </div>
                ) : (
                  'Chưa có project nào được tải. Hãy tạo project mới hoặc nhập UUID để xem dữ liệu.'
                )}
              </div>
            </CardContent>
          </Card>
        </section>

        {(errorMessage || successMessage) && (
          <Alert className={errorMessage ? 'border-red-200 bg-red-50' : 'border-emerald-200 bg-emerald-50'}>
            <div className="flex items-start gap-3">
              {errorMessage ? <AlertCircle className="mt-0.5 h-4 w-4 text-red-600" /> : <CheckCircle2 className="mt-0.5 h-4 w-4 text-emerald-600" />}
              <div>
                <AlertTitle>{errorMessage ? 'Có lỗi xảy ra' : 'Thành công'}</AlertTitle>
                <AlertDescription>{errorMessage ?? successMessage}</AlertDescription>
              </div>
            </div>
          </Alert>
        )}

        <section className="grid gap-6 lg:grid-cols-[0.9fr_1.1fr]">
          <Card className="bg-white/95">
            <CardHeader>
              <CardTitle className="text-2xl">Checklist tích hợp</CardTitle>
              <CardDescription>Các API hiện có trong backend mà frontend này đã bám theo.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 text-sm text-slate-700">
              <div className="rounded-lg border border-border p-4">
                <p className="font-medium">POST /api/v1/projects</p>
                <p className="mt-1 text-muted-foreground">Tạo project mới với title và input_mode.</p>
              </div>
              <div className="rounded-lg border border-border p-4">
                <p className="font-medium">GET /api/v1/projects/:projectId</p>
                <p className="mt-1 text-muted-foreground">Tải thông tin project, trạng thái hiện tại và danh sách document.</p>
              </div>
              <div className="rounded-lg border border-border p-4">
                <p className="font-medium">POST /api/v1/projects/:projectId/documents</p>
                <p className="mt-1 text-muted-foreground">Upload file đầu vào và cập nhật project sang trạng thái uploaded.</p>
              </div>
            </CardContent>
          </Card>

          <Card className="bg-white/95">
            <CardHeader>
              <CardTitle className="text-2xl">Danh sách tài liệu</CardTitle>
              <CardDescription>Hiển thị dữ liệu trả về từ endpoint chi tiết project.</CardDescription>
            </CardHeader>
            <CardContent>
              {project?.documents.length ? (
                <div className="space-y-3">
                  {project.documents.map((document) => (
                    <div key={document.id} className="flex flex-col gap-3 rounded-xl border border-border bg-slate-50 p-4 sm:flex-row sm:items-center sm:justify-between">
                      <div>
                        <div className="flex items-center gap-2">
                          <p className="font-medium text-foreground">{document.filename}</p>
                          <Badge variant="outline">{document.status}</Badge>
                        </div>
                        <p className="mt-1 text-sm text-muted-foreground">{document.mime_type} · {formatBytes(document.size_bytes)}</p>
                      </div>
                      <div className="text-sm text-muted-foreground sm:text-right">
                        <p>Storage key</p>
                        <p className="font-medium text-foreground break-all">{document.storage_key}</p>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="rounded-xl border border-dashed border-border p-10 text-center text-sm text-muted-foreground">
                  Chưa có tài liệu nào. Chọn file rồi upload để xem dữ liệu phản hồi từ backend.
                </div>
              )}
            </CardContent>
          </Card>
        </section>
      </div>
    </main>
  )
}

export default App
