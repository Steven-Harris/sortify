import { LitElement, html, unsafeCSS } from 'lit';
import { customElement, state, property } from 'lit/decorators.js';
import { apiService, type UploadResponse, type ProcessResponse } from '../services/api.js';
import { generateUUID } from '../utils/uuid.js';
import uploadStyles from '../styles/upload.css?inline';

export interface UploadFile {
  id: string;
  file: File;
  progress: number;
  status: 'pending' | 'uploading' | 'completed' | 'error' | 'paused' | 'processing';
  uploadId?: string;
  processId?: string;
  uploadResponse?: UploadResponse;
  processResponse?: ProcessResponse;
  error?: string;
  abortController?: AbortController;
}

/**
 * Upload Component with Drag & Drop
 * Handles file selection, upload queue, and progress tracking
 */
@customElement('sortify-upload')
export class SortifyUpload extends LitElement {
  static styles = unsafeCSS(uploadStyles);

  @property({ type: Boolean })
  disabled = false;

  @property({ type: Number })
  maxFileSize = 0; // 0 = no limit

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
            <h3 class="dropzone-title">Drop your photos and videos here</h3>
            <p class="dropzone-subtitle">
              Supports JPG, PNG, GIF, MP4, MOV and more. No file size limit.
            </p>
            <button 
              class="btn-primary"
              ?disabled=${this.disabled}
            >
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

        ${this.uploadQueue.length > 0 ? html`
          <div class="upload-queue">
            <div class="queue-header">
              <h4 class="queue-title">
                Upload Queue
                <span style="font-weight: 400; color: var(--gray-500); font-size: 1rem;">
                  ${this.uploadQueue.length} ${this.uploadQueue.length === 1 ? 'file' : 'files'}
                </span>
              </h4>
              <div class="queue-actions">
                <button 
                  class="btn-secondary"
                  @click=${this.pauseAll}
                >
                  ${this.isUploading ? '⏸️ Pause All' : '▶️ Resume All'}
                </button>
                <button 
                  class="btn-secondary"
                  @click=${this.clearCompleted}
                >
                  🗑️ Clear Completed
                </button>
                <button 
                  class="btn-danger"
                  @click=${this.clearAll}
                >
                  ❌ Clear All
                </button>
              </div>
            </div>

            ${this.uploadQueue.map(item => this.renderUploadItem(item))}
          </div>
        ` : html`
          <div class="empty-state">
            <div class="empty-state-icon">📂</div>
            <p class="empty-state-text">No files selected yet</p>
            <p style="font-size: 0.875rem; margin-top: 0.5rem; opacity: 0.8;">
              Drop some files above or click to browse your computer
            </p>
          </div>
        `}
      </div>
    `;
  }

  private renderUploadItem(item: UploadFile) {
    const getFileIcon = (file: File) => {
      if (file.type.startsWith('image/')) {
        return html`<div class="file-icon image">
          <img 
            src="${this.getThumbnailUrl(file)}" 
            alt="${file.name}"
            class="thumbnail-image"
            @error=${() => this.handleThumbnailError}
          />
        </div>`;
      } else if (file.type.startsWith('video/')) {
        return html`<div class="file-icon video">🎬</div>`;
      }
      return html`<div class="file-icon">📄</div>`;
    };

    const getStatusBadge = (status: string) => {
      const badges = {
        pending: { icon: '⏳', class: 'status-pending' },
        uploading: { icon: '⬆️', class: 'status-uploading' },
        processing: { icon: '⚙️', class: 'status-processing' },
        completed: { icon: '✅', class: 'status-completed' },
        error: { icon: '❌', class: 'status-error' },
        paused: { icon: '⏸️', class: 'status-paused' }
      };
      
      const badge = badges[status as keyof typeof badges] || badges.pending;
      return html`
        <div class="status-badge ${badge.class}" title="${status}">
          ${badge.icon}
        </div>
      `;
    };

    return html`
      <div class="upload-item">
        ${getFileIcon(item.file)}
        
