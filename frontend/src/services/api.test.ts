import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ApiService } from '../services/api'

describe('ApiService', () => {
  let apiService: ApiService
  let mockFetch: any

  beforeEach(() => {
    mockFetch = vi.fn()
    global.fetch = mockFetch
    apiService = new ApiService('http://localhost:8080')
    
    // Mock crypto.subtle using Object.defineProperty to avoid readonly error
    Object.defineProperty(global, 'crypto', {
      value: {
        subtle: {
          digest: vi.fn().mockResolvedValue(new ArrayBuffer(32))
        },
        randomUUID: vi.fn().mockReturnValue('test-uuid')
      },
      writable: true,
      configurable: true
    })
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('constructor', () => {
    it('should initialize with default base URL', () => {
      const service = new ApiService()
      expect(service).toBeInstanceOf(ApiService)
    })

    it('should initialize with custom base URL', () => {
      const service = new ApiService('http://custom.url')
      expect(service).toBeInstanceOf(ApiService)
    })
  })

  describe('uploadFile', () => {
    it('should upload a file successfully', async () => {
      const mockFile = new File(['test content'], 'test.txt', { type: 'text/plain' })
      
      // Mock the upload initialization
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ uploadId: 'upload123' })
      })

      // Mock chunk upload
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ status: 'uploading' })
      })

      // Mock finalize upload
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: '123',
          filename: 'test.txt',
          size: 12,
          status: 'completed'
        })
      })

      const progressCallback = vi.fn()
      const result = await apiService.uploadFile(mockFile, { onProgress: progressCallback })

      expect(result).toEqual({
        id: '123',
        filename: 'test.txt',
        size: 12,
        status: 'completed'
      })
      expect(progressCallback).toHaveBeenCalled()
    })

    it('should handle upload errors', async () => {
      const mockFile = new File(['test'], 'test.txt', { type: 'text/plain' })

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error'
      })

      await expect(apiService.uploadFile(mockFile)).rejects.toThrow('Failed to initialize upload: Internal Server Error')
    })

    it('should handle abort signal', async () => {
      const mockFile = new File(['test content'], 'test.txt', { type: 'text/plain' })
      const abortController = new AbortController()

      // Mock successful initialization
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ uploadId: 'upload123' })
      })

      // Create the upload promise
      const uploadPromise = apiService.uploadFile(mockFile, { signal: abortController.signal })

      // Abort immediately before any chunks are processed
      abortController.abort()

      await expect(uploadPromise).rejects.toThrow('Upload cancelled')
    })

    it('should call progress callback during upload', async () => {
      const mockFile = new File(['test content'], 'test.txt', { type: 'text/plain' })
      const progressCallback = vi.fn()

      // Mock successful responses
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ uploadId: 'upload123' })
      })
      mockFetch.mockResolvedValueOnce({ ok: true })
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ id: '123', filename: 'test.txt', status: 'completed' })
      })

      await apiService.uploadFile(mockFile, { onProgress: progressCallback })

      expect(progressCallback).toHaveBeenCalledWith(0) // Progress called with initial value
    })
  })

  describe('getUploadStatus', () => {
    it('should get upload status successfully', async () => {
      const mockResponse = {
        id: 'upload123',
        filename: 'test.txt',
        status: 'completed',
        size: 100
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const result = await apiService.getUploadStatus('upload123')

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/upload/status/upload123'
      )
      expect(result).toEqual(mockResponse)
    })

    it('should handle status check errors', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        statusText: 'Not Found'
      })

      await expect(apiService.getUploadStatus('nonexistent')).rejects.toThrow('Failed to get upload status: Not Found')
    })
  })

  describe('listFiles', () => {
    it('should list files successfully', async () => {
      const mockResponse = {
        files: [
          { id: '1', filename: 'test1.jpg', type: 'photo' },
          { id: '2', filename: 'test2.jpg', type: 'photo' }
        ],
        total: 2
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const result = await apiService.listFiles()

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/media/files'
      )
      expect(result).toEqual(mockResponse)
    })

    it('should list files with query parameters', async () => {
      const mockResponse = { files: [], total: 0 }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      await apiService.listFiles('vacation', 'photo', 10, 20)

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/media/files?q=vacation&type=photo&limit=10&offset=20'
      )
    })

    it('should handle list files errors', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        statusText: 'Internal Server Error'
      })

      await expect(apiService.listFiles()).rejects.toThrow('Failed to list files: Internal Server Error')
    })
  })

  describe('private methods', () => {
    it('should calculate file checksum', async () => {
      const mockFile = new File(['test'], 'test.txt', { type: 'text/plain' })
      
      // Mock the crypto.subtle.digest to return a specific hash
      const mockHash = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8])
      global.crypto.subtle.digest = vi.fn().mockResolvedValue(mockHash.buffer)

      const result = await (apiService as any).calculateChecksum(mockFile)

      expect(result).toBe('0102030405060708')
      expect(global.crypto.subtle.digest).toHaveBeenCalledWith('SHA-256', expect.any(ArrayBuffer))
    })

    it('should initialize upload', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ uploadId: 'upload123' })
      })

      const result = await (apiService as any).initializeUpload('test.txt', 100, 'checksum123')

      expect(result).toBe('upload123')
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/upload/start',
        expect.objectContaining({
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            fileName: 'test.txt',
            fileSize: 100,
            checksum: 'checksum123'
          })
        })
      )
    })

    it('should upload chunk', async () => {
      const chunk = new Blob(['test'])
      
      mockFetch.mockResolvedValueOnce({ ok: true })

      await (apiService as any).uploadChunk('upload123', 0, chunk)

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/upload/chunk',
        expect.objectContaining({
          method: 'POST',
          body: expect.any(FormData)
        })
      )
    })

    it('should finalize upload', async () => {
      const mockResponse = {
        id: '123',
        filename: 'test.txt',
        status: 'completed'
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const result = await (apiService as any).finalizeUpload('upload123', 'checksum123')

      expect(result).toEqual(mockResponse)
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/upload/complete',
        expect.objectContaining({
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            sessionId: 'upload123',
            checksum: 'checksum123'
          })
        })
      )
    })
  })
})
