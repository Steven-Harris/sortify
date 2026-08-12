import { LitElement, html, svg, unsafeCSS } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { apiService, type ProcessResponse, type UploadResponse } from '../services/api.js';
import { generateUUID } from '../utils/uuid.js';
import uploadStyles from '../styles/upload.css?inline';

const icon = (paths: unknown, strokeWidth = 2) => html`
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="${strokeWidth}"
    stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    ${paths}
  </svg>
`;

const ICONS = {
  upload: svg`<path d="M12 16V4" /><path d="m7 9 5-5 5 5" /><path d="M4 16v2a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-2" />`,
  image: svg`<rect x="3" y="4" width="18" height="16" rx="2" /><circle cx="8.5" cy="9.5" r="1.5" /><path d="m21 16-5-5L6 20" />`,
  video: svg`<rect x="2" y="5" width="14" height="14" rx="2" /><path d="m22 8-6 4 6 4Z" />`,
  file: svg`<path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8Z" /><path d="M14 3v5h5" />`,
  pause: svg`<path d="M10 5v14" /><path d="M14 5v14" />`,
  play: svg`<path d="M7 4.5v15l12-7.5Z" />`,
  retry: svg`<path d="M21 12a9 9 0 1 1-3-6.7" /><path d="M21 4v5h-5" />`,
  trash: svg`<path d="M4 7h16" /><path d="M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2" /><path d="M6 7l1 13h10l1-13" />`,
  broom: svg`<path d="M4 20h16" /><path d="M9 16 5 20" /><path d="M15 4 9 10l5 5 6-6Z" />`,
  clock: svg`<circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" />`,
  check: svg`<path d="M20 6 9 17l-5-5" />`,
  alert: svg`<path d="M12 8v5" /><path d="M12 17h.01" /><circle cx="12" cy="12" r="9" />`,
  gear: svg`<circle cx="12" cy="12" r="3" /><path d="M12 2v3M12 19v3M4.2 4.2l2.1 2.1M17.7 17.7l2.1 2.1M2 12h3M19 12h3M4.2 19.8l2.1-2.1M17.7 6.3l2.1-2.1" />`,
  folder: svg`<path d="M4 7V5.5A1.5 1.5 0 0 1 5.5 4H9l2 2h7.5A1.5 1.5 0 0 1 20 7.5V8" /><path d="M3 10h18l-1.4 8.3a2 2 0 0 1-2 1.7H6.4a2 2 0 0 1-2-1.7Z" />`,
} as const;

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
              ${icon(ICONS.upload, 2.2)}
            </div>
            <h3 class="dropzone-title">Bulk upload your photo and video library</h3>
            <p class="dropzone-subtitle">
              Sortify treats uploads as a queue, verifies each file, skips exact duplicates, and
              organizes by trusted metadata first.
            </p>
            <button class="btn-primary" ?disabled=${this.disabled}>
              Choose files
            </button>
            <p class="dropzone-hint">
              or drop them anywhere in this panel — <kbd>JPG</kbd> <kbd>HEIC</kbd> <kbd>PNG</kbd>
              <kbd>MP4</kbd> <kbd>MOV</kbd>
            </p>
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
            <div class="summary-card accent-neutral">
              <span class="summary-value">${summary.total}</span>
              <span class="summary-label">Queued</span>
            </div>
            <div class="summary-card accent-success">
              <span class="summary-value">${summary.completed}</span>
              <span class="summary-label">Stored</span>
            </div>
            <div class="summary-card accent-info">
              <span class="summary-value">${summary.duplicates}</span>
              <span class="summary-label">Duplicates skipped</span>
            </div>
            <div class="summary-card accent-danger">
              <span class="summary-value">${summary.failed}</span>
              <span class="summary-label">Failed</span>
            </div>
          </div>
        ` : null}

        ${this.uploadQueue.length > 0 ? html`
          <div class="upload-queue">
            <div class="queue-header">
              <h4 class="queue-title">
                Upload queue
                <span class="queue-subtitle">
                  ${this.describeQueueState(summary)}
                </span>
              </h4>
              <div class="queue-actions">
                <button class="btn-secondary" @click=${this.pauseAll}>
                  ${this.isUploading ? icon(ICONS.pause) : icon(ICONS.play)}
                  ${this.isUploading ? 'Pause queue' : 'Resume queue'}
                </button>
                <button class="btn-secondary" @click=${this.clearCompleted}>
                  ${icon(ICONS.broom)} Clear completed
                </button>
                <button class="btn-danger" @click=${this.clearAll}>
                  ${icon(ICONS.trash)} Clear all
                </button>
              </div>
            </div>

            ${this.uploadQueue.map((item) => this.renderUploadItem(item))}
          </div>
        ` : html`
          <div class="empty-state">
            <div class="empty-state-icon">${icon(ICONS.folder)}</div>
            <p class="empty-state-text">Nothing in the queue yet</p>
            <p class="empty-state-hint">
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
              <span class="error-icon">${icon(ICONS.alert)}</span>
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
          <span>${this.getProgressLabel(item.status)}</span>
        </div>

        <div class="item-actions">
          ${item.status === 'uploading' ? html`
            <button class="btn-icon btn-icon-neutral" @click=${() => this.pauseUpload(item.id)}
              title="Pause upload" aria-label="Pause upload">
              ${icon(ICONS.pause)}
            </button>
          ` : null}
          ${item.status === 'paused' ? html`
            <button class="btn-icon btn-icon-success" @click=${() => this.resumeUpload(item.id)}
              title="Resume upload" aria-label="Resume upload">
              ${icon(ICONS.play)}
            </button>
          ` : null}
          ${item.status === 'error' ? html`
            <button class="btn-icon btn-icon-success" @click=${() => this.retryUpload(item.id)}
              title="Retry upload" aria-label="Retry upload">
              ${icon(ICONS.retry)}
            </button>
          ` : null}
          ${item.status !== 'completed' ? html`
            <button class="btn-icon btn-icon-danger" @click=${() => this.removeFromQueue(item.id)}
              title="Remove from queue" aria-label="Remove from queue">
              ${icon(ICONS.trash)}
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
      return html`<div class="file-icon video">${icon(ICONS.video)}</div>`;
    }

    return html`<div class="file-icon">${icon(ICONS.file)}</div>`;
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

  private getStatusIcon(status: UploadItemStatus) {
    switch (status) {
      case 'uploading':
        return icon(ICONS.upload);
      case 'processing':
        return icon(ICONS.gear);
      case 'completed':
        return icon(ICONS.check);
      case 'error':
        return icon(ICONS.alert);
      case 'paused':
        return icon(ICONS.pause);
      default:
        return icon(ICONS.clock);
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
      fileIcon.innerHTML =
        '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" ' +
        'stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">' +
        '<rect x="3" y="4" width="18" height="16" rx="2"></rect>' +
        '<circle cx="8.5" cy="9.5" r="1.5"></circle>' +
        '<path d="m21 16-5-5L6 20"></path></svg>';
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
