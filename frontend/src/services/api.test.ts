import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiService } from '../services/api';

describe('ApiService', () => {
  let apiService: ApiService;
  let mockFetch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockFetch = vi.fn();
    global.fetch = mockFetch as typeof fetch;

    Object.defineProperty(global, 'crypto', {
      value: {
        subtle: {
          digest: vi.fn().mockResolvedValue(new Uint8Array([1, 2, 3, 4]).buffer),
        },
      },
      configurable: true,
    });

    apiService = new ApiService('http://localhost:8080');
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('uploads a file and returns the ingest result contract', async () => {
    const mockFile = new File(['test content'], 'IMG_20240315_143022.jpg', { type: 'image/jpeg' });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ uploadId: 'upload123' }),
    });
    mockFetch.mockResolvedValueOnce({ ok: true });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        sessionId: 'upload123',
        fileName: 'IMG_20240315_143022.jpg',
        finalFileName: 'IMG_20240315_143022.jpg',
        metadataDateSource: 'filename',
        metadataDateTaken: '2024-03-15T14:30:22Z',
        mediaInfo: {
          mediaType: 'photo',
        },
        organized: true,
        duplicate: false,
        relativePath: '2024/March/IMG_20240315_143022.jpg',
      }),
    });

    const progress = vi.fn();
    const result = await apiService.uploadFile(mockFile, { onProgress: progress });

    expect(result).toEqual({
      id: 'upload123',
      sessionId: 'upload123',
      filename: 'IMG_20240315_143022.jpg',
      fileName: 'IMG_20240315_143022.jpg',
      finalFileName: 'IMG_20240315_143022.jpg',
      mediaInfo: {
        filename: 'IMG_20240315_143022.jpg',
        mediaType: 'photo',
        dateSource: 'filename',
        dateTaken: '2024-03-15T14:30:22Z',
      },
      organized: true,
      duplicate: false,
      relativePath: '2024/March/IMG_20240315_143022.jpg',
      metadataDateSource: 'filename',
      metadataDateTaken: '2024-03-15T14:30:22Z',
      status: 'completed',
    });
    expect(progress).toHaveBeenCalledTimes(2);
    expect(progress).toHaveBeenNthCalledWith(1, 0);
    expect(progress).toHaveBeenNthCalledWith(2, 0);
  });

  it('passes chunk size during initialization', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ uploadId: 'upload123' }),
    });

    await (apiService as any).initializeUpload(
      'test.jpg',
      100,
      { hash: 'abc123', algorithm: 'sha256' },
      2048,
    );

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/upload/start',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          fileName: 'test.jpg',
          fileSize: 100,
          checksum: 'abc123',
          algorithm: 'sha256',
          chunkSize: 2048,
        }),
      }),
    );
  });

  it('lists files with relative URLs when no base url is configured', async () => {
    const localApi = new ApiService('');
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ files: [], total: 0 }),
    });

    await localApi.listFiles('vacation', 'photo', 10, 20);

    expect(mockFetch).toHaveBeenCalledWith('/api/media/files?q=vacation&type=photo&limit=10&offset=20');
  });

  it('surfaces finalize errors', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      statusText: 'Internal Server Error',
      json: async () => ({ error: 'server exploded' }),
    });

    await expect(
      (apiService as any).finalizeUpload('upload123', { hash: 'abc123', algorithm: 'sha256' }),
    ).rejects.toThrow('Failed to finalize upload: server exploded');
  });
});
