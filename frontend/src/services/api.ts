/**
 * API service for communicating with the Sortify backend
 */

export interface UploadOptions {
  chunkSize?: number;
  onProgress?: (progress: number) => void;
  onError?: (error: Error) => void;
  signal?: AbortSignal;
}

export interface UploadResponse {
  id: string;
  filename: string;
  size: number;
  checksum: string;
  status: 'uploaded' | 'processing' | 'completed' | 'error';
}

export interface ProcessResponse {
  id: string;
  originalPath: string;
  organizedPath: string;
  metadata: {
    date?: string;
    camera?: string;
    location?: string;
    width?: number;
    height?: number;
    duration?: number;
  };
  status: 'processing' | 'completed' | 'error';
  error?: string;
}

export class ApiService {
  private baseUrl: string;

  constructor(baseUrl = 'http://localhost:8080') {
    this.baseUrl = baseUrl;
  }

  /**
   * Upload a file with chunked upload support
   */
  async uploadFile(file: File, options: UploadOptions = {}): Promise<UploadResponse> {
    const { chunkSize = 1024 * 1024, onProgress, onError, signal } = options;
    
    try {
      // Calculate file checksum (simple hash for demo)
      const checksum = await this.calculateChecksum(file);
      
      // Check if file already exists
      const existingFile = await this.checkFileExists(checksum);
      if (existingFile) {
        return existingFile;
      }

      // Start chunked upload
      const uploadId = await this.initializeUpload(file.name, file.size, checksum);
      
      const totalChunks = Math.ceil(file.size / chunkSize);
      let uploadedBytes = 0;

      for (let chunkIndex = 0; chunkIndex < totalChunks; chunkIndex++) {
        if (signal?.aborted) {
          throw new Error('Upload cancelled');
        }

        const start = chunkIndex * chunkSize;
        const end = Math.min(start + chunkSize, file.size);
        const chunk = file.slice(start, end);

        await this.uploadChunk(uploadId, chunkIndex, chunk);
        
        uploadedBytes += chunk.size;
        onProgress?.(uploadedBytes / file.size);
      }

      // Finalize upload
      const result = await this.finalizeUpload(uploadId);
      return result;

    } catch (error) {
      const apiError = error instanceof Error ? error : new Error('Upload failed');
      onError?.(apiError);
      throw apiError;
    }
  }

  /**
   * Check upload status
   */
  async getUploadStatus(uploadId: string): Promise<UploadResponse> {
    const response = await fetch(`${this.baseUrl}/api/upload/status/${uploadId}`);
    if (!response.ok) {
      throw new Error(`Failed to get upload status: ${response.statusText}`);
    }
    return response.json();
  }

  /**
   * Process uploaded file (extract metadata and organize)
   */
  async processFile(uploadId: string): Promise<ProcessResponse> {
    const response = await fetch(`${this.baseUrl}/api/process`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ uploadId }),
    });

    if (!response.ok) {
      throw new Error(`Failed to process file: ${response.statusText}`);
    }

    return response.json();
  }

  /**
   * Get processing status
   */
  async getProcessStatus(processId: string): Promise<ProcessResponse> {
    const response = await fetch(`${this.baseUrl}/api/process/status/${processId}`);
    if (!response.ok) {
      throw new Error(`Failed to get process status: ${response.statusText}`);
    }
    return response.json();
  }

  /**
   * List organized files
   */
  async listFiles(query?: string, type?: string, limit?: number, offset?: number): Promise<any> {
    const url = new URL(`${this.baseUrl}/api/media/files`);
    
    if (query) {
      url.searchParams.set('q', query);
    }
    if (type && type !== 'all') {
      url.searchParams.set('type', type);
    }
    if (limit) {
      url.searchParams.set('limit', limit.toString());
    }
    if (offset) {
      url.searchParams.set('offset', offset.toString());
    }

    const response = await fetch(url.toString());
    if (!response.ok) {
      throw new Error(`Failed to list files: ${response.statusText}`);
    }

    return response.json();
  }

  // Private helper methods

  private async calculateChecksum(file: File): Promise<string> {
    // Simple checksum calculation (in production, use a more robust hash)
    const buffer = await file.arrayBuffer();
    let hash = 0;
    const view = new Uint8Array(buffer);
    
    for (let i = 0; i < view.length; i++) {
      hash = ((hash << 5) - hash + view[i]) & 0xffffffff;
    }
    
    return Math.abs(hash).toString(16);
  }

  private async checkFileExists(checksum: string): Promise<UploadResponse | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/upload/check/${checksum}`);
      if (response.ok) {
        return response.json();
      }
      return null;
    } catch {
      return null;
    }
  }

  private async initializeUpload(filename: string, size: number, checksum: string): Promise<string> {
    const response = await fetch(`${this.baseUrl}/api/upload/init`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        filename,
        size,
        checksum,
      }),
    });

    if (!response.ok) {
      throw new Error(`Failed to initialize upload: ${response.statusText}`);
    }

    const result = await response.json();
    return result.uploadId;
  }

  private async uploadChunk(uploadId: string, chunkIndex: number, chunk: Blob): Promise<void> {
    const formData = new FormData();
    formData.append('chunk', chunk);
    formData.append('chunkIndex', chunkIndex.toString());

    const response = await fetch(`${this.baseUrl}/api/upload/chunk/${uploadId}`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`Failed to upload chunk ${chunkIndex}: ${response.statusText}`);
    }
  }

  private async finalizeUpload(uploadId: string): Promise<UploadResponse> {
    const response = await fetch(`${this.baseUrl}/api/upload/finalize/${uploadId}`, {
      method: 'POST',
    });

    if (!response.ok) {
      throw new Error(`Failed to finalize upload: ${response.statusText}`);
    }

    return response.json();
  }
}

// Export a default instance
export const apiService = new ApiService();
