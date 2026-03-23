import { useMemo, useState } from 'react'
import {
  CheckCircle2,
  FileUp,
  FolderKanban,
  LoaderCircle,
  Sparkles,
  UploadCloud,
  Waypoints,
} from 'lucide-react'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Textarea } from '@/components/ui/textarea'

type InputMode = 'file' | 'text'
type ProjectStatus = 'draft' | 'uploaded'
type ProjectStep = 'project_created' | 'waiting_for_upload' | 'uploaded'

type ApiMeta = {
  request_id: string
  timestamp: string
}

type ApiErrorDetail = {
  code: string
  message: string
  field: string | null
}

type ApiEnvelope<T> = {
  data: T
  error: ApiErrorDetail | null
  meta: ApiMeta
}

type Project = {
  id: string
  title: string
  input_mode: InputMode
  status: ProjectStatus
  current_step: ProjectStep
  created_at: string
  updated_at: string
}

type Document = {
  id: string
  project_id: string
  filename: string
  mime_type: string
  size_bytes: number
  storage_key: string
  status: 'uploaded'
  created_at: string
}

type ProjectDetail = Project & {
  documents: Document[]
}

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL?.trim() || 'http://localhost:8080'

const sprintStats = [
  { label: 'API contract', value: 'Sprint 1', hint: 'create project · upload file · get detail' },
  { label: 'Frontend stack', value: 'React + shadcn/ui', hint: 'Vite, TypeScript, Tailwind' },
  { label: 'Backend separation', value: 'Independent app', hint: 'Runs as a dedicated frontend service' },
] as const

const acceptedTypes = '.pdf,.docx,.txt'