        <div class="file-info">
          <div class="file-name" title="${item.file.name}">
            ${item.file.name}
          </div>
          <div class="file-details">
            ${this.formatFileSize(item.file.size)} • ${item.file.type.split('/')[0] || 'Unknown'}
            ${item.error ? html` • <span style="color: var(--red-400);">${item.error}</span>` : ''}
          </div>
        </div>

        ${item.status !== 'completed' ? html`
          <div class="progress-container">
            <div class="progress-bar">
              <div 
                class="progress-fill" 
                style="width: ${item.progress}%"
              ></div>
            </div>
            <div class="progress-text">
              ${item.progress}% 
              ${item.status === 'uploading' ? 'uploading' : ''}
              ${item.status === 'processing' ? 'processing' : ''}
            </div>
          </div>
        ` : ''}

        ${getStatusBadge(item.status)}

        <div class="item-actions">
          ${item.status === 'uploading' ? html`
            <button 
              class="btn-icon btn-icon-neutral"
              @click=${() => this.pauseUpload(item.id)}
              title="Pause upload"
            >
              ⏸️
            </button>
          ` : item.status === 'paused' ? html`
            <button 
              class="btn-icon btn-icon-success"
              @click=${() => this.resumeUpload(item.id)}
              title="Resume upload"
            >
              ▶️
            </button>
          ` : item.status === 'error' ? html`
            <button 
              class="btn-icon btn-icon-success"
              @click=${() => this.retryUpload(item.id)}
              title="Retry upload"
            >
              🔄
            </button>
          ` : ''}
          
