import { vi } from 'vitest'

// Global fetch mock
globalThis.fetch = vi.fn(() =>
  Promise.resolve({
    ok: true,
    status: 200,
    json: () => Promise.resolve({}),
    text: () => Promise.resolve(''),
    blob: () => Promise.resolve(new Blob()),
    arrayBuffer: () => Promise.resolve(new ArrayBuffer(0)),
    headers: new Headers(),
    url: '',
    statusText: 'OK',
    type: 'basic',
    redirected: false,
    body: null,
    bodyUsed: false,
    clone: vi.fn(),
    // Add missing Response methods as no-op
    formData: () => Promise.resolve(new FormData()),
    bytes: () => Promise.resolve(new Uint8Array()),
  } as unknown as Response)
)

// Mock File with proper methods
class MockFile extends Blob {
  name: string
  lastModified: number
  webkitRelativePath: string

  constructor(fileBits: BlobPart[], fileName: string, options?: FilePropertyBag) {
    super(fileBits, options)
    this.name = fileName
    this.lastModified = options?.lastModified || Date.now()
    this.webkitRelativePath = ''
  }

  arrayBuffer(): Promise<ArrayBuffer> {
    return Promise.resolve(new ArrayBuffer(this.size))
  }

  text(): Promise<string> {
    return Promise.resolve('')
  }

  stream(): ReadableStream {
    return new ReadableStream()
  }

  slice(_start?: number, _end?: number, contentType?: string): Blob {
    return new MockFile([], this.name, { type: contentType })
  }
}

// Mock FileList
class MockFileList extends Array<File> {
  item(index: number): File | null {
    return this[index] || null
  }
}

// Mock FormData
class MockFormData {
  private data = new Map<string, any>()

  append(name: string, value: string | Blob, _fileName?: string): void {
    this.data.set(name, value)
  }

  get(name: string): FormDataEntryValue | null {
    return this.data.get(name) || null
  }

  has(name: string): boolean {
    return this.data.has(name)
  }

  delete(name: string): void {
    this.data.delete(name)
  }

  set(name: string, value: string | Blob, _fileName?: string): void {
    this.data.set(name, value)
  }

  entries(): IterableIterator<[string, FormDataEntryValue]> {
    return this.data.entries()
  }

  keys(): IterableIterator<string> {
    return this.data.keys()
  }

  values(): IterableIterator<FormDataEntryValue> {
    return this.data.values()
  }

  forEach(callback: (value: FormDataEntryValue, key: string, parent: FormData) => void): void {
    this.data.forEach((value, key) => callback(value, key, this as any))
  }

  [Symbol.iterator](): IterableIterator<[string, FormDataEntryValue]> {
    return this.data.entries()
  }
}

// Mock crypto.subtle
const mockSubtle = {
  digest: vi.fn().mockResolvedValue(new ArrayBuffer(32)),
}

// Global assignments
globalThis.File = MockFile as any
globalThis.FileList = MockFileList as any
globalThis.FormData = MockFormData as any

// Mock crypto - handle both global and Object.defineProperty
if (typeof globalThis.crypto === 'undefined') {
  globalThis.crypto = { subtle: mockSubtle } as any
} else {
  Object.defineProperty(globalThis.crypto, 'subtle', {
    value: mockSubtle,
    writable: true,
    configurable: true
  })
}

// Mock URL - jsdom should provide this, but ensure it's available
if (typeof globalThis.URL === 'undefined') {
  globalThis.URL = class MockURL {
    href: string
    origin: string
    protocol: string
    host: string
    hostname: string
    port: string
    pathname: string
    search: string
    hash: string
    searchParams: URLSearchParams

    constructor(url: string, _base?: string | URL) {
      this.href = url
      this.origin = 'http://localhost'
      this.protocol = 'http:'
      this.host = 'localhost'
      this.hostname = 'localhost'
      this.port = ''
      this.pathname = '/'
      this.search = ''
      this.hash = ''
      this.searchParams = new URLSearchParams()
    }

    toString(): string {
      return this.href
    }

    toJSON(): string {
      return this.href
    }
  } as any
}

// Mock URLSearchParams if not available
if (typeof globalThis.URLSearchParams === 'undefined') {
  globalThis.URLSearchParams = class MockURLSearchParams {
    private params = new Map<string, string>()

    constructor(_init?: string | URLSearchParams | Record<string, string>) {
      // Basic implementation
    }

    append(name: string, value: string): void {
      this.params.set(name, value)
    }

    get(name: string): string | null {
      return this.params.get(name) || null
    }

    set(name: string, value: string): void {
      this.params.set(name, value)
    }

    toString(): string {
      const pairs: string[] = []
      this.params.forEach((value, key) => {
        pairs.push(`${encodeURIComponent(key)}=${encodeURIComponent(value)}`)
      })
      return pairs.join('&')
    }
  } as any
}

// Mock AbortController if not available
if (typeof globalThis.AbortController === 'undefined') {
  globalThis.AbortController = class MockAbortController {
    signal: AbortSignal

    constructor() {
      this.signal = {
        aborted: false,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
        onabort: null,
        reason: undefined,
        throwIfAborted: vi.fn(),
      } as any
    }

    abort(_reason?: any): void {
      (this.signal as any).aborted = true
      if (this.signal.onabort) {
        this.signal.onabort(new Event('abort'))
      }
    }
  } as any
}

// Mock XMLHttpRequest for upload progress
globalThis.XMLHttpRequest = class MockXMLHttpRequest {
  upload = {
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  }
  readyState = 4
  status = 200
  statusText = 'OK'
  responseText = ''
  response = ''
  onreadystatechange: (() => void) | null = null

  open = vi.fn()
  send = vi.fn()
  setRequestHeader = vi.fn()
  addEventListener = vi.fn()
  removeEventListener = vi.fn()
  abort = vi.fn()
} as any

// Mock canvas for thumbnail generation
HTMLCanvasElement.prototype.getContext = vi.fn().mockReturnValue({
  drawImage: vi.fn(),
  getImageData: vi.fn(),
  putImageData: vi.fn(),
  canvas: {
    toBlob: vi.fn((callback) => callback(new Blob())),
    toDataURL: vi.fn().mockReturnValue('data:image/png;base64,test'),
  },
})

// Mock image loading
globalThis.Image = class MockImage {
  onload: (() => void) | null = null
  onerror: (() => void) | null = null
  src = ''
  width = 100
  height = 100

  constructor() {
    setTimeout(() => {
      if (this.onload) this.onload()
    }, 0)
  }
} as any

// Mock DragEvent if not available
if (typeof globalThis.DragEvent === 'undefined') {
  globalThis.DragEvent = class MockDragEvent extends Event {
    dataTransfer: DataTransfer | null

    constructor(type: string, eventInitDict?: DragEventInit) {
      super(type, eventInitDict)
      this.dataTransfer = {
        files: [] as any,
        getData: vi.fn(),
        setData: vi.fn(),
        clearData: vi.fn(),
        setDragImage: vi.fn(),
        effectAllowed: 'uninitialized' as any,
        dropEffect: 'none' as any,
        items: [] as any,
        types: [],
      } as any
    }
  } as any
}

// Mock createObjectURL and revokeObjectURL
if (typeof globalThis.URL.createObjectURL === 'undefined') {
  globalThis.URL.createObjectURL = vi.fn().mockReturnValue('blob:http://localhost/test')
  globalThis.URL.revokeObjectURL = vi.fn()
}
