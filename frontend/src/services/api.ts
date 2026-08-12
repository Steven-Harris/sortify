export interface UploadOptions {
  chunkSize?: number;
  onProgress?: (progress: number) => void;
  onError?: (error: Error) => void;
  signal?: AbortSignal;
}

export interface MediaMetadata {
  filename?: string;
  fileSize?: number;
  mimeType?: string;
  mediaType?: 'photo' | 'video' | 'other';
  dateTaken?: string;
  dateSource?: string;
  width?: number;
  height?: number;
  duration?: string;
  camera?: {
    make?: string;
    model?: string;
    software?: string;
    lensModel?: string;
    focalLength?: string;
    aperture?: string;
    shutterSpeed?: string;
    iso?: string;
    flash?: string;
  };
  location?: {
    latitude?: number;
    longitude?: number;
    altitude?: number;
  };
  extraMetadata?: Record<string, string>;
}

export interface UploadResult {
  sessionId: string;
  filename: string;
  fileName?: string;
  originalFileName?: string;
  finalFileName?: string;
  mediaInfo: MediaMetadata;
  organized: boolean;
  duplicate?: boolean;
  finalPath?: string;
  relativePath?: string;
  storedFilename?: string;
  conflictRenamed?: boolean;
  conflictRenamedFrom?: string;
  metadataDateSource?: string;
  metadataDateTaken?: string;
}

export interface UploadResponse extends UploadResult {
  id?: string;
  checksum?: string;
  status?: 'uploaded' | 'processing' | 'completed' | 'error';
}

export interface ProcessResponse {
  id: string;
  originalPath: string;
  organizedPath: string;
  metadata: MediaMetadata;
  status: 'processing' | 'completed' | 'error';
  duplicate?: boolean;
  error?: string;
}

export class ApiService {
  private baseUrl: string;

  constructor(baseUrl?: string) {
    this.baseUrl = baseUrl || '';
  }

  async uploadFile(file: File, options: UploadOptions = {}): Promise<UploadResponse> {
    const { chunkSize = 1024 * 1024, onProgress, onError, signal } = options;

    try {
      onProgress?.(0);
      const checksum = await this.calculateChecksum(file);
      const uploadId = await this.initializeUpload(file.name, file.size, checksum, chunkSize);

      const totalChunks = Math.ceil(file.size / chunkSize);
      let uploadedBytes = 0;

      for (let chunkIndex = 0; chunkIndex < totalChunks; chunkIndex++) {
        if (signal?.aborted) {
          throw new Error('Upload cancelled');
        }

        const start = chunkIndex * chunkSize;
        const end = Math.min(start + chunkSize, file.size);
        const chunk = file.slice(start, end);

        await this.uploadChunk(uploadId, chunkIndex, chunk, checksum);

        uploadedBytes += chunk.size;
        onProgress?.(uploadedBytes / file.size);
      }

      const result = await this.finalizeUpload(uploadId, checksum);
      const filename = result.finalFileName || result.fileName || result.filename;
      const normalizedMediaInfo: MediaMetadata = {
        ...(result.mediaInfo || {}),
        filename,
        dateSource: result.mediaInfo?.dateSource || result.metadataDateSource,
        dateTaken: result.mediaInfo?.dateTaken || result.metadataDateTaken,
      };

      return {
        ...result,
        filename,
        mediaInfo: normalizedMediaInfo,
        id: result.sessionId,
        status: 'completed',
      };
    } catch (error) {
      const apiError = error instanceof Error ? error : new Error('Upload failed');
      onError?.(apiError);
      throw apiError;
    }
  }

  async listFiles(query?: string, type?: string, limit?: number, offset?: number): Promise<any> {
    const url = new URL(`${this.baseUrl}/api/media/files`, window.location.origin);

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

    const response = await fetch(this.toRequestUrl(url));
    if (!response.ok) {
      throw new Error(`Failed to list files: ${response.statusText}`);
    }

    return response.json();
  }

  private async calculateChecksum(file: File): Promise<{ hash: string; algorithm: string }> {
    if (window.crypto?.subtle) {
      const buffer = await file.arrayBuffer();
      const hashBuffer = await window.crypto.subtle.digest('SHA-256', buffer);
      const hashArray = Array.from(new Uint8Array(hashBuffer));
      const hashHex = hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
      return { hash: hashHex, algorithm: 'sha256' };
    }

    let hash = 0;
    const chunkSize = 2 * 1024 * 1024;
    let offset = 0;

    while (offset < file.size) {
      const chunk = await file.slice(offset, Math.min(offset + chunkSize, file.size)).arrayBuffer();
      const bytes = new Uint8Array(chunk);
      for (let index = 0; index < bytes.length; index++) {
        hash = (hash + bytes[index]) & 0xffffffff;
      }
      offset += chunkSize;
    }

    return { hash: hash.toString(16), algorithm: 'simple' };
  }

  private async initializeUpload(
    filename: string,
    size: number,
    checksum: { hash: string; algorithm: string },
    chunkSize: number,
  ): Promise<string> {
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
        chunkSize,
      }),
    });

    if (!response.ok) {
      throw new Error(`Failed to initialize upload: ${response.statusText}`);
    }

    const result = await response.json();
    return result.uploadId;
  }

  private async uploadChunk(
    uploadId: string,
    chunkIndex: number,
    chunk: Blob,
    checksum: { hash: string; algorithm: string },
  ): Promise<void> {
    const formData = new FormData();
    formData.append('chunk', chunk);
    formData.append('sessionId', uploadId);
    formData.append('chunkNumber', chunkIndex.toString());
    formData.append('checksum', checksum.hash);
    formData.append('algorithm', checksum.algorithm);

    const response = await fetch(`${this.baseUrl}/api/upload/chunk`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      let errorDetail = response.statusText;
      try {
        const errorData = await response.json();
        if (errorData?.error) {
          errorDetail = errorData.error;
        }
      } catch {
        // Fall back to statusText.
      }

      throw new Error(`Failed to upload chunk ${chunkIndex}: ${errorDetail}`);
    }
  }

  private async finalizeUpload(
    uploadId: string,
    checksum: { hash: string; algorithm: string },
  ): Promise<UploadResult> {
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
      let errorDetail = response.statusText;
      try {
        const errorData = await response.json();
        if (errorData?.error) {
          errorDetail = errorData.error;
        }
      } catch {
        // Fall back to statusText.
      }

      throw new Error(`Failed to finalize upload: ${errorDetail}`);
    }

    return response.json();
  }

  private toRequestUrl(url: URL): string {
    return this.baseUrl ? url.toString() : `${url.pathname}${url.search}`;
  }
}

function createApiService(): ApiService {
  const isDevelopment = import.meta.env.DEV;

  if (isDevelopment) {
    const explicitApiUrl = import.meta.env.VITE_API_URL;
    if (explicitApiUrl) {
      return new ApiService(explicitApiUrl);
    }
    return new ApiService('');
  }

  return new ApiService('');
}

export const apiService = createApiService();
