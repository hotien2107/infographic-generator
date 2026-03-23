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
      payload.error?.message ?? 'Unexpected API error',
      payload.error?.code ?? 'UNKNOWN_ERROR',
      payload.error?.field ?? null,
      response.status,
    )
  }

  return payload
}

async function request(path, init) {
  const response = await fetch(`${apiBaseUrl}${path}`, init)
  return parseResponse(response)
}

export const projectApi = {
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
  uploadDocument(projectId, file) {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('original_filename', file.name)

    return request(`/api/v1/projects/${projectId}/documents`, {
      method: 'POST',
      body: formData,
    })
  },
  triggerProcessing(projectId) {
    return request(`/api/v1/projects/${projectId}/processing`, {
      method: 'POST',
    })
  },
}