function formatDate(value: string) {
  return new Intl.DateTimeFormat('en', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

function formatFileSize(value: number) {
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  return `${(value / (1024 * 1024)).toFixed(2)} MB`
}

async function parseJson<T>(response: Response): Promise<ApiEnvelope<T>> {
  const payload = (await response.json()) as ApiEnvelope<T>
  if (!response.ok || payload.error) {
    throw new Error(payload.error?.message || 'Request failed')
  }
  return payload
}

export default function App() {
  const [inputMode, setInputMode] = useState<InputMode>('file')
  const [title, setTitle] = useState('Q2 Product Launch Overview')
  const [textDraft, setTextDraft] = useState('Paste narrative or bullet points here. The text-to-infographic processing flow can be wired in the next sprint.')
  const [project, setProject] = useState<Project | null>(null)
  const [projectDetail, setProjectDetail] = useState<ProjectDetail | null>(null)
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [createPending, setCreatePending] = useState(false)
  const [uploadPending, setUploadPending] = useState(false)
  const [refreshPending, setRefreshPending] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const currentProject = useMemo(() => projectDetail ?? project, [project, projectDetail])

  async function handleCreateProject() {
    setCreatePending(true)
    setErrorMessage(null)
    setSuccessMessage(null)

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/projects`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          title,
          input_mode: inputMode,
        }),
      })

      const payload = await parseJson<Project>(response)
      setProject(payload.data)
      setProjectDetail({ ...payload.data, documents: [] })
      setSuccessMessage(`Created project ${payload.data.title}. You can now continue the ${inputMode} workflow.`)
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : 'Failed to create project')
    } finally {
      setCreatePending(false)
    }
  }

  async function handleRefreshProject() {
    if (!currentProject) return

    setRefreshPending(true)
    setErrorMessage(null)

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/projects/${currentProject.id}`)
      const payload = await parseJson<ProjectDetail>(response)
      setProject(payload.data)
      setProjectDetail(payload.data)
      setSuccessMessage('Project detail refreshed from backend.')
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : 'Failed to refresh project')
    } finally {
      setRefreshPending(false)
    }
  }

  async function handleUploadDocument() {
    if (!currentProject || !selectedFile) {
      setErrorMessage('Choose a file before uploading.')
      return
    }

    setUploadPending(true)
    setErrorMessage(null)

    try {
      const formData = new FormData()
      formData.append('file', selectedFile)
      formData.append('original_filename', selectedFile.name)

      const response = await fetch(`${apiBaseUrl}/api/v1/projects/${currentProject.id}/documents`, {
        method: 'POST',
        body: formData,
      })

      await parseJson<{ project: Project; document: Document }>(response)
      setSuccessMessage(`Uploaded ${selectedFile.name}.`)
      setSelectedFile(null)
      await handleRefreshProject()
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : 'Failed to upload document')
    } finally {
      setUploadPending(false)
    }
  }

  return (
    <main className="relative overflow-hidden">
      <div className="absolute inset-x-0 top-0 h-72 bg-[radial-gradient(circle_at_top,rgba(56,189,248,0.15),transparent_45%)]" />
      <div className="container relative py-8 sm:py-12 lg:py-16">
        <section className="grid gap-6 lg:grid-cols-[1.3fr_0.7fr] lg:items-start">
          <div className="space-y-6">
            <Badge variant="secondary" className="w-fit border border-cyan-400/20 bg-cyan-400/10 text-cyan-100">
              Separate frontend workspace
            </Badge>
            <div className="space-y-4">
              <h1 className="max-w-3xl text-4xl font-semibold tracking-tight sm:text-5xl">
                React frontend for the AI infographic workflow, shipped independently from the Go backend.
              </h1>
              <p className="max-w-2xl text-base text-muted-foreground sm:text-lg">
                This UI uses shadcn-style components to create projects, upload source files, and inspect Sprint 1 backend responses without coupling frontend delivery to backend deployment.
              </p>
            </div>

            <div className="grid gap-4 md:grid-cols-3">
              {sprintStats.map((item) => (
                <Card key={item.label} className="border-white/10 bg-white/5 backdrop-blur">
                  <CardHeader className="pb-3">
                    <CardDescription>{item.label}</CardDescription>
                    <CardTitle className="text-lg">{item.value}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm text-muted-foreground">{item.hint}</p>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>

          <Card className="border-white/10 bg-white/5 backdrop-blur">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-xl">
                <Sparkles className="h-5 w-5 text-cyan-300" />
                Delivery notes
              </CardTitle>
              <CardDescription>Frontend and backend now run as separate apps in local development.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 text-sm text-muted-foreground">
              <div className="flex gap-3 rounded-lg border border-white/10 bg-slate-950/40 p-4">
                <FolderKanban className="mt-0.5 h-4 w-4 text-violet-300" />
                <p>Start the Go API from <code className="rounded bg-white/10 px-1.5 py-0.5 text-xs text-white">backend/</code> and the React app from <code className="rounded bg-white/10 px-1.5 py-0.5 text-xs text-white">frontend/</code>.</p>
              </div>
              <div className="flex gap-3 rounded-lg border border-white/10 bg-slate-950/40 p-4">
                <Waypoints className="mt-0.5 h-4 w-4 text-cyan-300" />
                <p>The frontend reads <code className="rounded bg-white/10 px-1.5 py-0.5 text-xs text-white">VITE_API_BASE_URL</code>, so each environment can point to a different API without rebuilding backend code.</p>
              </div>
              <div className="flex gap-3 rounded-lg border border-white/10 bg-slate-950/40 p-4">
                <CheckCircle2 className="mt-0.5 h-4 w-4 text-emerald-300" />
                <p>Local CORS support is enabled so the frontend can call the Go API directly from a separate origin in development.</p>
              </div>
            </CardContent>
          </Card>
        </section>

        <section className="mt-8 grid gap-6 xl:grid-cols-[1.05fr_0.95fr]">
          <Card className="border-white/10 bg-slate-950/40">
            <CardHeader>
              <CardTitle>1. Create a project</CardTitle>
              <CardDescription>Choose the input mode first so the UI can render the correct workflow.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-2">
                <Label htmlFor="title">Project title</Label>
                <Input
                  id="title"
                  value={title}
                  onChange={(event) => setTitle(event.target.value)}
                  placeholder="Enter a descriptive project name"
                />
              </div>

              <div className="space-y-3">
                <Label>Input mode</Label>
                <RadioGroup value={inputMode} onValueChange={(value) => setInputMode(value as InputMode)} className="grid gap-3 md:grid-cols-2">
                  <label className="flex cursor-pointer items-start gap-3 rounded-xl border border-white/10 bg-white/5 p-4 transition hover:border-violet-400/40 hover:bg-white/10">
                    <RadioGroupItem value="file" id="mode-file" className="mt-1" />
                    <div className="space-y-1">
                      <p className="font-medium text-white">Upload file</p>
                      <p className="text-sm text-muted-foreground">Best for PDF, DOCX, or TXT source documents.</p>
                    </div>
                  </label>
                  <label className="flex cursor-pointer items-start gap-3 rounded-xl border border-white/10 bg-white/5 p-4 transition hover:border-cyan-400/40 hover:bg-white/10">
                    <RadioGroupItem value="text" id="mode-text" className="mt-1" />
                    <div className="space-y-1">
                      <p className="font-medium text-white">Paste text</p>
                      <p className="text-sm text-muted-foreground">Creates the project now; text-processing action can be added in the next sprint.</p>
                    </div>
                  </label>
                </RadioGroup>
              </div>

              {inputMode === 'text' ? (
                <div className="space-y-2">
                  <Label htmlFor="text-draft">Text draft preview</Label>
                  <Textarea
                    id="text-draft"
                    value={textDraft}
                    onChange={(event) => setTextDraft(event.target.value)}
                  />
                  <p className="text-xs text-muted-foreground">This field prepares the UX for the text pipeline while the current Sprint 1 backend contract remains unchanged.</p>
                </div>
              ) : (
                <div className="rounded-xl border border-dashed border-violet-400/30 bg-violet-500/5 p-4 text-sm text-muted-foreground">
                  File mode mirrors the current backend contract exactly: create project first, then upload a supported document.
                </div>
              )}

              <div className="flex flex-wrap gap-3">
                <Button onClick={handleCreateProject} disabled={createPending || title.trim().length < 3}>
                  {createPending ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <FolderKanban className="h-4 w-4" />}
                  Create project
                </Button>
                <Button variant="outline" onClick={handleRefreshProject} disabled={!currentProject || refreshPending}>
                  {refreshPending ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <Waypoints className="h-4 w-4" />}
                  Refresh detail
                </Button>
              </div>

              {(errorMessage || successMessage) && (
                <div className={`rounded-xl border px-4 py-3 text-sm ${errorMessage ? 'border-red-400/30 bg-red-500/10 text-red-100' : 'border-emerald-400/30 bg-emerald-500/10 text-emerald-100'}`}>
                  {errorMessage ?? successMessage}
                </div>
              )}
            </CardContent>
          </Card>

          <Card className="border-white/10 bg-slate-950/40">
            <CardHeader>
              <CardTitle>2. Upload and inspect project state</CardTitle>
              <CardDescription>Use the generated project id to continue the backend flow from a separate frontend app.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid gap-3 sm:grid-cols-2">
                <StatusItem label="Project ID" value={currentProject?.id ?? 'Not created yet'} />
                <StatusItem label="API base URL" value={apiBaseUrl} />
                <StatusItem label="Status" value={currentProject?.status ?? 'draft'} />
                <StatusItem label="Current step" value={currentProject?.current_step ?? 'waiting_for_upload'} />
              </div>

              <Separator />

              <div className="space-y-3">
                <Label htmlFor="file">Document upload</Label>
                <div className="rounded-xl border border-dashed border-cyan-400/30 bg-cyan-500/5 p-4">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div className="space-y-1">
                      <p className="flex items-center gap-2 font-medium text-white">
                        <UploadCloud className="h-4 w-4 text-cyan-300" />
                        Supported files: PDF, DOCX, TXT
                      </p>
                      <p className="text-sm text-muted-foreground">Create a project in file mode, then upload a document from this separate React app.</p>
                    </div>
                    <Input
                      id="file"
                      type="file"
                      accept={acceptedTypes}
                      className="max-w-xs file:text-foreground"
                      onChange={(event) => setSelectedFile(event.target.files?.[0] ?? null)}
                    />
                  </div>
                  {selectedFile && (
                    <p className="mt-3 text-sm text-cyan-100">
                      Selected <span className="font-medium">{selectedFile.name}</span> · {formatFileSize(selectedFile.size)}
                    </p>
                  )}
                </div>
                <Button onClick={handleUploadDocument} disabled={!currentProject || currentProject.input_mode !== 'file' || !selectedFile || uploadPending}>
                  {uploadPending ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <FileUp className="h-4 w-4" />}
                  Upload document
                </Button>
                {currentProject?.input_mode === 'text' && (
                  <p className="text-xs text-muted-foreground">Text mode project creation is supported now; text submission can be added once the backend contract expands.</p>
                )}
              </div>

              <Separator />

              <div className="space-y-3">
                <div>
                  <p className="font-medium text-white">Documents returned by backend</p>
                  <p className="text-sm text-muted-foreground">The panel below reads the canonical Sprint 1 project detail response.</p>
                </div>
                <ScrollArea className="h-[260px] rounded-xl border border-white/10 bg-black/20">
                  <div className="space-y-3 p-4">
                    {projectDetail?.documents?.length ? (
                      projectDetail.documents.map((document) => (
                        <div key={document.id} className="rounded-xl border border-white/10 bg-white/5 p-4">
                          <div className="flex items-start justify-between gap-3">
                            <div>
                              <p className="font-medium text-white">{document.filename}</p>
                              <p className="text-sm text-muted-foreground">{document.mime_type} · {formatFileSize(document.size_bytes)}</p>
                            </div>
                            <Badge>{document.status}</Badge>
                          </div>
                          <div className="mt-3 grid gap-2 text-xs text-muted-foreground sm:grid-cols-2">
                            <p>ID: {document.id}</p>
                            <p>Stored at: {document.storage_key}</p>
                            <p>Project: {document.project_id}</p>
                            <p>Created: {formatDate(document.created_at)}</p>
                          </div>
                        </div>
                      ))
                    ) : (
                      <div className="rounded-xl border border-dashed border-white/10 p-6 text-sm text-muted-foreground">
                        No documents uploaded yet. Once you upload a file, refresh or wait for the automatic reload to see backend state here.
                      </div>
                    )}
                  </div>
                </ScrollArea>
              </div>
            </CardContent>
          </Card>
        </section>
      </div>
    </main>
  )
}

function StatusItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-white/10 bg-white/5 p-4">
      <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">{label}</p>
      <p className="mt-2 break-all text-sm font-medium text-white">{value}</p>
    </div>
  )
}
