import { LitElement, html, unsafeCSS } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { apiService, type ProcessResponse, type UploadResponse } from '../services/api.js';
import { generateUUID } from '../utils/uuid.js';
import uploadStyles from '../styles/upload.css?inline';

type UploadItemStatus = 'pending' | 'uploading' | 'completed' | 'error' | 'paused' | 'processing';

export interface UploadFile {
  id: string;
  file: File;
  progress: number;
  status: UploadItemStatus;
  uploadId?: string;
  uploadResponse?: UploadResponse;
  processResponse?: ProcessResponse;
  error?: string;
  abortController?: AbortController;
}

interface UploadSummary {
  total: number;
  completed: number;
  duplicates: number;
  failed: number;
  processing: number;
  pending: number;
}

@customElement('sortify-upload')
export class SortifyUpload extends LitElement {
  static styles = unsafeCSS(uploadStyles);

  @property({ type: Boolean })
  disabled = false;

  @property({ type: Number })
  maxFileSize = 0;

  @state()
  private uploadQueue: UploadFile[] = [];

  @state()
  private isDragOver = false;

  @state()
  private isUploading = false;

  @state()
  private thumbnailCache = new Map<string, string>();

  private fileInputRef?: HTMLInputElement;

  render() {
    const summary = this.getSummary();

    return html`
      <div class="upload-container">
        <div
          class="dropzone ${this.isDragOver ? 'drag-over' : ''}"
          @click=${this.openFileDialog}
          @dragover=${this.handleDragOver}
          @dragleave=${this.handleDragLeave}
          @drop=${this.handleDrop}
        >
          <div class="dropzone-content">
            <div class="dropzone-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M14 2H6C4.9 2 4 2.9 4 4V20C4 21.1 4.89 22 5.99 22H18C19.1 22 20 21.1 20 20V8L14 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <polyline points="14,2 14,8 20,8" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <line x1="16" y1="13" x2="8" y2="13" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <line x1="16" y1="17" x2="8" y2="17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <polyline points="10,9 9,9 8,9" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <h3 class="dropzone-title">Bulk upload your photo and video library</h3>
            <p class="dropzone-subtitle">
              Sortify now treats uploads as a queue, verifies each file, skips exact duplicates, and organizes by trusted metadata first.
            </p>
            <button class="btn-primary" ?disabled=${this.disabled}>
              Choose Files
            </button>
          </div>
        </div>

        <input
          type="file"
          class="file-input"
          multiple
          accept="image/*,video/*"
          @change=${this.handleFileSelect}
        />

        ${summary.total > 0 ? html`
          <div class="upload-summary">
            <div class="summary-card">
              <span class="summary-value">${summary.total}</span>
              <span class="summary-label">Queued</span>
            </div>
            <div class="summary-card">
              <span class="summary-value">${summary.completed}</span>
              <span class="summary-label">Stored</span>
            </div>
            <div class="summary-card">
              <span class="summary-value">${summary.duplicates}</span>
              <span class="summary-label">Duplicates skipped</span>
            </div>
            <div class="summary-card">
              <span class="summary-value">${summary.failed}</span>
              <span class="summary-label">Failed</span>
            </div>
          </div>
        ` : null}

        ${this.uploadQueue.length > 0 ? html`
          <div class="upload-queue">
            <div class="queue-header">
              <h4 class="queue-title">
                Upload Queue
                <span class="queue-subtitle">
                  ${this.describeQueueState(summary)}
                </span>
              </h4>
              <div class="queue-actions">
                <button class="btn-secondary" @click=${this.pauseAll}>
                  ${this.isUploading ? '⏸️ Pause Queue' : '▶️ Resume Queue'}
                </button>
                <button class="btn-secondary" @click=${this.clearCompleted}>
                  🗑️ Clear Completed
                </button>
                <button class="btn-danger" @click=${this.clearAll}>
                  ❌ Clear All
                </button>
              </div>
            </div>

            ${this.uploadQueue.map((item) => this.renderUploadItem(item))}
          </div>
        ` : html`
          <div class="empty-state">
            <div class="empty-state-icon">📂</div>
            <p class="empty-state-text">No files selected yet</p>
            <p style="font-size: 0.875rem; margin-top: 0.5rem; opacity: 0.8;">
              Drop a batch above or click to browse your library.
            </p>
          </div>
        `}
      </div>
    `;
  }

