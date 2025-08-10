import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { SortifyUpload } from './upload.js';

// Mock the API service
vi.mock('../services/api.js', () => ({
  apiService: {
    uploadFile: vi.fn().mockResolvedValue({
      id: 'upload-123',
      sessionId: 'session-123',
      filename: 'test.jpg',
      mediaInfo: { type: 'image', width: 1920, height: 1080 }
    })
  }
}));

describe('SortifyUpload', () => {
  let element: SortifyUpload;

  beforeEach(async () => {
    element = new SortifyUpload();
    document.body.appendChild(element);
    await element.updateComplete;
  });

  afterEach(() => {
    if (element.parentNode) {
      element.parentNode.removeChild(element);
    }
    vi.clearAllMocks();
  });

  describe('initialization', () => {
    it('should create an instance', () => {
      expect(element).toBeInstanceOf(SortifyUpload);
    });

    it('should have default properties', () => {
      expect(element.disabled).toBe(false);
      expect(element.maxFileSize).toBe(0);
    });

    it('should render the dropzone', () => {
      const dropzone = element.shadowRoot?.querySelector('.dropzone');
      expect(dropzone).toBeTruthy();
    });

    it('should render file input', () => {
      const fileInput = element.shadowRoot?.querySelector('.file-input');
      expect(fileInput).toBeTruthy();
    });

    it('should initialize with empty upload queue', () => {
      const queue = (element as any).uploadQueue;
      expect(queue).toEqual([]);
    });
  });

  describe('properties', () => {
    it('should set disabled property', async () => {
      element.disabled = true;
      await element.updateComplete;
      expect(element.disabled).toBe(true);
    });

    it('should set maxFileSize property', async () => {
      element.maxFileSize = 5000000; // 5MB
      await element.updateComplete;
      expect(element.maxFileSize).toBe(5000000);
    });
  });

  describe('file validation', () => {
    it('should validate image files', () => {
      const imageFile = new File(['test'], 'test.jpg', { type: 'image/jpeg' });
      const isValid = (element as any).isValidFile(imageFile);
      expect(isValid).toBe(true);
    });

    it('should validate video files', () => {
      const videoFile = new File(['test'], 'test.mp4', { type: 'video/mp4' });
      const isValid = (element as any).isValidFile(videoFile);
      expect(isValid).toBe(true);
    });

    it('should reject non-media files', () => {
      const textFile = new File(['test'], 'test.txt', { type: 'text/plain' });
      const isValid = (element as any).isValidFile(textFile);
      expect(isValid).toBe(false);
    });

    it('should validate file size when maxFileSize is set', () => {
      element.maxFileSize = 100;
      const largeFile = new File(['x'.repeat(200)], 'large.jpg', { type: 'image/jpeg' });
      const isValid = (element as any).isValidFile(largeFile);
      expect(isValid).toBe(false);
    });

    it('should allow files under maxFileSize', () => {
      element.maxFileSize = 1000;
      const smallFile = new File(['small'], 'small.jpg', { type: 'image/jpeg' });
      const isValid = (element as any).isValidFile(smallFile);
      expect(isValid).toBe(true);
    });

    it('should allow any file size when maxFileSize is 0', () => {
      element.maxFileSize = 0;
      const largeFile = new File(['x'.repeat(10000)], 'large.jpg', { type: 'image/jpeg' });
      const isValid = (element as any).isValidFile(largeFile);
      expect(isValid).toBe(true);
    });
  });

  describe('file management', () => {
    it('should add files to queue', async () => {
      const file1 = new File(['test1'], 'test1.jpg', { type: 'image/jpeg' });
      const file2 = new File(['test2'], 'test2.jpg', { type: 'image/jpeg' });
      
      (element as any).addFilesToQueue([file1, file2]);
      await element.updateComplete;
      
      const queue = (element as any).uploadQueue;
      expect(queue).toHaveLength(2);
      expect(queue[0].file).toBe(file1);
      expect(queue[1].file).toBe(file2);
    });

    it('should generate unique IDs for files', async () => {
      const file1 = new File(['test1'], 'test.jpg', { type: 'image/jpeg' });
      const file2 = new File(['test2'], 'test.jpg', { type: 'image/jpeg' });
      
      (element as any).addFilesToQueue([file1, file2]);
      await element.updateComplete;
      
      const queue = (element as any).uploadQueue;
      expect(queue[0].id).toBeDefined();
      expect(queue[1].id).toBeDefined();
      expect(queue[0].id).not.toBe(queue[1].id);
    });

    it('should remove file from queue', async () => {
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' });
      (element as any).addFilesToQueue([file]);
      await element.updateComplete;
      
      const queue = (element as any).uploadQueue;
      const fileId = queue[0].id;
      
      (element as any).removeFromQueue(fileId);
      await element.updateComplete;
      
      expect((element as any).uploadQueue).toHaveLength(0);
    });

    it('should clear all files from queue', async () => {
      const file1 = new File(['test1'], 'test1.jpg', { type: 'image/jpeg' });
      const file2 = new File(['test2'], 'test2.jpg', { type: 'image/jpeg' });
      (element as any).addFilesToQueue([file1, file2]);
      await element.updateComplete;
      
      (element as any).clearAll();
      await element.updateComplete;
      
      expect((element as any).uploadQueue).toHaveLength(0);
    });

    it('should clear only completed files', async () => {
      const file1 = new File(['test1'], 'test1.jpg', { type: 'image/jpeg' });
      const file2 = new File(['test2'], 'test2.jpg', { type: 'image/jpeg' });
      
      (element as any).addFilesToQueue([file1, file2]);
      await element.updateComplete;
      
      // Simulate one completed upload
      const queue = (element as any).uploadQueue;
      queue[0].status = 'completed';
      
      (element as any).clearCompleted();
      await element.updateComplete;
      
      const remainingQueue = (element as any).uploadQueue;
      expect(remainingQueue).toHaveLength(1);
      expect(remainingQueue[0].file).toBe(file2);
    });
  });

  describe('drag and drop functionality', () => {
    it('should handle dragover event', () => {
      const dropzone = element.shadowRoot?.querySelector('.dropzone') as HTMLElement;
      const event = new DragEvent('dragover', { bubbles: true, cancelable: true });
      const spy = vi.spyOn(event, 'preventDefault');
      
      dropzone?.dispatchEvent(event);
      
      expect(spy).toHaveBeenCalled();
    });

    it('should handle drop event with files', async () => {
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' });
      
      // Mock the files property on dataTransfer
      const mockDataTransfer = {
        files: [file]
      };
      
      const dropzone = element.shadowRoot?.querySelector('.dropzone') as HTMLElement;
      const event = new DragEvent('drop', { 
        bubbles: true, 
        cancelable: true
      });
      
      // Override the dataTransfer property
      Object.defineProperty(event, 'dataTransfer', {
        value: mockDataTransfer,
        writable: false
      });
      
      dropzone?.dispatchEvent(event);
      await element.updateComplete;
      
      const queue = (element as any).uploadQueue;
      expect(queue).toHaveLength(1);
      expect(queue[0].file).toBe(file);
    });
  });

  describe('upload functionality', () => {
    it('should handle file upload flow', async () => {
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' });
      
      (element as any).addFilesToQueue([file]);
      await element.updateComplete;
      
      const queue = (element as any).uploadQueue;
      expect(queue[0].status).toBe('completed'); // Upload completes immediately with mocked API
      expect(queue[0].file).toBe(file);
    });

    it('should pause upload', async () => {
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' });
      
      (element as any).addFilesToQueue([file]);
      await element.updateComplete;
      
      const queue = (element as any).uploadQueue;
      const fileId = queue[0].id;
      
      // Simulate upload in progress
      queue[0].status = 'uploading';
      queue[0].abortController = new AbortController();
      
      (element as any).pauseUpload(fileId);
      
      expect(queue[0].status).toBe('paused');
    });

    it('should resume paused upload', async () => {
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' });
      
      (element as any).addFilesToQueue([file]);
      await element.updateComplete;
      
      const queue = (element as any).uploadQueue;
      const fileId = queue[0].id;
      
      // Simulate paused upload
      queue[0].status = 'paused';
      
      (element as any).resumeUpload(fileId);
      
      expect(queue[0].status).toBe('pending');
    });

    it('should retry failed upload', async () => {
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' });
      
      (element as any).addFilesToQueue([file]);
      await element.updateComplete;
      
      const queue = (element as any).uploadQueue;
      const fileId = queue[0].id;
      
      // Simulate failed upload
      queue[0].status = 'error';
      queue[0].error = 'Upload failed';
      
      (element as any).retryUpload(fileId);
      
      expect(queue[0].status).toBe('pending');
      expect(queue[0].error).toBeUndefined();
    });
  });

  describe('thumbnail generation', () => {
    it('should generate thumbnail URL for files', () => {
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' });
      
      // Mock URL.createObjectURL
      global.URL.createObjectURL = vi.fn(() => 'mock-thumbnail-url');
      
      const thumbnailUrl = (element as any).getThumbnailUrl(file);
      
      expect(thumbnailUrl).toBe('mock-thumbnail-url');
      expect(global.URL.createObjectURL).toHaveBeenCalledWith(file);
    });

    it('should cache thumbnails', () => {
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' });
      
      global.URL.createObjectURL = vi.fn(() => 'mock-thumbnail-url');
      
      // Generate thumbnail twice
      const url1 = (element as any).getThumbnailUrl(file);
      const url2 = (element as any).getThumbnailUrl(file);
      
      expect(url1).toBe(url2);
      expect(global.URL.createObjectURL).toHaveBeenCalledTimes(1);
    });
  });

  describe('utility methods', () => {
    it('should format file sizes correctly', () => {
      expect((element as any).formatFileSize(0)).toBe('0 B');
      expect((element as any).formatFileSize(1024)).toBe('1 KB');
      expect((element as any).formatFileSize(1048576)).toBe('1 MB');
      expect((element as any).formatFileSize(1073741824)).toBe('1 GB');
    });

    it('should handle file input dialog', () => {
      const fileInput = element.shadowRoot?.querySelector('.file-input') as HTMLInputElement;
      const clickSpy = vi.spyOn(fileInput, 'click');
      
      (element as any).openFileDialog();
      
      expect(clickSpy).toHaveBeenCalled();
    });
  });

  describe('UI interactions', () => {
    it('should open file picker when dropzone is clicked', () => {
      const dropzone = element.shadowRoot?.querySelector('.dropzone') as HTMLElement;
      const fileInput = element.shadowRoot?.querySelector('.file-input') as HTMLInputElement;
      const clickSpy = vi.spyOn(fileInput, 'click');
      
      dropzone?.click();
      
      expect(clickSpy).toHaveBeenCalled();
    });

    it('should handle file input change', async () => {
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' });
      const fileInput = element.shadowRoot?.querySelector('.file-input') as HTMLInputElement;
      
      // Mock the files property
      Object.defineProperty(fileInput, 'files', {
        value: [file],
        writable: false
      });
      
      const event = new Event('change', { bubbles: true });
      fileInput.dispatchEvent(event);
      
      await element.updateComplete;
      
      const queue = (element as any).uploadQueue;
      expect(queue).toHaveLength(1);
      expect(queue[0].file).toBe(file);
    });
  });

  describe('cleanup', () => {
    it('should clean up thumbnails on disconnect', () => {
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' });
      
      global.URL.createObjectURL = vi.fn(() => 'mock-thumbnail-url');
      global.URL.revokeObjectURL = vi.fn();
      
      // Generate a thumbnail
      (element as any).getThumbnailUrl(file);
      
      // Disconnect the element
      element.disconnectedCallback();
      
      expect(global.URL.revokeObjectURL).toHaveBeenCalledWith('mock-thumbnail-url');
    });
  });
});
