import { LitElement, css, html } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { theme, globalStyles, buttonStyles, cardStyles } from './styles/theme.js';
import './components/upload.js';

/**
 * Main Sortify Application Component
 * Manages the overall app layout and navigation
 */
@customElement('sortify-app')
export class SortifyApp extends LitElement {
  static styles = [
    theme,
    globalStyles,
    buttonStyles,
    cardStyles,
    css`
      :host {
        display: block;
        min-height: 100vh;
        background-color: var(--color-background);
        color: var(--color-text-primary);
      }

      .app-layout {
        display: grid;
        grid-template-columns: 250px 1fr;
        grid-template-rows: auto 1fr;
        grid-template-areas: 
          "sidebar header"
          "sidebar main";
        min-height: 100vh;
      }

      .sidebar {
        grid-area: sidebar;
        background-color: var(--color-surface);
        border-right: 1px solid var(--color-border);
        padding: var(--spacing-lg);
      }

      .header {
        grid-area: header;
        background-color: var(--color-surface-elevated);
        border-bottom: 1px solid var(--color-border);
        padding: var(--spacing-md) var(--spacing-lg);
        display: flex;
        align-items: center;
        justify-content: space-between;
      }

      .main-content {
        grid-area: main;
        padding: var(--spacing-lg);
        overflow-y: auto;
      }

      .logo {
        display: flex;
        align-items: center;
        gap: var(--spacing-sm);
        margin-bottom: var(--spacing-xl);
      }

      .logo-icon {
        width: 32px;
        height: 32px;
        background: linear-gradient(135deg, var(--color-primary), var(--color-accent-2));
        border-radius: var(--border-radius-md);
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: var(--font-weight-bold);
        color: white;
      }

      .logo-text {
        font-size: var(--font-size-xl);
        font-weight: var(--font-weight-bold);
        color: var(--color-text-primary);
      }

      .nav-menu {
        list-style: none;
        padding: 0;
        margin: 0;
      }

      .nav-item {
        margin-bottom: var(--spacing-sm);
      }

      .nav-link {
        display: flex;
        align-items: center;
        gap: var(--spacing-sm);
        padding: var(--spacing-sm) var(--spacing-md);
        border-radius: var(--border-radius-md);
        text-decoration: none;
        color: var(--color-text-secondary);
        transition: all var(--transition-fast);
        cursor: pointer;
      }

      .nav-link:hover,
      .nav-link.active {
        background-color: var(--color-surface-elevated);
        color: var(--color-text-primary);
      }

      .nav-icon {
        width: 20px;
        height: 20px;
        opacity: 0.7;
      }

      .header-title {
        font-size: var(--font-size-2xl);
        font-weight: var(--font-weight-semibold);
        margin: 0;
      }

      .header-actions {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
      }

      .welcome-section {
        text-align: center;
        max-width: 600px;
        margin: var(--spacing-xxl) auto;
      }

      .welcome-title {
        font-size: var(--font-size-3xl);
        font-weight: var(--font-weight-bold);
        margin-bottom: var(--spacing-md);
        background: linear-gradient(135deg, var(--color-primary), var(--color-accent-2));
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
      }

      .welcome-subtitle {
        font-size: var(--font-size-lg);
        color: var(--color-text-secondary);
        margin-bottom: var(--spacing-xl);
      }

      @media (max-width: 768px) {
        .app-layout {
          grid-template-columns: 1fr;
          grid-template-areas: 
            "header"
            "main";
        }

        .sidebar {
          display: none;
        }
      }
    `
  ];

  @state()
  private activeView = 'upload';

  private navigationItems = [
    { id: 'upload', label: 'Upload', icon: '📤' },
    { id: 'browse', label: 'Browse', icon: '📁' },
    { id: 'search', label: 'Search', icon: '🔍' },
    { id: 'settings', label: 'Settings', icon: '⚙️' }
  ];

  render() {
    return html`
      <div class="app-layout">
        <aside class="sidebar">
          <div class="logo">
            <div class="logo-icon">S</div>
            <span class="logo-text">Sortify</span>
          </div>
          
          <nav>
            <ul class="nav-menu">
              ${this.navigationItems.map(item => html`
                <li class="nav-item">
                  <a 
                    class="nav-link ${this.activeView === item.id ? 'active' : ''}"
                    @click=${() => this.setActiveView(item.id)}
                  >
                    <span class="nav-icon">${item.icon}</span>
                    ${item.label}
                  </a>
                </li>
              `)}
            </ul>
          </nav>
        </aside>

        <header class="header">
          <h1 class="header-title">${this.getViewTitle()}</h1>
          <div class="header-actions">
            <button class="btn btn-secondary btn-sm">
              <span>ℹ️</span>
              Help
            </button>
          </div>
        </header>

        <main class="main-content">
          ${this.renderActiveView()}
        </main>
      </div>
    `;
  }

  private setActiveView(view: string) {
    this.activeView = view;
  }

  private getViewTitle(): string {
    const item = this.navigationItems.find(item => item.id === this.activeView);
    return item ? item.label : 'Sortify';
  }

  private renderActiveView() {
    switch (this.activeView) {
      case 'upload':
        return this.renderUploadView();
      case 'browse':
        return this.renderBrowseView();
      case 'search':
        return this.renderSearchView();
      case 'settings':
        return this.renderSettingsView();
      default:
        return this.renderUploadView();
    }
  }

  private renderUploadView() {
    return html`
      <div class="welcome-section">
        <h2 class="welcome-title">Welcome to Sortify</h2>
        <p class="welcome-subtitle">
          Automatically organize your photos and videos by date
        </p>
        
        <sortify-upload></sortify-upload>
      </div>
    `;
  }

  private renderBrowseView() {
    return html`
      <div class="welcome-section">
        <h2 class="welcome-title">Browse Your Media</h2>
        <p class="welcome-subtitle">
          Navigate through your organized photos and videos
        </p>
        
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Media Library</h3>
            <p class="card-subtitle">Organized by date</p>
          </div>
          
          <p>📂 Media browser component will be implemented here</p>
        </div>
      </div>
    `;
  }

  private renderSearchView() {
    return html`
      <div class="welcome-section">
        <h2 class="welcome-title">Search Media</h2>
        <p class="welcome-subtitle">
          Find your photos and videos quickly
        </p>
        
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Search</h3>
            <p class="card-subtitle">Search by filename, date, or metadata</p>
          </div>
          
          <p>🔍 Search component will be implemented here</p>
        </div>
      </div>
    `;
  }

  private renderSettingsView() {
    return html`
      <div class="welcome-section">
        <h2 class="welcome-title">Settings</h2>
        <p class="welcome-subtitle">
          Configure your Sortify preferences
        </p>
        
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Configuration</h3>
            <p class="card-subtitle">Customize your experience</p>
          </div>
          
          <p>⚙️ Settings component will be implemented here</p>
        </div>
      </div>
    `;
  }
}