  private renderUploadItem(item: UploadFile) {
    const result = item.processResponse;
    const metadata = result?.metadata;
    const duplicate = result?.duplicate ?? item.uploadResponse?.duplicate ?? false;
    const organizationLabel = duplicate
      ? 'Duplicate skipped'
      : result?.organizedPath || item.uploadResponse?.relativePath || 'Awaiting organization details';
    const detailLabel = metadata?.dateTaken && metadata?.dateSource
      ? `${metadata.dateSource} • ${new Date(metadata.dateTaken).toLocaleString()}`
      : metadata?.dateSource || 'No trusted capture date';

    return html`
      <div class="upload-item">
        ${this.getFileIcon(item.file)}

        <div class="file-info">
          <div class="file-name" title="${item.file.name}">
            ${item.file.name}
          </div>
          <div class="file-details">
            ${this.formatFileSize(item.file.size)} • ${item.file.type.split('/')[0] || 'Unknown'}
          </div>
          ${result ? html`
            <div class="result-chip ${duplicate ? 'duplicate' : 'stored'}">
              ${organizationLabel}
            </div>
            <div class="result-meta">${detailLabel}</div>
          ` : null}
          ${item.error ? html`
            <div class="error-message">
              <span class="error-icon">⚠️</span>
              <span>${this.formatErrorMessage(item.error)}</span>
            </div>
          ` : null}
        </div>

        <div class="progress-container">
          <div class="progress-bar">
            <div class="progress-fill" style="width: 100%; transform: scaleX(${item.progress / 100})"></div>
          </div>
          <div class="progress-text">${item.progress}% ${this.getProgressLabel(item.status)}</div>
        </div>

        <div class="status-badge ${this.getStatusClass(item.status)}" title="${item.status}">
          ${this.getStatusIcon(item.status)}
        </div>

        <div class="item-actions">
          ${item.status === 'uploading' ? html`
            <button class="btn-icon btn-icon-neutral" @click=${() => this.pauseUpload(item.id)} title="Pause upload">
              ⏸️
            </button>
          ` : null}
          ${item.status === 'paused' ? html`
            <button class="btn-icon btn-icon-success" @click=${() => this.resumeUpload(item.id)} title="Resume upload">
              ▶️
            </button>
          ` : null}
          ${item.status === 'error' ? html`
            <button class="btn-icon btn-icon-success" @click=${() => this.retryUpload(item.id)} title="Retry upload">
              🔄
            </button>
          ` : null}
          ${item.status !== 'completed' ? html`
            <button class="btn-icon btn-icon-danger" @click=${() => this.removeFromQueue(item.id)} title="Remove from queue">
              🗑️
            </button>
          ` : null}
        </div>
      </div>
    `;
  }

  private getFileIcon(file: File) {
    if (file.type.startsWith('image/')) {
      return html`
        <div class="file-icon image">
          <img
            src="${this.getThumbnailUrl(file)}"
            alt="${file.name}"
            class="thumbnail-image"
            @error=${this.handleThumbnailError}
          />
        </div>
      `;
    }

    if (file.type.startsWith('video/')) {
      return html`<div class="file-icon video">🎬</div>`;
    }

    return html`<div class="file-icon">📄</div>`;
  }

  private openFileDialog() {
    if (this.disabled) {
      return;
    }

    if (!this.fileInputRef) {
      this.fileInputRef = this.shadowRoot?.querySelector('.file-input') as HTMLInputElement;
    }

    this.fileInputRef?.click();
  }

  private handleDragOver(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
    this.isDragOver = true;
  }

  private handleDragLeave(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
    this.isDragOver = false;
  }

  private handleDrop(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
    this.isDragOver = false;

    if (this.disabled) {
      return;
    }

    const files = Array.from(event.dataTransfer?.files || []);
    this.addFilesToQueue(files);
  }

