import { LitElement, css, html } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import './components/upload.js';

/**
 * Main Sortify Application Component
 * Manages the overall app layout and navigation
 */
@customElement('sortify-app')
export class SortifyApp extends LitElement {
  static styles = css`
    @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&display=swap');
    
    :host {
      display: block;
      width: 100vw;
      height: 100vh;
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
      background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
      color: #f8fafc;
      margin: 0;
      padding: 0;
      --primary: #3b82f6;
      --slate-100: #f1f5f9;
      --slate-400: #94a3b8;
      --slate-700: #334155;
      --slate-800: #1e293b;
      --slate-900: #0f172a;
    }

    .app-layout {
      width: 100vw;
      height: 100vh;
      background: var(--slate-800);
      overflow: hidden;
    }

    .main-content {
      background: var(--slate-900);
      height: 100vh;
      overflow-y: auto;
    }

    .header {
      background: rgba(15, 23, 42, 0.9);
      border-bottom: 1px solid var(--slate-700);
      padding: 1.5rem 2rem;
      position: sticky;
      top: 0;
      z-index: 10;
      backdrop-filter: blur(8px);
    }

    .header-title {
      font-size: 2rem;
      font-weight: 700;
      color: var(--slate-100);
      margin: 0;
    }

    .content-area {
      padding: 2rem;
      min-height: calc(100vh - 6rem);
    }

    .welcome-section {
      text-align: center;
      max-width: 48rem;
      margin: 0 auto;
    }

    @media (max-width: 1024px) {
      .header {
        padding: 1rem 1.5rem;
      }
      .header-title {
        font-size: 1.5rem;
      }
      .content-area {
        padding: 1rem;
      }
    }
  `;

  @state()
  private activeView = 'upload';

  render() {
    return html`
      <div class="app-layout">
        <main class="main-content">
          <header class="header">
            <h1 class="header-title">Sortify</h1>
          </header>

          <div class="content-area">
            ${this.renderActiveView()}
          </div>
        </main>
      </div>
    `;
  }

  private renderActiveView() {
    switch (this.activeView) {
      case 'upload':
        return this.renderUploadView();
      default:
        return this.renderUploadView();
    }
  }

  private renderUploadView() {
    return html`
      <div class="welcome-section">
        <sortify-upload></sortify-upload>
      </div>
    `;
  }

}
