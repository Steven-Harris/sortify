import { LitElement, css, html } from 'lit';
import { customElement, state, property } from 'lit/decorators.js';
import { theme, buttonStyles, cardStyles } from '../styles/theme.js';
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
  static styles = [
    theme,
    buttonStyles,
    cardStyles,
    css`
      :host {
        display: block;
      }

      .upload-container {
        max-width: 800px;
        margin: 0 auto;
      }

      .dropzone {
        border: 2px dashed var(--color-border);
        border-radius: var(--border-radius-lg);
        padding: var(--spacing-xxl);
        text-align: center;
        transition: all var(--transition-normal);
        cursor: pointer;
        background-color: var(--color-surface);
      }

      .dropzone:hover,
      .dropzone.drag-over {
        border-color: var(--color-primary);
        background-color: rgba(22, 71, 115, 0.1);
      }

      .dropzone-content {
        pointer-events: none;
      }

      .dropzone-icon {
        font-size: 4rem;
        margin-bottom: var(--spacing-md);
        opacity: 0.7;
      }

      .dropzone-title {
        font-size: var(--font-size-xl);
        font-weight: var(--font-weight-semibold);
        margin-bottom: var(--spacing-sm);
        color: var(--color-text-primary);
      }

      .dropzone-subtitle {
        color: var(--color-text-secondary);
        margin-bottom: var(--spacing-lg);
      }

      .file-input {
        display: none;
      }

      .upload-queue {
        margin-top: var(--spacing-xl);
      }

      .queue-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: var(--spacing-md);
      }

      .queue-title {
        font-size: var(--font-size-lg);
        font-weight: var(--font-weight-semibold);
        margin: 0;
      }

      .queue-actions {
        display: flex;
        gap: var(--spacing-sm);
      }

      .upload-item {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
        padding: var(--spacing-md);
        background-color: var(--color-surface-elevated);
        border: 1px solid var(--color-border);
        border-radius: var(--border-radius-md);
        margin-bottom: var(--spacing-sm);
      }

      .file-info {
        flex: 1;
        min-width: 0;
      }

      .file-name {
        font-weight: var(--font-weight-medium);
        margin-bottom: 4px;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }

      .file-details {
        font-size: var(--font-size-sm);
        color: var(--color-text-secondary);
      }

      .progress-container {
        flex: 0 0 200px;
      }

      .progress-bar {
        width: 100%;
        height: 6px;
        background-color: var(--color-surface);
        border-radius: 3px;
        overflow: hidden;
        margin-bottom: 4px;
      }

      .progress-fill {
        height: 100%;
        background: linear-gradient(90deg, var(--color-primary), var(--color-accent-2));
        transition: width var(--transition-normal);
      }

      .progress-text {
        font-size: var(--font-size-xs);
        color: var(--color-text-secondary);
        text-align: center;
      }

      .status-icon {
        flex: 0 0 auto;
        font-size: var(--font-size-lg);
      }

      .item-actions {
        flex: 0 0 auto;
        display: flex;
        gap: var(--spacing-xs);
      }

      .status-pending { color: var(--color-text-secondary); }
      .status-uploading { color: var(--color-primary); }
      .status-completed { color: var(--color-success); }
      .status-error { color: var(--color-error); }
      .status-paused { color: var(--color-warning); }

      .empty-state {
        text-align: center;
        padding: var(--spacing-lg);
        color: var(--color-text-secondary);
      }

      @media (max-width: 640px) {
        .upload-item {
          flex-direction: column;
          align-items: stretch;
        }

        .progress-container {
          flex: 1;
        }

        .queue-header {
          flex-direction: column;
          align-items: stretch;
          gap: var(--spacing-sm);
        }
      }
    `
  ];

  @property({ type: Boolean })
  disabled = false;

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
            <div class="dropzone-icon">📁</div>
            <h3 class="dropzone-title">Drop your photos and videos here</h3>
            <p class="dropzone-subtitle">
              Or click to browse your files
            </p>
            <button 
              class="btn btn-primary btn-lg"
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
                Upload Queue (${this.uploadQueue.length} files)
              </h4>
              <div class="queue-actions">
                <button 
                  class="btn btn-secondary btn-sm"
                  @click=${this.pauseAll}
                  ?disabled=${!this.isUploading}
                >
                  ${this.isUploading ? 'Pause All' : 'Resume All'}
                </button>
                <button 
                  class="btn btn-secondary btn-sm"
                  @click=${this.clearCompleted}
                >
                  Clear Completed
                </button>
                <button 
                  class="btn btn-secondary btn-sm"
                  @click=${this.clearAll}
                >
                  Clear All
                </button>
              </div>
            </div>

            ${this.uploadQueue.map(item => this.renderUploadItem(item))}
          </div>
        ` : html`
          <div class="empty-state">
            <p>No files selected. Drop some files or click above to get started!</p>
          </div>
        `}
      </div>
    `;
  }

  private renderUploadItem(item: UploadFile) {
    const statusIcons = {
      pending: '⏳',
      uploading: '⬆️',
      completed: '✅',
      error: '❌',
      paused: '⏸️',
      processing: '🔄'
    };

    return html`
      <div class="upload-item">
        <div class="file-info">
          <div class="file-name" title=${item.file.name}>
            ${item.file.name}
          </div>
          <div class="file-details">
            ${this.formatFileSize(item.file.size)} • ${item.file.type || 'Unknown type'}
            ${item.error ? html` • ${item.error}` : ''}
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
            ${item.progress}% ${item.status === 'uploading' ? 'uploading' : ''}
          </div>
        </div>

        <div class="status-icon status-${item.status}">
          ${statusIcons[item.status]}
        </div>

        <div class="item-actions">
          ${item.status === 'uploading' ? html`
            <button 
              class="btn btn-secondary btn-sm"
              @click=${() => this.pauseUpload(item.id)}
              title="Pause upload"
            >
              ⏸️
            </button>
          ` : item.status === 'paused' ? html`
            <button 
              class="btn btn-secondary btn-sm"
              @click=${() => this.resumeUpload(item.id)}
              title="Resume upload"
            >
              ▶️
            </button>
          ` : ''}
          
          <button 
            class="btn btn-secondary btn-sm"
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
    
    // Start uploading automatically
    this.startNextUpload();
  }

  private isValidFile(file: File): boolean {
    // Check file type
    if (!file.type.startsWith('image/') && !file.type.startsWith('video/')) {
      console.warn('Invalid file type:', file.type);
      return false;
    }

    // Check file size (100MB limit)
    const maxSize = 100 * 1024 * 1024;
    if (file.size > maxSize) {
      console.warn('File too large:', file.size);
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
      // Create abort controller for this upload
      item.abortController = new AbortController();
      
      // Upload the file using the API service
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

      // Store upload response
      item.uploadResponse = uploadResponse;
      item.uploadId = uploadResponse.id;
      
      // Start processing the uploaded file
      item.status = 'processing';
      item.progress = 100;
      this.requestUpdate();

      try {
        const processResponse = await apiService.processFile(uploadResponse.id);
        item.processResponse = processResponse;
        item.processId = processResponse.id;
        
        // Poll for processing completion
        await this.waitForProcessing(item);
        
      } catch (processError) {
        item.status = 'error';
        item.error = processError instanceof Error ? processError.message : 'Processing failed';
        this.requestUpdate();
      }

    } catch (uploadError) {
      if (uploadError instanceof Error && uploadError.message === 'Upload cancelled') {
        // Upload was cancelled, don't mark as error
        return;
      }
      
      item.status = 'error';
      item.error = uploadError instanceof Error ? uploadError.message : 'Upload failed';
      this.requestUpdate();
    }
  }

  private async waitForProcessing(item: UploadFile) {
    if (!item.processId) return;

    const maxAttempts = 60; // 5 minutes max
    let attempts = 0;

    while (attempts < maxAttempts) {
      try {
        const status = await apiService.getProcessStatus(item.processId);
        item.processResponse = status;

        if (status.status === 'completed') {
          item.status = 'completed';
          this.requestUpdate();
          return;
        }

        if (status.status === 'error') {
          item.status = 'error';
          item.error = status.error || 'Processing failed';
          this.requestUpdate();
          return;
        }

        // Still processing, wait and try again
        await new Promise(resolve => setTimeout(resolve, 5000));
        attempts++;

      } catch (error) {
        attempts++;
        if (attempts >= maxAttempts) {
          item.status = 'error';
          item.error = 'Processing timeout';
          this.requestUpdate();
          return;
        }
        
        // Wait before retrying
        await new Promise(resolve => setTimeout(resolve, 5000));
      }
    }

    // Timeout
    item.status = 'error';
    item.error = 'Processing timeout';
    this.requestUpdate();
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