  private handleFileSelect(event: Event) {
    const input = event.target as HTMLInputElement;
    const files = Array.from(input.files || []);
    this.addFilesToQueue(files);
    input.value = '';
  }

  private addFilesToQueue(files: File[]) {
    const newItems: UploadFile[] = files
      .filter((file) => this.isValidFile(file))
      .map((file) => ({
        id: generateUUID(),
        file,
        progress: 0,
        status: 'pending',
      }));

    this.uploadQueue = [...this.uploadQueue, ...newItems];
    this.startNextUpload();
  }

  private isValidFile(file: File): boolean {
    if (!file.type.startsWith('image/') && !file.type.startsWith('video/')) {
      return false;
    }

    if (this.maxFileSize > 0 && file.size > this.maxFileSize) {
      return false;
    }

    return true;
  }

  private async startNextUpload() {
    if (this.isUploading) {
      return;
    }

    const nextItem = this.uploadQueue.find((item) => item.status === 'pending');
    if (!nextItem) {
      return;
    }

    this.isUploading = true;
    nextItem.status = 'uploading';
    this.requestUpdate();

    try {
      await this.uploadFile(nextItem);
    } catch (error) {
      nextItem.status = 'error';
      nextItem.error = error instanceof Error ? error.message : 'Upload failed';
    }

    this.isUploading = false;
    this.requestUpdate();
    setTimeout(() => this.startNextUpload(), 50);
  }

  private async uploadFile(item: UploadFile) {
    item.abortController = new AbortController();

    try {
      const uploadResponse = await apiService.uploadFile(item.file, {
        onProgress: (progress) => {
          item.progress = Math.round(progress * 100);
          this.requestUpdate();
        },
        onError: (error) => {
          item.status = 'error';
          item.error = error.message;
          this.requestUpdate();
        },
        signal: item.abortController.signal,
      });

      item.uploadResponse = uploadResponse;
      item.uploadId = uploadResponse.sessionId;
      item.status = 'processing';
      item.progress = 100;
      this.requestUpdate();

      item.processResponse = {
        id: uploadResponse.sessionId,
        originalPath: item.file.name,
        organizedPath: uploadResponse.relativePath || uploadResponse.finalPath || uploadResponse.finalFileName || uploadResponse.filename,
        metadata: uploadResponse.mediaInfo,
        status: 'completed',
        duplicate: uploadResponse.duplicate,
      };

      item.status = 'completed';
      this.requestUpdate();
    } catch (error) {
      if (error instanceof Error && error.message === 'Upload cancelled') {
        return;
      }

      item.status = 'error';
      item.error = error instanceof Error ? error.message : 'Upload failed';
      this.requestUpdate();
    }
  }

  private pauseUpload(id: string) {
    const item = this.uploadQueue.find((entry) => entry.id === id);
    if (!item || item.status !== 'uploading') {
      return;
    }

    item.abortController?.abort();
    item.status = 'paused';
    this.isUploading = false;
    this.requestUpdate();
  }

  private resumeUpload(id: string) {
    const item = this.uploadQueue.find((entry) => entry.id === id);
    if (!item || item.status !== 'paused') {
      return;
    }

    item.status = 'pending';
    item.progress = 0;
    item.error = undefined;
    this.requestUpdate();
    this.startNextUpload();
  }

  private retryUpload(id: string) {
    const item = this.uploadQueue.find((entry) => entry.id === id);
    if (!item || item.status !== 'error') {
      return;
    }

    item.status = 'pending';
    item.progress = 0;
    item.error = undefined;
    item.abortController = undefined;
    item.uploadResponse = undefined;
    item.processResponse = undefined;
    this.requestUpdate();
    this.startNextUpload();
  }

  private pauseAll() {
    if (this.isUploading) {
      this.uploadQueue.forEach((item) => {
        if (item.status === 'uploading') {
          item.abortController?.abort();
          item.status = 'paused';
        }
      });
      this.isUploading = false;
    } else {
      this.uploadQueue.forEach((item) => {
        if (item.status === 'paused') {
          item.status = 'pending';
          item.progress = 0;
          item.error = undefined;
        }
      });
      this.startNextUpload();
    }

    this.requestUpdate();
  }

