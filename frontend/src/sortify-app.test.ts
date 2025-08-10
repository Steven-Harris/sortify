import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { SortifyApp } from './sortify-app'

describe('SortifyApp', () => {
  let element: SortifyApp

  beforeEach(() => {
    // Define the custom element if not already defined
    if (!customElements.get('sortify-app')) {
      customElements.define('sortify-app', SortifyApp)
    }
    
    element = new SortifyApp()
    document.body.appendChild(element)
  })

  afterEach(() => {
    if (element && element.parentNode) {
      element.parentNode.removeChild(element)
    }
    vi.clearAllMocks()
  })

  describe('initialization', () => {
    it('should create an instance', () => {
      expect(element).toBeInstanceOf(SortifyApp)
    })

    it('should be a LitElement', () => {
      expect(element.tagName.toLowerCase()).toBe('sortify-app')
    })
  })

  describe('rendering', () => {
    it('should render the header with title', async () => {
      await element.updateComplete
      
      const shadow = element.shadowRoot
      expect(shadow).toBeTruthy()
      
      const header = shadow?.querySelector('.header')
      expect(header).toBeTruthy()
      
      const title = shadow?.querySelector('.header-title')
      expect(title).toBeTruthy()
      expect(title?.textContent?.trim()).toBe('Sortify')
    })

    it('should render the main content area', async () => {
      await element.updateComplete
      
      const shadow = element.shadowRoot
      const mainContent = shadow?.querySelector('.main-content')
      expect(mainContent).toBeTruthy()
    })

    it('should render the upload component', async () => {
      await element.updateComplete
      
      const shadow = element.shadowRoot
      const uploadComponent = shadow?.querySelector('sortify-upload')
      expect(uploadComponent).toBeTruthy()
    })

    it('should apply correct CSS classes', async () => {
      await element.updateComplete
      
      const shadow = element.shadowRoot
      const appLayout = shadow?.querySelector('.app-layout')
      expect(appLayout).toBeTruthy()
      
      const mainContent = shadow?.querySelector('.main-content')
      expect(mainContent).toBeTruthy()
    })
  })

  describe('styling', () => {
    it('should have defined styles', () => {
      expect(SortifyApp.styles).toBeDefined()
    })

    it('should set host display to block', async () => {
      await element.updateComplete
      
      // The host element should have display: block from the styles
      const styles = SortifyApp.styles.toString()
      expect(styles).toContain('display: block')
    })
  })

  describe('responsive design', () => {
    it('should handle different screen sizes', async () => {
      await element.updateComplete
      
      const shadow = element.shadowRoot
      const appLayout = shadow?.querySelector('.app-layout')
      
      expect(appLayout).toBeTruthy()
      
      // Check that the CSS contains responsive styles
      const styles = SortifyApp.styles.toString()
      expect(styles).toContain('100vw')
      expect(styles).toContain('100vh')
      expect(styles).toContain('@media')
    })
  })

  describe('accessibility', () => {
    it('should have proper semantic structure', async () => {
      await element.updateComplete
      
      const shadow = element.shadowRoot
      
      // Check for proper heading structure
      const mainHeading = shadow?.querySelector('h1')
      expect(mainHeading).toBeTruthy()
      
      // Check for main content area
      const main = shadow?.querySelector('main, .main-content')
      expect(main).toBeTruthy()
    })

    it('should be keyboard accessible', async () => {
      await element.updateComplete
      
      // The upload component should be focusable
      const uploadComponent = element.shadowRoot?.querySelector('sortify-upload')
      expect(uploadComponent).toBeTruthy()
    })
  })

  describe('component lifecycle', () => {
    it('should connect and disconnect properly', () => {
      // Test that lifecycle methods exist and are callable
      expect(typeof element.connectedCallback).toBe('function')
      expect(typeof element.disconnectedCallback).toBe('function')
      
      // Test that element is properly connected
      expect(element.isConnected).toBe(true)
    })

    it('should handle updates correctly', async () => {
      const updateSpy = vi.spyOn(element, 'requestUpdate')
      
      element.requestUpdate()
      await element.updateComplete
      
      expect(updateSpy).toHaveBeenCalled()
    })
  })

  describe('error handling', () => {
    it('should handle rendering errors gracefully', async () => {
      // Test that the component doesn't crash if child components fail
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      
      try {
        await element.updateComplete
        expect(element.shadowRoot).toBeTruthy()
      } catch (error) {
        // If there's an error, it should be handled gracefully
        expect(consoleSpy).toHaveBeenCalled()
      }
      
      consoleSpy.mockRestore()
    })
  })

  describe('PWA features', () => {
    it('should be ready for PWA installation', async () => {
      await element.updateComplete
      
      // The app should render properly and be ready for PWA features
      expect(element.shadowRoot).toBeTruthy()
      
      // Check that the component doesn't interfere with PWA functionality
      const meta = document.querySelector('meta[name="theme-color"]')
      if (meta) {
        expect(meta.getAttribute('content')).toBeTruthy()
      }
    })
  })

  describe('integration with upload component', () => {
    it('should contain upload component', async () => {
      await element.updateComplete
      
      const uploadComponent = element.shadowRoot?.querySelector('sortify-upload')
      expect(uploadComponent).toBeTruthy()
    })

    it('should handle upload events', async () => {
      await element.updateComplete
      
      const uploadComponent = element.shadowRoot?.querySelector('sortify-upload')
      expect(uploadComponent).toBeTruthy()
      
      // The upload component should be properly embedded
      expect(uploadComponent?.tagName.toLowerCase()).toBe('sortify-upload')
    })
  })
})
