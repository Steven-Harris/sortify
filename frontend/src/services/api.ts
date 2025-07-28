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
  sessionId?: string;
  id?: string;
  filename: string;
  mediaInfo?: any;
  organized?: boolean;
  size?: number;
  checksum?: string;
  status?: 'uploaded' | 'processing' | 'completed' | 'error';
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
      
      // Check if file already exists (disabled for now)
      // const existingFile = await this.checkFileExists(checksum);
      // if (existingFile) {
      //   return existingFile;
      // }

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
      const result = await this.finalizeUpload(uploadId, checksum);
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
   * @deprecated No longer needed - files are processed during upload completion
   * Process uploaded file (extract metadata and organize)
   */
  async processFile(_uploadId: string): Promise<ProcessResponse> {
    // Files are now processed automatically during upload completion
    throw new Error('processFile is deprecated - files are processed during upload completion');
  }

  /**
   * @deprecated No longer needed - files are processed during upload completion
   * Get processing status
   */
  async getProcessStatus(_processId: string): Promise<ProcessResponse> {
    // Files are now processed automatically during upload completion
    throw new Error('getProcessStatus is deprecated - files are processed during upload completion');
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
    // Use SHA256 to match backend implementation
    const buffer = await file.arrayBuffer();
    const hashBuffer = await crypto.subtle.digest('SHA-256', buffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    return hashHex;
  }

  private async initializeUpload(filename: string, size: number, checksum: string): Promise<string> {
    const response = await fetch(`${this.baseUrl}/api/upload/start`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        fileName: filename,
        fileSize: size,
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
    formData.append('sessionId', uploadId);
    formData.append('chunkNumber', chunkIndex.toString());

    const response = await fetch(`${this.baseUrl}/api/upload/chunk`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`Failed to upload chunk ${chunkIndex}: ${response.statusText}`);
    }
  }

  private async finalizeUpload(uploadId: string, checksum: string): Promise<UploadResponse> {
    const response = await fetch(`${this.baseUrl}/api/upload/complete`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        sessionId: uploadId,
        checksum: checksum,
      }),
    });

    if (!response.ok) {
      throw new Error(`Failed to finalize upload: ${response.statusText}`);
    }

    return response.json();
  }
}

// Export a default instance
export const apiService = new ApiService();