          ${item.status !== 'completed' ? html`
            <button 
              class="btn-icon btn-icon-danger"
              @click=${() => this.removeFromQueue(item.id)}
              title="Remove from queue"
            >
              🗑️
            </button>
          ` : ''}
        </div>
      </div>
    `;
  }

  private openFileDialog() {
    if (this.disabled) return;
    
    if (!this.fileInputRef) {
      this.fileInputRef = this.shadowRoot?.querySelector('.file-input') as HTMLInputElement;
    }
    
    this.fileInputRef?.click();
  }

  private handleDragOver(e: DragEvent) {
    e.preventDefault();
    e.stopPropagation();
    if (!this.isDragOver) {
      this.isDragOver = true;
    }
  }

  private handleDragLeave(e: DragEvent) {
    e.preventDefault();
    e.stopPropagation();
    this.isDragOver = false;
  }

  private handleDrop(e: DragEvent) {
    e.preventDefault();
    e.stopPropagation();
    this.isDragOver = false;

    if (this.disabled) return;

    const files = Array.from(e.dataTransfer?.files || []);
    this.addFilesToQueue(files);
  }

  private handleFileSelect(e: Event) {
    const input = e.target as HTMLInputElement;
    const files = Array.from(input.files || []);
    this.addFilesToQueue(files);
    
    // Reset input so same files can be selected again
    input.value = '';
  }

  private addFilesToQueue(files: File[]) {
    const newItems: UploadFile[] = files
      .filter(file => this.isValidFile(file))
      .map(file => ({
        id: generateUUID(),
        file,
        progress: 0,
        status: 'pending'
      }));

    this.uploadQueue = [...this.uploadQueue, ...newItems];
    
    this.startNextUpload();
  }

  private isValidFile(file: File): boolean {
    if (!file.type.startsWith('image/') && !file.type.startsWith('video/')) {
      console.warn('Invalid file type:', file.type);
      return false;
    }

    // Check file size using configurable limit
    if (this.maxFileSize > 0 && file.size > this.maxFileSize) {
      console.warn('File too large:', file.size, 'Max allowed:', this.maxFileSize);
      return false;
    }

    return true;
  }

  private async startNextUpload() {
    if (this.isUploading) return;

    const nextItem = this.uploadQueue.find(item => item.status === 'pending');
    if (!nextItem) return;

    this.isUploading = true;
    nextItem.status = 'uploading';
    this.requestUpdate();

    try {
      await this.uploadFile(nextItem);
      // Status is set within uploadFile method based on actual outcome
    } catch (error) {
      nextItem.status = 'error';
      nextItem.error = error instanceof Error ? error.message : 'Upload failed';
    }

    this.isUploading = false;
    this.requestUpdate();

    // Start next upload
    setTimeout(() => this.startNextUpload(), 100);
  }

  private async uploadFile(item: UploadFile) {
    try {
      item.abortController = new AbortController();
      
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
        signal: item.abortController.signal
      });

      item.uploadResponse = uploadResponse;
      item.uploadId = uploadResponse.id;
      
      item.status = 'processing';
      item.progress = 100;
      this.requestUpdate();

      try {
        item.status = 'completed';
        item.processResponse = {
          id: uploadResponse.sessionId ?? uploadResponse.id ?? '',
          originalPath: uploadResponse.filename,
          organizedPath: uploadResponse.filename,
          metadata: uploadResponse.mediaInfo || {},
          status: 'completed'
        };
        this.requestUpdate();
        
      } catch (processError) {
        item.status = 'error';
        item.error = processError instanceof Error ? processError.message : 'Processing failed';
        this.requestUpdate();
      }

    } catch (uploadError) {
      if (uploadError instanceof Error && uploadError.message === 'Upload cancelled') {
        return;
      }
      
      item.status = 'error';
      item.error = uploadError instanceof Error ? uploadError.message : 'Upload failed';
      this.requestUpdate();
    }
  }

  private pauseUpload(id: string) {
    const item = this.uploadQueue.find(item => item.id === id);
    if (item && item.status === 'uploading') {
      // Abort the current upload
      item.abortController?.abort();
      item.status = 'paused';
      this.isUploading = false;
      this.requestUpdate();
    }
  }

  private resumeUpload(id: string) {
    const item = this.uploadQueue.find(item => item.id === id);
    if (item && item.status === 'paused') {
      item.status = 'pending';
      item.progress = 0; // Reset progress for retry
      item.error = undefined;
      this.requestUpdate();
      this.startNextUpload();
    }
  }

  private retryUpload(id: string) {
    const item = this.uploadQueue.find(item => item.id === id);
    if (item && item.status === 'error') {
      item.status = 'pending';
      item.progress = 0; // Reset progress for retry
      item.error = undefined;
      item.abortController = undefined; // Clear old abort controller
      this.requestUpdate();
      this.startNextUpload();
    }
  }

  private pauseAll() {
    if (this.isUploading) {
      // Pause current uploads
      this.uploadQueue.forEach(item => {
        if (item.status === 'uploading') {
          item.abortController?.abort();
          item.status = 'paused';
        }
      });
      this.isUploading = false;
    } else {
      // Resume all paused uploads
      this.uploadQueue.forEach(item => {
        if (item.status === 'paused') {
          item.status = 'pending';
          item.progress = 0; // Reset progress for retry
          item.error = undefined;
        }
      });
      this.startNextUpload();
    }
    this.requestUpdate();
  }

  private removeFromQueue(id: string) {
    const item = this.uploadQueue.find(item => item.id === id);
    if (item) {
      // Abort upload if it's in progress
      if (item.status === 'uploading' || item.status === 'processing') {
        item.abortController?.abort();
      }
    }
    this.uploadQueue = this.uploadQueue.filter(item => item.id !== id);
  }

  private clearCompleted() {
    this.uploadQueue = this.uploadQueue.filter(item => item.status !== 'completed');
  }

  private clearAll() {
    // Abort any ongoing uploads
    this.uploadQueue.forEach(item => {
      if (item.status === 'uploading' || item.status === 'processing') {
        item.abortController?.abort();
      }
    });
    
    this.uploadQueue = [];
    this.isUploading = false;
  }

  private formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
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

  private handleThumbnailError = (e: Event) => {
    const img = e.target as HTMLImageElement;
    const fileIcon = img.closest('.file-icon');
    if (fileIcon) {
      fileIcon.innerHTML = '🖼️';
    }
  };

  disconnectedCallback() {
    super.disconnectedCallback();
    // Clean up object URLs to prevent memory leaks
    for (const url of this.thumbnailCache.values()) {
      URL.revokeObjectURL(url);
    }
    this.thumbnailCache.clear();
  }
}
