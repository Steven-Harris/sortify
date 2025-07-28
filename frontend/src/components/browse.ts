import { LitElement, css, html } from 'lit';
import { customElement, state, property } from 'lit/decorators.js';
import { apiService } from '../services/api.js';

export interface MediaFile {
  id: string;
  filename: string;
  originalPath: string;
  organizedPath: string;
  size: number;
  type: 'image' | 'video';
  metadata: {
    date?: string;
    camera?: string;
    location?: string;
    width?: number;
    height?: number;
    duration?: number;
  };
  thumbnail?: string;
}

export interface FileGroup {
  date: string;
  files: MediaFile[];
}

/**
 * Browse Component with Search
 * Displays organized media files with search and filter capabilities
 */
@customElement('sortify-browse')
export class SortifyBrowse extends LitElement {
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
    }

    .browse-container {
      max-width: 80rem;
      margin: 0 auto;
      padding: 1.5rem;
    }

    .search-bar {
      background: var(--slate-800);
      border: 1px solid var(--slate-700);
      border-radius: 1rem;
      padding: 1.5rem;
      margin-bottom: 2rem;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3);
    }

    .search-input-wrapper {
      display: flex;
      gap: 1rem;
      margin-bottom: 1rem;
    }

    .search-input {
      flex: 1;
      background: var(--slate-700);
      border: 1px solid var(--slate-600);
      border-radius: 0.75rem;
      padding: 0.75rem 1rem;
      color: var(--slate-100);
      font-size: 1rem;
      font-family: inherit;
      transition: all 0.2s ease-in-out;
    }

    .search-input:focus {
      outline: none;
      border-color: var(--primary-500);
      box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
    }

    .search-input::placeholder {
      color: var(--slate-400);
    }

    .search-button {
      background: linear-gradient(135deg, var(--primary-600), var(--primary-500));
      color: white;
      border: none;
      border-radius: 0.75rem;
      padding: 0.75rem 1.5rem;
      font-size: 1rem;
      font-weight: 600;
      cursor: pointer;
      transition: all 0.2s ease-in-out;
      font-family: inherit;
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .search-button:hover {
      transform: translateY(-1px);
      box-shadow: 0 8px 25px 0 rgba(59, 130, 246, 0.4);
    }

    .filters {
      display: flex;
      gap: 1rem;
      flex-wrap: wrap;
    }

    .filter-group {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }

    .filter-label {
      font-size: 0.875rem;
      font-weight: 500;
      color: var(--slate-300);
    }

    .filter-select {
      background: var(--slate-700);
      border: 1px solid var(--slate-600);
      border-radius: 0.5rem;
      padding: 0.5rem 0.75rem;
      color: var(--slate-100);
      font-size: 0.875rem;
      font-family: inherit;
      min-width: 8rem;
    }

    .filter-select:focus {
      outline: none;
      border-color: var(--primary-500);
    }

    .view-controls {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1.5rem;
    }

    .view-toggle {
      display: flex;
      background: var(--slate-800);
      border-radius: 0.75rem;
      padding: 0.25rem;
      border: 1px solid var(--slate-700);
    }

    .view-button {
      background: none;
      border: none;
      padding: 0.5rem 1rem;
      border-radius: 0.5rem;
      color: var(--slate-400);
      cursor: pointer;
      transition: all 0.2s ease-in-out;
      font-family: inherit;
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .view-button.active {
      background: var(--primary-500);
      color: white;
    }

    .view-button:hover:not(.active) {
      background: var(--slate-700);
      color: var(--slate-200);
    }

    .results-info {
      color: var(--slate-400);
      font-size: 0.875rem;
    }

    .loading {
      text-align: center;
      padding: 3rem;
      color: var(--slate-400);
    }

    .loading-spinner {
      width: 2rem;
      height: 2rem;
      border: 2px solid var(--slate-700);
      border-top: 2px solid var(--primary-500);
      border-radius: 50%;
      animation: spin 1s linear infinite;
      margin: 0 auto 1rem;
    }

    .error {
      text-align: center;
      padding: 3rem;
      color: var(--red-400);
      background: rgba(239, 68, 68, 0.1);
      border-radius: 1rem;
      border: 1px solid rgba(239, 68, 68, 0.2);
    }

    .empty-state {
      text-align: center;
      padding: 4rem 2rem;
      color: var(--slate-400);
    }

    .empty-icon {
      font-size: 4rem;
      margin-bottom: 1rem;
      opacity: 0.6;
    }

    .empty-title {
      font-size: 1.5rem;
      font-weight: 600;
      color: var(--slate-300);
      margin-bottom: 0.5rem;
    }

    .empty-description {
      font-size: 1rem;
      line-height: 1.6;
    }

    .file-groups {
      display: flex;
      flex-direction: column;
      gap: 2rem;
    }

    .file-group {
      background: var(--slate-800);
      border-radius: 1rem;
      border: 1px solid var(--slate-700);
      overflow: hidden;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3);
    }

    .group-header {
      background: var(--slate-750);
      padding: 1rem 1.5rem;
      border-bottom: 1px solid var(--slate-700);
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .group-date {
      font-size: 1.125rem;
      font-weight: 600;
      color: var(--slate-100);
    }

    .group-count {
      font-size: 0.875rem;
      color: var(--slate-400);
      background: var(--slate-700);
      padding: 0.25rem 0.75rem;
      border-radius: 1rem;
    }

    .file-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
      gap: 1rem;
      padding: 1.5rem;
    }

    .file-list {
      display: flex;
      flex-direction: column;
    }

    .file-item {
      background: var(--slate-700);
      border-radius: 0.75rem;
      overflow: hidden;
      transition: all 0.2s ease-in-out;
      cursor: pointer;
      border: 1px solid var(--slate-600);
    }

    .file-item:hover {
      transform: translateY(-2px);
      box-shadow: 0 8px 25px 0 rgba(0, 0, 0, 0.4);
      border-color: var(--slate-500);
    }

    .file-thumbnail {
      width: 100%;
      height: 8rem;
      background: var(--slate-600);
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 2rem;
      color: var(--slate-400);
      position: relative;
    }

    .file-thumbnail img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .file-type-badge {
      position: absolute;
      top: 0.5rem;
      right: 0.5rem;
      background: rgba(0, 0, 0, 0.7);
      color: white;
      padding: 0.25rem 0.5rem;
      border-radius: 0.25rem;
      font-size: 0.75rem;
      font-weight: 500;
    }

    .file-info {
      padding: 1rem;
    }

    .file-name {
      font-size: 0.875rem;
      font-weight: 500;
      color: var(--slate-100);
      margin-bottom: 0.5rem;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .file-details {
      font-size: 0.75rem;
      color: var(--slate-400);
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }

    .file-list-item {
      display: flex;
      align-items: center;
      gap: 1rem;
      padding: 1rem 1.5rem;
      border-bottom: 1px solid var(--slate-700);
      transition: all 0.2s ease-in-out;
      cursor: pointer;
    }

    .file-list-item:hover {
      background: var(--slate-750);
    }

    .file-list-item:last-child {
      border-bottom: none;
    }

    .list-thumbnail {
      width: 3rem;
      height: 3rem;
      background: var(--slate-600);
      border-radius: 0.5rem;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
    }

    .list-thumbnail img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      border-radius: 0.5rem;
    }

    .list-file-info {
      flex: 1;
      min-width: 0;
    }

    .list-file-name {
      font-weight: 500;
      color: var(--slate-100);
      margin-bottom: 0.25rem;
    }

    .list-file-details {
      font-size: 0.875rem;
      color: var(--slate-400);
    }

    @keyframes spin {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
    }

    @media (max-width: 768px) {
      .browse-container {
        padding: 1rem;
      }

      .search-input-wrapper {
        flex-direction: column;
      }

      .filters {
        flex-direction: column;
      }

      .view-controls {
        flex-direction: column;
        gap: 1rem;
        align-items: stretch;
      }

      .file-grid {
        grid-template-columns: repeat(auto-fill, minmax(8rem, 1fr));
        padding: 1rem;
      }
    }
  `;

  @property({ type: String })
  searchQuery = '';

  @state()
  private files: MediaFile[] = [];

  @state()
  private filteredFiles: MediaFile[] = [];

  @state()
  private fileGroups: FileGroup[] = [];

  @state()
  private loading = true;

  @state()
  private error: string | null = null;

  @state()
  private viewMode: 'grid' | 'list' = 'grid';

  @state()
  private sortBy: 'date' | 'name' | 'size' = 'date';

  @state()
  private filterType: 'all' | 'image' | 'video' = 'all';

  async connectedCallback() {
    super.connectedCallback();
    await this.loadFiles();
  }

  render() {
    return html`
      <div class="browse-container">
        ${this.renderSearchBar()}
        ${this.renderViewControls()}
        ${this.renderContent()}
      </div>
    `;
  }

  private renderSearchBar() {
    return html`
      <div class="search-bar">
        <div class="search-input-wrapper">
          <input
            type="text"
            class="search-input"
            placeholder="Search by filename, camera, location..."
            .value=${this.searchQuery}
            @input=${this.handleSearchInput}
            @keydown=${this.handleSearchKeydown}
          />
          <button
            class="search-button"
            @click=${this.performSearch}
          >
            🔍 Search
          </button>
        </div>
        
        <div class="filters">
          <div class="filter-group">
            <label class="filter-label">Type</label>
            <select
              class="filter-select"
              .value=${this.filterType}
              @change=${this.handleTypeFilter}
            >
              <option value="all">All Media</option>
              <option value="image">Images</option>
              <option value="video">Videos</option>
            </select>
          </div>
          
          <div class="filter-group">
            <label class="filter-label">Sort By</label>
            <select
              class="filter-select"
              .value=${this.sortBy}
              @change=${this.handleSortChange}
            >
              <option value="date">Date</option>
              <option value="name">Name</option>
              <option value="size">Size</option>
            </select>
          </div>
        </div>
      </div>
    `;
  }

  private renderViewControls() {
    return html`
      <div class="view-controls">
        <div class="results-info">
          ${this.filteredFiles.length} ${this.filteredFiles.length === 1 ? 'file' : 'files'} found
        </div>
        
        <div class="view-toggle">
          <button
            class="view-button ${this.viewMode === 'grid' ? 'active' : ''}"
            @click=${() => this.setViewMode('grid')}
          >
            📱 Grid
          </button>
          <button
            class="view-button ${this.viewMode === 'list' ? 'active' : ''}"
            @click=${() => this.setViewMode('list')}
          >
            📋 List
          </button>
        </div>
      </div>
    `;
  }

  private renderContent() {
    if (this.loading) {
      return html`
        <div class="loading">
          <div class="loading-spinner"></div>
          <p>Loading your media...</p>
        </div>
      `;
    }

    if (this.error) {
      return html`
        <div class="error">
          <h3>Error Loading Files</h3>
          <p>${this.error}</p>
          <button class="search-button" @click=${this.loadFiles}>
            Try Again
          </button>
        </div>
      `;
    }

    if (this.filteredFiles.length === 0) {
      return html`
        <div class="empty-state">
          <div class="empty-icon">${this.files.length === 0 ? '📂' : '🔍'}</div>
          <h3 class="empty-title">
            ${this.files.length === 0 ? 'No media files found' : 'No results found'}
          </h3>
          <p class="empty-description">
            ${this.files.length === 0 
              ? 'Upload some photos or videos to get started with organizing your media collection.'
              : 'Try adjusting your search terms or filters to find what you\'re looking for.'
            }
          </p>
        </div>
      `;
    }

    return html`
      <div class="file-groups">
        ${this.fileGroups.map(group => this.renderFileGroup(group))}
      </div>
    `;
  }

  private renderFileGroup(group: FileGroup) {
    return html`
      <div class="file-group">
        <div class="group-header">
          <div class="group-date">${this.formatGroupDate(group.date)}</div>
          <div class="group-count">${group.files.length} files</div>
        </div>
        
        ${this.viewMode === 'grid' 
          ? this.renderFileGrid(group.files)
          : this.renderFileList(group.files)
        }
      </div>
    `;
  }

  private renderFileGrid(files: MediaFile[]) {
    return html`
      <div class="file-grid">
        ${files.map(file => html`
          <div class="file-item" @click=${() => this.openFile(file)}>
            <div class="file-thumbnail">
              ${file.thumbnail 
                ? html`<img src="${file.thumbnail}" alt="${file.filename}" />`
                : html`${file.type === 'image' ? '🖼️' : '🎬'}`
              }
              <div class="file-type-badge">${file.type.toUpperCase()}</div>
            </div>
            <div class="file-info">
              <div class="file-name" title="${file.filename}">${file.filename}</div>
              <div class="file-details">
                <div>${this.formatFileSize(file.size)}</div>
                ${file.metadata.camera ? html`<div>📷 ${file.metadata.camera}</div>` : ''}
                ${file.metadata.date ? html`<div>📅 ${this.formatDate(file.metadata.date)}</div>` : ''}
              </div>
            </div>
          </div>
        `)}
      </div>
    `;
  }

  private renderFileList(files: MediaFile[]) {
    return html`
      <div class="file-list">
        ${files.map(file => html`
          <div class="file-list-item" @click=${() => this.openFile(file)}>
            <div class="list-thumbnail">
              ${file.thumbnail 
                ? html`<img src="${file.thumbnail}" alt="${file.filename}" />`
                : html`${file.type === 'image' ? '🖼️' : '🎬'}`
              }
            </div>
            <div class="list-file-info">
              <div class="list-file-name">${file.filename}</div>
              <div class="list-file-details">
                ${this.formatFileSize(file.size)} • 
                ${file.metadata.camera || 'Unknown camera'} • 
                ${this.formatDate(file.metadata.date)}
              </div>
            </div>
          </div>
        `)}
      </div>
    `;
  }

  private async loadFiles() {
    try {
      this.loading = true;
      this.error = null;
      
      // Use the real API to get files
      const response = await apiService.listFiles(
        this.searchQuery || undefined,
        this.filterType === 'all' ? undefined : this.filterType,
        1000, // Get more files initially
        0
      );
      
      if (response && response.data && response.data.files) {
        this.files = response.data.files.map((file: any) => this.convertApiFileToMediaFile(file));
      } else {
        this.files = [];
      }
      
      this.filterAndGroupFiles();
      
    } catch (error) {
      this.error = error instanceof Error ? error.message : 'Failed to load files';
      console.error('Failed to load files:', error);
    } finally {
      this.loading = false;
    }
  }

  private convertApiFileToMediaFile(apiFile: any): MediaFile {
    return {
      id: apiFile.id,
      filename: apiFile.filename,
      originalPath: apiFile.relative_path,
      organizedPath: apiFile.relative_path,
      size: apiFile.size,
      type: apiFile.type === 'image' ? 'image' : 'video',
      metadata: {
        date: apiFile.date_taken,
        camera: apiFile.camera,
        location: apiFile.location,
        width: apiFile.width,
        height: apiFile.height,
        duration: apiFile.duration,
      },
      thumbnail: apiFile.type === 'image' ? `http://localhost:8080${apiFile.url}` : undefined,
    };
  }

  private filterAndGroupFiles() {
    // Files are already filtered by the API based on search and type
    // We just need to apply client-side sorting and grouping
    let filtered = [...this.files];
    
    // Apply sorting
    filtered.sort((a, b) => {
      switch (this.sortBy) {
        case 'name':
          return a.filename.localeCompare(b.filename);
        case 'size':
          return b.size - a.size;
        case 'date':
        default:
          const dateA = a.metadata.date ? new Date(a.metadata.date).getTime() : a.size; // fallback to size if no date
          const dateB = b.metadata.date ? new Date(b.metadata.date).getTime() : b.size;
          return dateB - dateA;
      }
    });
    
    this.filteredFiles = filtered;
    
    // Group by date
    const groups = new Map<string, MediaFile[]>();
    
    filtered.forEach(file => {
      let dateKey: string;
      if (file.metadata.date) {
        const date = new Date(file.metadata.date);
        dateKey = date.toISOString().split('T')[0]; // YYYY-MM-DD
      } else {
        dateKey = 'unknown-date';
      }
      
      if (!groups.has(dateKey)) {
        groups.set(dateKey, []);
      }
      groups.get(dateKey)!.push(file);
    });
    
    this.fileGroups = Array.from(groups.entries()).map(([date, files]) => ({
      date,
      files
    })).sort((a, b) => b.date.localeCompare(a.date));
  }

  private handleSearchInput(e: Event) {
    const input = e.target as HTMLInputElement;
    this.searchQuery = input.value;
  }

  private handleSearchKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      this.performSearch();
    }
  }

  private performSearch() {
    this.loadFiles(); // Reload files with new search query
  }

  private handleTypeFilter(e: Event) {
    const select = e.target as HTMLSelectElement;
    this.filterType = select.value as 'all' | 'image' | 'video';
    this.loadFiles(); // Reload files with new filter
  }

  private handleSortChange(e: Event) {
    const select = e.target as HTMLSelectElement;
    this.sortBy = select.value as 'date' | 'name' | 'size';
    this.filterAndGroupFiles();
  }

  private setViewMode(mode: 'grid' | 'list') {
    this.viewMode = mode;
  }

  private openFile(file: MediaFile) {
    // TODO: Implement file preview/modal
    console.log('Opening file:', file);
    // Could dispatch an event to show a modal or navigate to a detail view
  }

  private formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  private formatDate(dateString?: string): string {
    if (!dateString) return 'Unknown date';
    
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  private formatGroupDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  }
}
