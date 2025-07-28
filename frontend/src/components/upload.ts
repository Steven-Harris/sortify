import { LitElement, css, html } from 'lit';
import { customElement, state, property } from 'lit/decorators.js';
import { apiService, type UploadResponse, type ProcessResponse } from '../services/api.js';

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
  static styles = css`
    @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
    
    :host {
      display: block;
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      --primary-400: #60a5fa;
      --primary-500: #3b82f6;
      --primary-600: #2563eb;
      --slate-50: #f8fafc;
      --slate-100: #f1f5f9;
      --slate-200: #e2e8f0;
      --slate-300: #cbd5e1;
      --slate-400: #94a3b8;
      --slate-500: #64748b;
      --slate-600: #475569;
      --slate-700: #334155;
      --slate-800: #1e293b;
      --slate-900: #0f172a;
      --green-400: #4ade80;
      --green-500: #22c55e;
      --red-400: #f87171;
      --red-500: #ef4444;
      --yellow-400: #facc15;
      --yellow-500: #eab308;
    }

    .upload-container {
      max-width: 56rem; /* max-w-4xl */
      margin: 0 auto;
      padding: 1.5rem; /* p-6 */
    }

    .dropzone {
      position: relative;
      border: 2px dashed var(--slate-600);
      border-radius: 1rem; /* rounded-2xl */
      padding: 3rem 2rem; /* py-12 px-8 */
      text-align: center;
      transition: all 0.2s ease-in-out;
      cursor: pointer;
      background: linear-gradient(135deg, var(--slate-800) 0%, var(--slate-700) 100%);
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3);
    }

    .dropzone:hover,
    .dropzone.drag-over {
      border-color: var(--primary-500);
      background: linear-gradient(135deg, var(--slate-700) 0%, var(--slate-600) 100%);
      transform: translateY(-2px);
      box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.4), 0 10px 10px -5px rgba(0, 0, 0, 0.2);
    }

    .dropzone-content {
      pointer-events: none;
    }

    .dropzone-icon {
      width: 4rem; /* w-16 */
      height: 4rem; /* h-16 */
      margin: 0 auto 1.5rem; /* mx-auto mb-6 */
      background: linear-gradient(135deg, var(--primary-500), var(--primary-400));
      border-radius: 1rem; /* rounded-2xl */
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 1.5rem; /* text-2xl */
      color: white;
      box-shadow: 0 8px 25px 0 rgba(59, 130, 246, 0.4);
    }

    .dropzone-title {
      font-size: 1.875rem; /* text-3xl */
      font-weight: 700; /* font-bold */
      margin-bottom: 0.75rem; /* mb-3 */
      color: var(--slate-100);
      line-height: 1.2;
    }

    .dropzone-subtitle {
      color: var(--slate-400);
      margin-bottom: 2rem; /* mb-8 */
      font-size: 1.125rem; /* text-lg */
      font-weight: 400;
      line-height: 1.6;
    }

    .upload-button {
      background: linear-gradient(135deg, var(--primary-600), var(--primary-500));
      color: white;
      border: none;
      border-radius: 0.75rem; /* rounded-xl */
      padding: 0.875rem 2rem; /* py-3.5 px-8 */
      font-size: 1rem; /* text-base */
      font-weight: 600; /* font-semibold */
      cursor: pointer;
      transition: all 0.2s ease-in-out;
      box-shadow: 0 8px 25px 0 rgba(59, 130, 246, 0.4);
      font-family: inherit;
    }

    .upload-button:hover {
      transform: translateY(-1px);
      box-shadow: 0 20px 25px -5px rgba(59, 130, 246, 0.6);
    }

    .upload-button:disabled {
      opacity: 0.6;
      cursor: not-allowed;
      transform: none;
    }

    .file-input {
      display: none;
    }

    .upload-queue {
      margin-top: 2.5rem; /* mt-10 */
    }

    .queue-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1.5rem; /* mb-6 */
      padding-bottom: 1rem; /* pb-4 */
      border-bottom: 1px solid var(--slate-700);
    }

    .queue-title {
      font-size: 1.5rem; /* text-2xl */
      font-weight: 700; /* font-bold */
      margin: 0;
      color: var(--slate-100);
    }

    .queue-actions {
      display: flex;
      gap: 0.75rem; /* gap-3 */
    }

    .btn {
      border: none;
      border-radius: 0.5rem; /* rounded-lg */
      padding: 0.5rem 1rem; /* py-2 px-4 */
      font-size: 0.875rem; /* text-sm */
      font-weight: 500; /* font-medium */
      cursor: pointer;
      transition: all 0.2s ease-in-out;
      font-family: inherit;
    }

    .btn-secondary {
      background: var(--slate-700);
      color: var(--slate-300);
      border: 1px solid var(--slate-600);
    }

    .btn-secondary:hover {
      background: var(--slate-600);
      color: var(--slate-100);
      transform: translateY(-1px);
    }

    .upload-item {
      display: flex;
      align-items: center;
      gap: 1rem; /* gap-4 */
      padding: 1.5rem; /* p-6 */
      background: var(--slate-800);
      border: 1px solid var(--slate-700);
      border-radius: 1rem; /* rounded-2xl */
      margin-bottom: 0.75rem; /* mb-3 */
      transition: all 0.2s ease-in-out;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3);
    }

    .upload-item:hover {
      transform: translateY(-1px);
      box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.4);
      border-color: var(--slate-600);
    }

    .file-icon {
      width: 3rem; /* w-12 */
      height: 3rem; /* h-12 */
      border-radius: 0.75rem; /* rounded-xl */
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 1.25rem; /* text-xl */
      flex-shrink: 0;
    }

    .file-icon.image {
      background: linear-gradient(135deg, var(--green-500), var(--green-400));
      color: white;
    }

    .file-icon.video {
      background: linear-gradient(135deg, #a855f7, #8b5cf6);
      color: white;
    }

    .file-info {
      flex: 1;
      min-width: 0;
    }

    .file-name {
      font-weight: 600; /* font-semibold */
      margin-bottom: 0.25rem; /* mb-1 */
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      color: var(--slate-100);
      font-size: 1rem; /* text-base */
    }

    .file-details {
      font-size: 0.875rem; /* text-sm */
      color: var(--slate-400);
      font-weight: 400;
    }

    .progress-container {
      flex: 0 0 12rem; /* flex-none w-48 */
    }

    .progress-bar {
      width: 100%;
      height: 0.5rem; /* h-2 */
      background: var(--slate-700);
      border-radius: 9999px; /* rounded-full */
      overflow: hidden;
      margin-bottom: 0.5rem; /* mb-2 */
    }

    .progress-fill {
      height: 100%;
      background: linear-gradient(90deg, var(--primary-500), var(--primary-400));
      transition: width 0.3s ease-in-out;
      border-radius: 9999px; /* rounded-full */
    }

    .progress-text {
      font-size: 0.75rem; /* text-xs */
      color: var(--slate-400);
      text-align: center;
      font-weight: 500;
    }

    .status-badge {
      flex: 0 0 auto;
      width: 2.5rem; /* w-10 */
      height: 2.5rem; /* h-10 */
      border-radius: 9999px; /* rounded-full */
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 0.875rem; /* text-sm */
      font-weight: 500;
    }

    .status-pending {
      background: var(--slate-700);
      color: var(--slate-400);
    }

    .status-uploading {
      background: rgba(59, 130, 246, 0.2);
      color: var(--primary-400);
      animation: pulse 2s infinite;
    }

    .status-processing {
      background: rgba(59, 130, 246, 0.2);
      color: var(--primary-400);
      animation: spin 1s linear infinite;
    }

    .status-completed {
      background: rgba(34, 197, 94, 0.2);
      color: var(--green-400);
    }

    .status-error {
      background: rgba(239, 68, 68, 0.2);
      color: var(--red-400);
    }

    .status-paused {
      background: rgba(234, 179, 8, 0.2);
      color: var(--yellow-400);
    }

    .item-actions {
      flex: 0 0 auto;
      display: flex;
      gap: 0.5rem; /* gap-2 */
    }

    .action-btn {
      width: 2rem; /* w-8 */
      height: 2rem; /* h-8 */
      border: none;
      border-radius: 0.5rem; /* rounded-lg */
      background: var(--slate-700);
      color: var(--slate-400);
      cursor: pointer;
      transition: all 0.2s ease-in-out;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 0.875rem; /* text-sm */
    }

    .action-btn:hover {
      background: var(--slate-600);
      color: var(--slate-200);
      transform: scale(1.05);
    }

    .empty-state {
      text-align: center;
      padding: 3rem 1rem; /* py-12 px-4 */
      color: var(--slate-400);
      background: var(--slate-800);
      border-radius: 1rem; /* rounded-2xl */
      border: 1px dashed var(--slate-600);
      margin-top: 2rem; /* mt-8 */
    }

    .empty-state-icon {
      font-size: 3rem; /* text-5xl */
      margin-bottom: 1rem; /* mb-4 */
      opacity: 0.6;
    }

    .empty-state-text {
      font-size: 1.125rem; /* text-lg */
      font-weight: 500;
    }

    @keyframes pulse {
      0%, 100% { opacity: 1; }
      50% { opacity: 0.5; }
    }

    @keyframes spin {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
    }

    @media (max-width: 768px) {
      .upload-container {
        padding: 1rem; /* p-4 */
      }

      .dropzone {
        padding: 2rem 1rem; /* py-8 px-4 */
      }

      .dropzone-title {
        font-size: 1.5rem; /* text-2xl */
      }

      .upload-item {
        flex-direction: column;
        align-items: stretch;
        gap: 1rem; /* gap-4 */
      }

      .progress-container {
        flex: 1;
      }

      .queue-header {
        flex-direction: column;
        align-items: stretch;
        gap: 1rem; /* gap-4 */
      }

      .queue-actions {
        justify-content: center;
      }
    }
  `;

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
              class="upload-button"
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
                  class="btn btn-secondary"
                  @click=${this.pauseAll}
                >
                  ${this.isUploading ? '⏸️ Pause All' : '▶️ Resume All'}
                </button>
                <button 
                  class="btn btn-secondary"
                  @click=${this.clearCompleted}
                >
                  🗑️ Clear Completed
                </button>
                <button 
                  class="btn btn-secondary"
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
        return html`<div class="file-icon image">🖼️</div>`;
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

        ${getStatusBadge(item.status)}

        <div class="item-actions">
          ${item.status === 'uploading' ? html`
            <button 
              class="action-btn"
              @click=${() => this.pauseUpload(item.id)}
              title="Pause upload"
            >
              ⏸️
            </button>
          ` : item.status === 'paused' ? html`
            <button 
              class="action-btn"
              @click=${() => this.resumeUpload(item.id)}
              title="Resume upload"
            >
              ▶️
            </button>
          ` : ''}
          
          <button 
            class="action-btn"
            @click=${() => this.removeFromQueue(item.id)}
            title="Remove from queue"
          >
            🗑️
          </button>
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
        id: crypto.randomUUID(),
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
      nextItem.status = 'completed';
      nextItem.progress = 100;
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
}