  private removeFromQueue(id: string) {
    const item = this.uploadQueue.find((entry) => entry.id === id);
    if (item && (item.status === 'uploading' || item.status === 'processing')) {
      item.abortController?.abort();
    }

    this.uploadQueue = this.uploadQueue.filter((entry) => entry.id !== id);
  }

  private clearCompleted() {
    this.uploadQueue = this.uploadQueue.filter((item) => item.status !== 'completed');
  }

  private clearAll() {
    this.uploadQueue.forEach((item) => {
      if (item.status === 'uploading' || item.status === 'processing') {
        item.abortController?.abort();
      }
    });

    this.uploadQueue = [];
    this.isUploading = false;
  }

  private getSummary(): UploadSummary {
    return this.uploadQueue.reduce<UploadSummary>(
      (summary, item) => {
        summary.total += 1;
        if (item.status === 'completed') {
          summary.completed += 1;
          if (item.processResponse?.duplicate) {
            summary.duplicates += 1;
          }
        } else if (item.status === 'error') {
          summary.failed += 1;
        } else if (item.status === 'processing' || item.status === 'uploading') {
          summary.processing += 1;
        } else {
          summary.pending += 1;
        }
        return summary;
      },
      { total: 0, completed: 0, duplicates: 0, failed: 0, processing: 0, pending: 0 },
    );
  }

  private describeQueueState(summary: UploadSummary): string {
    if (summary.processing > 0) {
      return `${summary.processing} active • ${summary.pending} waiting`;
    }
    if (summary.failed > 0) {
      return `${summary.failed} need attention`;
    }
    if (summary.completed === summary.total && summary.total > 0) {
      return 'All files processed';
    }
    return `${summary.pending} waiting`;
  }

  private getStatusIcon(status: UploadItemStatus): string {
    switch (status) {
      case 'uploading':
        return '⬆️';
      case 'processing':
        return '⚙️';
      case 'completed':
        return '✅';
      case 'error':
        return '❌';
      case 'paused':
        return '⏸️';
      default:
        return '⏳';
    }
  }

  private getStatusClass(status: UploadItemStatus): string {
    return `status-${status}`;
  }

  private getProgressLabel(status: UploadItemStatus): string {
    switch (status) {
      case 'uploading':
        return 'uploading';
      case 'processing':
        return 'processing';
      case 'completed':
        return 'done';
      case 'error':
        return 'failed';
      case 'paused':
        return 'paused';
      default:
        return 'queued';
    }
  }

  private formatFileSize(bytes: number): string {
    if (bytes === 0) {
      return '0 B';
    }

    const units = ['B', 'KB', 'MB', 'GB'];
    const unitIndex = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
    return `${parseFloat((bytes / 1024 ** unitIndex).toFixed(1))} ${units[unitIndex]}`;
  }

  private formatErrorMessage(error: string): string {
    if (error.includes('Failed to upload chunk')) {
      return 'Chunk transfer failed. Retry the file after checking the connection.';
    }

    if (error.includes('Failed to finalize upload')) {
      return 'The file uploaded, but Sortify could not finalize organization.';
    }

    if (error.includes('Upload cancelled')) {
      return 'Upload paused before completion.';
    }

    if (error.length > 120) {
      return `${error.slice(0, 120)}...`;
    }

    return error;
  }

  private getThumbnailUrl(file: File): string {
    const cacheKey = `${file.name}-${file.size}-${file.lastModified}`;
    if (this.thumbnailCache.has(cacheKey)) {
      return this.thumbnailCache.get(cacheKey)!;
    }

    const url = URL.createObjectURL(file);
    this.thumbnailCache.set(cacheKey, url);
    return url;
  }

  private handleThumbnailError = (event: Event) => {
    const img = event.target as HTMLImageElement;
    const fileIcon = img.closest('.file-icon');
    if (fileIcon) {
      fileIcon.innerHTML = '🖼️';
    }
  };

  disconnectedCallback() {
    super.disconnectedCallback();
    for (const url of this.thumbnailCache.values()) {
      URL.revokeObjectURL(url);
    }
    this.thumbnailCache.clear();
  }
}
