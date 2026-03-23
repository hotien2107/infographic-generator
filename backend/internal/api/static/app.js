const state = {
  currentProjectId: null,
  currentProject: null,
};

const ui = {
  createProjectForm: document.getElementById('createProjectForm'),
  uploadForm: document.getElementById('uploadForm'),
  title: document.getElementById('title'),
  originalFilename: document.getElementById('originalFilename'),
  fileInput: document.getElementById('fileInput'),
  uploadButton: document.getElementById('uploadButton'),
  uploadHint: document.getElementById('uploadHint'),
  refreshButton: document.getElementById('refreshButton'),
  feedback: document.getElementById('feedback'),
  activityLog: document.getElementById('activityLog'),
  apiBaseLabel: document.getElementById('apiBaseLabel'),
  emptyState: document.getElementById('emptyState'),
  projectDetail: document.getElementById('projectDetail'),
  projectId: document.getElementById('projectId'),
  projectStatus: document.getElementById('projectStatus'),
  projectStep: document.getElementById('projectStep'),
  projectInputMode: document.getElementById('projectInputMode'),
  projectTitle: document.getElementById('projectTitle'),
  projectCreatedAt: document.getElementById('projectCreatedAt'),
  projectUpdatedAt: document.getElementById('projectUpdatedAt'),
  documentList: document.getElementById('documentList'),
};

ui.apiBaseLabel.textContent = window.location.origin;

ui.createProjectForm.addEventListener('submit', async (event) => {
  event.preventDefault();

  const formData = new FormData(ui.createProjectForm);
  const payload = {
    title: formData.get('title')?.trim(),
    input_mode: formData.get('input_mode'),
  };

  try {
    setFeedback('Đang tạo project...', '');
    const response = await fetch('/api/v1/projects', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });

    const result = await response.json();
    if (!response.ok) {
      throw new Error(formatApiError(result));
    }

    state.currentProjectId = result.data.id;
    logActivity(`Created project ${result.data.id} với mode ${result.data.input_mode}.`);
    setFeedback('Tạo project thành công.', 'success');
    renderProject(result.data, []);
    syncUploadState(result.data.input_mode);
    await fetchProjectDetail();
  } catch (error) {
    setFeedback(error.message || 'Không thể tạo project.', 'error');
    logActivity(`Create project failed: ${error.message}`);
  }
});

ui.uploadForm.addEventListener('submit', async (event) => {
  event.preventDefault();

  if (!state.currentProjectId) {
    setFeedback('Bạn cần tạo project trước khi upload.', 'error');
    return;
  }

  if (!ui.fileInput.files || ui.fileInput.files.length === 0) {
    setFeedback('Vui lòng chọn file để upload.', 'error');
    return;
  }

  const formData = new FormData();
  formData.append('file', ui.fileInput.files[0]);
  if (ui.originalFilename.value.trim()) {
    formData.append('original_filename', ui.originalFilename.value.trim());
  }

  try {
    setFeedback('Đang upload tài liệu...', '');
    const response = await fetch(`/api/v1/projects/${state.currentProjectId}/documents`, {
      method: 'POST',
      body: formData,
    });

    const result = await response.json();
    if (!response.ok) {
      throw new Error(formatApiError(result));
    }

    setFeedback('Upload tài liệu thành công.', 'success');
    logActivity(`Uploaded document ${result.data.document.filename} vào project ${state.currentProjectId}.`);
    ui.uploadForm.reset();
    await fetchProjectDetail();
  } catch (error) {
    setFeedback(error.message || 'Upload thất bại.', 'error');
    logActivity(`Upload failed: ${error.message}`);
  }
});

ui.refreshButton.addEventListener('click', async () => {
  try {
    setFeedback('Đang làm mới project detail...', '');
    await fetchProjectDetail();
    setFeedback('Đã làm mới dữ liệu project.', 'success');
  } catch (error) {
    setFeedback(error.message || 'Không thể tải project detail.', 'error');
  }
});

function syncUploadState(inputMode) {
  const enabled = Boolean(state.currentProjectId) && inputMode === 'file';
  ui.uploadButton.disabled = !enabled;
  ui.refreshButton.disabled = !state.currentProjectId;
  ui.uploadHint.textContent = enabled
    ? 'Project mode file đã sẵn sàng để upload.'
    : 'Project mode text không dùng upload file trong Sprint 1.';
}

async function fetchProjectDetail() {
  if (!state.currentProjectId) {
    return;
  }

  const response = await fetch(`/api/v1/projects/${state.currentProjectId}`);
  const result = await response.json();
  if (!response.ok) {
    throw new Error(formatApiError(result));
  }

  state.currentProject = result.data;
  renderProject(result.data, result.data.documents || []);
  logActivity(`Fetched project detail for ${state.currentProjectId}.`);
}

function renderProject(project, documents) {
  ui.emptyState.classList.add('hidden');
  ui.projectDetail.classList.remove('hidden');
  ui.projectId.textContent = project.id;
  ui.projectStatus.textContent = project.status;
  ui.projectStep.textContent = project.current_step;
  ui.projectInputMode.textContent = project.input_mode;
  ui.projectTitle.textContent = project.title;
  ui.projectCreatedAt.textContent = formatDate(project.created_at);
  ui.projectUpdatedAt.textContent = formatDate(project.updated_at);
  ui.documentList.innerHTML = '';

  if (!documents.length) {
    const empty = document.createElement('li');
    empty.textContent = 'Chưa có tài liệu nào được upload.';
    ui.documentList.appendChild(empty);
  } else {
    documents.forEach((documentItem) => {
      const item = document.createElement('li');
      item.innerHTML = `<strong>${documentItem.filename}</strong> · ${documentItem.mime_type} · ${documentItem.status} · ${documentItem.size_bytes} bytes`;
      ui.documentList.appendChild(item);
    });
  }

  syncUploadState(project.input_mode);
}

function formatApiError(result) {
  if (result?.error?.field) {
    return `${result.error.message} (field: ${result.error.field})`;
  }
  return result?.error?.message || 'Unknown API error';
}

function setFeedback(message, variant) {
  ui.feedback.textContent = message;
  ui.feedback.className = `feedback ${variant}`.trim();
}

function logActivity(message) {
  const item = document.createElement('li');
  item.textContent = `${new Date().toLocaleTimeString('vi-VN', { hour12: false })} — ${message}`;
  ui.activityLog.prepend(item);
}

function formatDate(value) {
  if (!value) {
    return '-';
  }
  return new Date(value).toLocaleString('vi-VN');
}

syncUploadState('text');
