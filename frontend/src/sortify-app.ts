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
    :host {
      display: block;
      width: 100vw;
      height: 100vh;
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
      background: rgba(15, 23, 42, 0.95);
      border-bottom: 1px solid var(--slate-600);
      padding: 2rem;
      position: sticky;
      top: 0;
      z-index: 10;
      backdrop-filter: blur(12px);
      display: flex;
      justify-content: center;
      align-items: center;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
    }

    .header-title {
      font-size: 2.5rem;
      font-weight: 800;
      color: var(--slate-50);
      margin: 0;
      text-align: center;
      background: linear-gradient(135deg, var(--primary-400), var(--primary-600));
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
      text-shadow: 0 0 30px rgba(59, 130, 246, 0.3);
      letter-spacing: -0.025em;
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
        padding: 1.5rem 1rem;
      }
      .header-title {
        font-size: 2rem;
      }
      .content-area {
        padding: 1rem;
      }
    }

    @media (max-width: 640px) {
      .header {
        padding: 1rem;
      }
      .header-title {
        font-size: 1.75rem;
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
