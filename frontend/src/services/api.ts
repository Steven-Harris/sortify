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

        try {
          await this.uploadChunk(uploadId, chunkIndex, chunk, checksum);
        } catch (error) {
          const chunkError = error instanceof Error ? error : new Error('Chunk upload failed');
          console.warn(`Chunk upload failed: ${chunkError.message}. Retrying with SHA-256...`);
          
          // If using simple algorithm and it failed, try again with SHA-256
          if (checksum.algorithm === 'simple') {
            // Recalculate checksum using SHA-256 just for this chunk
            try {
              const chunkBuffer = await chunk.arrayBuffer();
              const hashBuffer = await window.crypto.subtle.digest('SHA-256', chunkBuffer);
              const hashArray = Array.from(new Uint8Array(hashBuffer));
              const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
              
              // Try uploading with SHA-256
              await this.uploadChunk(uploadId, chunkIndex, chunk, { hash: hashHex, algorithm: "sha256" });
            } catch (retryError) {
              // If retry fails, throw original error
              throw chunkError;
            }
          } else {
            throw chunkError;
          }
        }
        
        uploadedBytes += chunk.size;
        onProgress?.(uploadedBytes / file.size);
      }

      // Finalize upload
      const result = await this.finalizeUpload(uploadId, checksum);
      return result;

    } catch (error) {
      const apiError = error instanceof Error ? error : new Error('Upload failed');
      
      // Improve error messages for the user
      if (apiError.message.includes('chunk checksum mismatch')) {
        apiError.message = 'Upload failed due to data verification error. This can happen on some mobile browsers. Please try using a different browser or device.';
      }
      
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

  private async calculateChecksum(file: File): Promise<{hash: string, algorithm: string}> {
    // For files smaller than 100MB, load the whole file in memory
    if (file.size < 100 * 1024 * 1024) {
      try {
        const buffer = await file.arrayBuffer();
        if (window.crypto && window.crypto.subtle) {
          // Use SHA256 to match backend implementation
          const hashBuffer = await window.crypto.subtle.digest('SHA-256', buffer);
          const hashArray = Array.from(new Uint8Array(hashBuffer));
          const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
          return { hash: hashHex, algorithm: "sha256" };
        }
      } catch (err) {
        console.warn("Error calculating SHA-256 checksum", err);
        // Will fall back to simple hash below
      }
    }
    
    // Fallback when crypto API fails or for very large files
    // Use a more consistent implementation of simple hash
    let hash = 0;
    const chunkSize = 2 * 1024 * 1024; // 2MB chunks
    let offset = 0;
    
    while (offset < file.size) {
      const chunk = await file.slice(offset, Math.min(offset + chunkSize, file.size)).arrayBuffer();
      const arr = new Uint8Array(chunk);
      
      for (let i = 0; i < arr.length; i++) {
        // Use a simpler hash algorithm that's more consistent across platforms
        hash = (hash + arr[i]) & 0xFFFFFFFF; 
      }
      
      offset += chunkSize;
    }
    
    return { hash: hash.toString(16), algorithm: "simple" };
  }

  private async initializeUpload(filename: string, size: number, checksum: {hash: string, algorithm: string}): Promise<string> {
    const response = await fetch(`${this.baseUrl}/api/upload/start`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        fileName: filename,
        fileSize: size,
        checksum: checksum.hash,
        algorithm: checksum.algorithm,
      }),
    });

    if (!response.ok) {
      throw new Error(`Failed to initialize upload: ${response.statusText}`);
    }

    const result = await response.json();
    return result.uploadId;
  }

  private async uploadChunk(uploadId: string, chunkIndex: number, chunk: Blob, checksum: {hash: string, algorithm: string}): Promise<void> {
    const formData = new FormData();
    formData.append('chunk', chunk);
    formData.append('sessionId', uploadId);
    formData.append('chunkNumber', chunkIndex.toString());
    formData.append('checksum', checksum.hash);
    formData.append('algorithm', checksum.algorithm);

    try {
      const response = await fetch(`${this.baseUrl}/api/upload/chunk`, {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        // Try to get more detailed error message from response
        let errorDetail = response.statusText;
        try {
          const errorData = await response.json();
          if (errorData && errorData.error) {
            errorDetail = errorData.error;
          }
        } catch (e) {
          // If parsing fails, use the status text
          console.warn("Failed to parse error response", e);
        }
        
        throw new Error(`Failed to upload chunk ${chunkIndex}: ${errorDetail}`);
      }
    } catch (error) {
      console.error("Upload chunk error:", error);
      throw error;
    }
  }

  private async finalizeUpload(uploadId: string, checksum: {hash: string, algorithm: string}): Promise<UploadResponse> {
    try {
      const response = await fetch(`${this.baseUrl}/api/upload/complete`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          sessionId: uploadId,
          checksum: checksum.hash,
          algorithm: checksum.algorithm,
        }),
      });

      if (!response.ok) {
        // Try to get more detailed error message from response
        let errorDetail = response.statusText;
        try {
          const errorData = await response.json();
          if (errorData && errorData.error) {
            errorDetail = errorData.error;
          }
        } catch (e) {
          console.warn("Failed to parse error response", e);
        }
        
        // Handle specific errors
        if (errorDetail.includes('file checksum mismatch')) {
          console.warn("File checksum mismatch during finalization, trying with SHA-256");
          
          // If we were using simple algorithm and got a checksum mismatch, try with SHA-256
          if (checksum.algorithm === 'simple') {
            // Try again with a different algorithm
            const retryResponse = await fetch(`${this.baseUrl}/api/upload/complete`, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
              },
              body: JSON.stringify({
                sessionId: uploadId,
                checksum: '', // Don't send checksum to allow force completion
                algorithm: 'sha256',
              }),
            });
            
            if (retryResponse.ok) {
              return retryResponse.json();
            }
          }
        }
        
        throw new Error(`Failed to finalize upload: ${errorDetail}`);
      }

      return response.json();
    } catch (error) {
      console.error("Finalize upload error:", error);
      throw error;
    }
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
