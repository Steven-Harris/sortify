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

  constructor(baseUrl?: string) {
    // Use relative URLs by default (works with any host/port)
    // Only use provided baseUrl for development/testing
    this.baseUrl = baseUrl || '';
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
    const buffer = await file.arrayBuffer();
    if (window.crypto && window.crypto.subtle) {
      // Use SHA256 to match backend implementation
      const hashBuffer = await window.crypto.subtle.digest('SHA-256', buffer);
      const hashArray = Array.from(new Uint8Array(hashBuffer));
      const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
      return hashHex;
    }
    // Fallback: simple hash (not cryptographically secure)
    let hash = 0;
    const arr = new Uint8Array(buffer);
    for (let i = 0; i < arr.length; i++) {
      hash = ((hash << 5) - hash) + arr[i];
      hash |= 0; // Convert to 32bit integer
    }
    return hash.toString(16);
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

// Create default instance with environment-aware base URL
function createApiService(): ApiService {
  // In development, use empty string to rely on Vite proxy
  // In production, use relative URLs (same origin)
  // Only use full URL if explicitly set via environment variable
  const isDevelopment = import.meta.env.DEV;
  
  if (isDevelopment) {
    // Check if we have an explicit API URL (for non-proxy development)
    const explicitApiUrl = import.meta.env.VITE_API_URL;
    if (explicitApiUrl) {
      console.log('Using explicit API URL for development:', explicitApiUrl);
      return new ApiService(explicitApiUrl);
    }
    // Use empty string to rely on Vite proxy
    console.log('Using Vite proxy for API requests');
    return new ApiService('');
  }
  
  // Production: always use relative URLs
  return new ApiService('');
}

// Export a default instance
export const apiService = createApiService();
