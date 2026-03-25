const apiBaseUrl = import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, '') ?? ''

export class ApiError extends Error {
  constructor(message, code, field, status) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.field = field
    this.status = status
  }
}

async function parseResponse(response) {
  const payload = await response.json()

  if (!response.ok || payload.error) {
    throw new ApiError(
      payload.error?.message ?? 'Không thể hoàn tất yêu cầu.',
      payload.error?.code ?? 'UNKNOWN_ERROR',
      payload.error?.field ?? null,
      response.status,
    )
  }

  return payload.data
}

async function request(path, init) {
  const response = await fetch(`${apiBaseUrl}${path}`, init)
  return parseResponse(response)
}

export const dashboardApi = {
  getSummary() {
    return request('/api/v1/dashboard/summary')
  },
}

export const projectApi = {
  listProjects() {
    return request('/api/v1/projects')
  },
  createProject(input) {
    return request('/api/v1/projects', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(input),
    })
  },
  getProject(projectId) {
    return request(`/api/v1/projects/${projectId}`)
  },
  updateProject(projectId, input) {
    return request(`/api/v1/projects/${projectId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(input),
    })
  },
  deleteProject(projectId) {
    return request(`/api/v1/projects/${projectId}`, {
      method: 'DELETE',
    })
  },
  uploadDocument(projectId, file) {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('original_filename', file.name)

    return request(`/api/v1/projects/${projectId}/documents`, {
      method: 'POST',
      body: formData,
    })
  },
  updateDocument(projectId, documentId, input) {
    return request(`/api/v1/projects/${projectId}/documents/${documentId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(input),
    })
  },
  deleteDocument(projectId, documentId) {
    return request(`/api/v1/projects/${projectId}/documents/${documentId}`, {
      method: 'DELETE',
    })
  },

  submitText(projectId, rawText) {
    return request(`/api/v1/projects/${projectId}/text`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ raw_text: rawText }),
    })
  },
  triggerProcessing(projectId) {
    return request(`/api/v1/projects/${projectId}/processing`, { method: 'POST' })
  },
}
