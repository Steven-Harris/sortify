import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { SortifyUpload, type UploadFile } from './upload.js';

vi.mock('../services/api.js', () => ({
  apiService: {
    uploadFile: vi.fn().mockResolvedValue({
      sessionId: 'session-123',
      filename: 'IMG_20240315_143022.jpg',
      mediaInfo: {
        mediaType: 'photo',
        dateTaken: '2024-03-15T14:30:22Z',
        dateSource: 'filename',
      },
      organized: true,
      duplicate: false,
      relativePath: '2024/March/IMG_20240315_143022.jpg',
    }),
  },
}));

describe('SortifyUpload', () => {
  let element: SortifyUpload;

  beforeEach(async () => {
    element = new SortifyUpload();
    document.body.appendChild(element);
    await element.updateComplete;
  });

  afterEach(() => {
    element.remove();
    vi.clearAllMocks();
  });

  it('renders the batch upload dropzone', () => {
    expect(element.shadowRoot?.querySelector('.dropzone-title')?.textContent).toContain('Bulk upload');
  });

  it('adds valid files to the queue', async () => {
    const imageFile = new File(['a'], 'a.jpg', { type: 'image/jpeg' });
    const videoFile = new File(['b'], 'b.mp4', { type: 'video/mp4' });

    (element as any).addFilesToQueue([imageFile, videoFile]);
    await element.updateComplete;

    expect((element as any).uploadQueue).toHaveLength(2);
  });

  it('rejects non-media files', () => {
    const textFile = new File(['test'], 'notes.txt', { type: 'text/plain' });
    expect((element as any).isValidFile(textFile)).toBe(false);
  });

  it('builds a completion summary including duplicates', () => {
    (element as any).uploadQueue = [
      { id: '1', file: new File(['a'], 'a.jpg', { type: 'image/jpeg' }), progress: 100, status: 'completed', processResponse: { duplicate: false } },
      { id: '2', file: new File(['b'], 'b.jpg', { type: 'image/jpeg' }), progress: 100, status: 'completed', processResponse: { duplicate: true } },
      { id: '3', file: new File(['c'], 'c.jpg', { type: 'image/jpeg' }), progress: 0, status: 'error' },
    ];

    const summary = (element as any).getSummary();
    expect(summary).toEqual({
      total: 3,
      completed: 2,
      duplicates: 1,
      failed: 1,
      processing: 0,
      pending: 0,
    });
  });

  it('maps upload results into process responses', async () => {
    const file = new File(['test'], 'IMG_20240315_143022.jpg', { type: 'image/jpeg' });
    const item: any = {
      id: 'item-1',
      file,
      progress: 0,
      status: 'pending',
    };

    await (element as any).uploadFile(item);

    expect(item.status).toBe('completed');
    expect(item.processResponse).toEqual({
      id: 'session-123',
      originalPath: 'IMG_20240315_143022.jpg',
      organizedPath: '2024/March/IMG_20240315_143022.jpg',
      metadata: {
        mediaType: 'photo',
        dateTaken: '2024-03-15T14:30:22Z',
        dateSource: 'filename',
      },
      status: 'completed',
      duplicate: false,
    });
  });

  it('pauses an in-flight upload', () => {
    const controller = new AbortController();
    const queue = (element as any).uploadQueue;
    queue.push({
      id: 'upload-1',
      file: new File(['test'], 'test.jpg', { type: 'image/jpeg' }),
      progress: 50,
      status: 'uploading',
      abortController: controller,
    });

    (element as any).pauseUpload('upload-1');

    expect(queue[0].status).toBe('paused');
    expect(controller.signal.aborted).toBe(true);
  });
});
