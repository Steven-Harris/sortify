import { LitElement, css, html } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import './components/upload.js';
import './components/browse.js';

/**
 * Main Sortify Application Component
 * Manages the overall app layout and navigation
 */
@customElement('sortify-app')
export class SortifyApp extends LitElement {
  static styles = css`
    @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
    
    :host {
      display: block;
      width: 100vw;
      height: 100vh;
      min-height: 100vh;
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
      color: #f8fafc;
      margin: 0;
      padding: 0;
      --primary-400: #60a5fa;
      --primary-500: #3b82f6;
      --primary-600: #2563eb;
      --primary-50: #1e3a8a;
      --slate-50: #f8fafc;
      --slate-100: #f1f5f9;
      --slate-200: #e2e8f0;
      --slate-300: #cbd5e1;
      --slate-400: #94a3b8;
      --slate-500: #64748b;
      --slate-600: #475569;
      --slate-700: #334155;
      --slate-800: #1e293b;
      --slate-900: #0f172a;
      --slate-950: #020617;
    }

    .app-layout {
      display: grid;
      grid-template-columns: 16rem 1fr; /* w-64 + flex-1 */
      width: 100vw;
      height: 100vh;
      min-height: 100vh;
      background: var(--slate-800);
      box-shadow: none; /* Remove shadow since we're fullscreen */
      border-radius: 0; /* Remove border radius */
      margin: 0; /* Remove margin */
      overflow: hidden;
      border: none; /* Remove border */
    }

    .sidebar {
      background: linear-gradient(180deg, var(--slate-900) 0%, var(--slate-800) 100%);
      border-right: 1px solid var(--slate-700);
      padding: 2rem 1.5rem; /* py-8 px-6 */
      display: flex;
      flex-direction: column;
    }

    .logo-section {
      margin-bottom: 3rem; /* mb-12 */
      text-align: center;
    }

    .logo {
      font-size: 1.875rem; /* text-3xl */
      font-weight: 800; /* font-extrabold */
      background: linear-gradient(135deg, var(--primary-400), var(--primary-500));
      -webkit-background-clip: text;
      background-clip: text;
      color: transparent;
      margin-bottom: 0.5rem; /* mb-2 */
    }

    .logo-subtitle {
      font-size: 0.875rem; /* text-sm */
      color: var(--slate-400);
      font-weight: 500;
    }

    .nav-menu {
      flex: 1;
    }

    .nav-section {
      margin-bottom: 2rem; /* mb-8 */
    }

    .nav-section-title {
      font-size: 0.75rem; /* text-xs */
      font-weight: 600; /* font-semibold */
      color: var(--slate-500);
      text-transform: uppercase;
      letter-spacing: 0.05em; /* tracking-wider */
      margin-bottom: 0.75rem; /* mb-3 */
      padding-left: 0.75rem; /* pl-3 */
    }

    .nav-item {
      display: flex;
      align-items: center;
      gap: 0.75rem; /* gap-3 */
      padding: 0.75rem; /* p-3 */
      border-radius: 0.75rem; /* rounded-xl */
      font-weight: 500; /* font-medium */
      color: var(--slate-300);
      cursor: pointer;
      transition: all 0.2s ease-in-out;
      margin-bottom: 0.25rem; /* mb-1 */
    }

    .nav-item:hover {
      background: var(--slate-700);
      color: var(--slate-100);
      transform: translateX(0.25rem); /* translate-x-1 */
    }

    .nav-item.active {
      background: linear-gradient(135deg, var(--primary-600), var(--primary-500));
      color: white;
      box-shadow: 0 4px 14px 0 rgba(59, 130, 246, 0.4);
      transform: translateX(0.25rem); /* translate-x-1 */
    }

    .nav-icon {
      width: 1.25rem; /* w-5 */
      height: 1.25rem; /* h-5 */
      flex-shrink: 0;
    }

    .main-content {
      background: var(--slate-900);
      overflow-y: auto;
      overflow-x: hidden;
      position: relative;
      height: 100vh;
    }

    .header {
      background: rgba(15, 23, 42, 0.9);
      border-bottom: 1px solid var(--slate-700);
      padding: 1.5rem 2rem; /* py-6 px-8 */
      display: flex;
      align-items: center;
      justify-content: space-between;
      position: sticky;
      top: 0;
      z-index: 10;
      backdrop-filter: blur(8px);
    }

    .header-title {
      font-size: 2rem; /* text-3xl */
      font-weight: 700; /* font-bold */
      color: var(--slate-100);
      margin: 0;
    }

    .header-actions {
      display: flex;
      align-items: center;
      gap: 1rem; /* gap-4 */
    }

    .header-btn {
      background: var(--slate-800);
      border: 1px solid var(--slate-600);
      color: var(--slate-300);
      border-radius: 0.75rem; /* rounded-xl */
      padding: 0.5rem 1rem; /* py-2 px-4 */
      font-size: 0.875rem; /* text-sm */
      font-weight: 500; /* font-medium */
      cursor: pointer;
      transition: all 0.2s ease-in-out;
      font-family: inherit;
      display: flex;
      align-items: center;
      gap: 0.5rem; /* gap-2 */
    }

    .header-btn:hover {
      background: var(--slate-700);
      color: var(--slate-100);
      transform: translateY(-1px);
    }

    .content-area {
      padding: 2rem; /* p-8 */
      min-height: calc(100vh - 10rem);
    }

    .welcome-section {
      text-align: center;
      max-width: 48rem; /* max-w-4xl */
      margin: 0 auto 3rem; /* mx-auto mb-12 */
    }

    .welcome-title {
      font-size: 3rem; /* text-5xl */
      font-weight: 800; /* font-extrabold */
      margin-bottom: 1rem; /* mb-4 */
      background: linear-gradient(135deg, var(--slate-100), var(--slate-300));
      -webkit-background-clip: text;
      background-clip: text;
      color: transparent;
      line-height: 1.1;
    }

    .welcome-subtitle {
      font-size: 1.25rem; /* text-xl */
      color: var(--slate-400);
      font-weight: 400;
      line-height: 1.6;
      margin-bottom: 2rem; /* mb-8 */
    }

    .feature-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(20rem, 1fr)); /* grid-cols-auto */
      gap: 2rem; /* gap-8 */
      margin-top: 3rem; /* mt-12 */
    }

    .feature-card {
      background: var(--slate-800);
      padding: 2rem; /* p-8 */
      border-radius: 1.5rem; /* rounded-3xl */
      border: 1px solid var(--slate-700);
      text-align: center;
      transition: all 0.2s ease-in-out;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3);
    }

    .feature-card:hover {
      transform: translateY(-0.5rem); /* -translate-y-2 */
      box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.4), 0 10px 10px -5px rgba(0, 0, 0, 0.2);
      border-color: var(--slate-600);
    }

    .feature-icon {
      width: 4rem; /* w-16 */
      height: 4rem; /* h-16 */
      margin: 0 auto 1.5rem; /* mx-auto mb-6 */
      background: linear-gradient(135deg, var(--primary-500), var(--primary-400));
      border-radius: 1rem; /* rounded-2xl */
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 1.5rem; /* text-2xl */
      color: white;
      box-shadow: 0 8px 25px 0 rgba(59, 130, 246, 0.4);
    }

    .feature-title {
      font-size: 1.25rem; /* text-xl */
      font-weight: 600; /* font-semibold */
      color: var(--slate-100);
      margin-bottom: 0.75rem; /* mb-3 */
    }

    .feature-description {
      color: var(--slate-400);
      line-height: 1.6;
    }

    @media (max-width: 1024px) {
      .app-layout {
        grid-template-columns: 1fr;
        margin: 0; /* Remove margin on mobile */
        border-radius: 0; /* Remove border radius on mobile */
      }

      .sidebar {
        display: none;
      }

      .header {
        padding: 1rem 1.5rem; /* py-4 px-6 */
      }

      .header-title {
        font-size: 1.5rem; /* text-2xl */
      }

      .content-area {
        padding: 1rem; /* p-4 */
      }

      .welcome-title {
        font-size: 2rem; /* text-4xl */
      }

      .welcome-subtitle {
        font-size: 1.125rem; /* text-lg */
      }
    }
  `;

  @state()
  private activeView = 'upload';

  private navigationItems = [
    { id: 'upload', label: 'Upload', icon: '📤' },
    { id: 'browse', label: 'Browse', icon: '📁' }
  ];

  render() {
    return html`
      <div class="app-layout">
        <aside class="sidebar">
          <div class="logo-section">
            <div class="logo">Sortify</div>
            <div class="logo-subtitle">Media Organizer</div>
          </div>
          
          <nav class="nav-menu">
            <div class="nav-section">
              <div class="nav-section-title">Main</div>
              ${this.navigationItems.map(item => html`
                <div 
                  class="nav-item ${this.activeView === item.id ? 'active' : ''}"
                  @click=${() => this.setActiveView(item.id)}
                >
                  <span class="nav-icon">${item.icon}</span>
                  ${item.label}
                </div>
              `)}
            </div>
          </nav>
        </aside>

        <main class="main-content">
          <header class="header">
            <h1 class="header-title">${this.getViewTitle()}</h1>
          </header>

          <div class="content-area">
            ${this.renderActiveView()}
          </div>
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
      default:
        return this.renderUploadView();
    }
  }

  private renderUploadView() {
    return html`
      <div class="welcome-section">
        <h2 class="welcome-title">Welcome to Sortify</h2>
        <p class="welcome-subtitle">
          Automatically organize your photos and videos by date and metadata
        </p>
        
        <sortify-upload></sortify-upload>
      </div>
    `;
  }

  private renderBrowseView() {
    return html`
      <sortify-browse></sortify-browse>
  }

}
          <div class="feature-card">
            <div class="feature-icon">📂</div>
            <h3 class="feature-title">Organized by Date</h3>
            <p class="feature-description">
              Your media is automatically organized into folders by year, month, and day
            </p>
          </div>
          
          <div class="feature-card">
            <div class="feature-icon">🔍</div>
            <h3 class="feature-title">Smart Search</h3>
            <p class="feature-description">
              Find photos by location, camera model, date range, or any metadata
            </p>
          </div>
          
          <div class="feature-card">
            <div class="feature-icon">�</div>
            <h3 class="feature-title">Camera & Settings</h3>
            <p class="feature-description">
              Filter by camera model, lens, shooting settings, or technical details
            </p>
          </div>
          
          <div class="feature-card">
            <div class="feature-icon">�</div>
            <h3 class="feature-title">Media Stats</h3>
            <p class="feature-description">
              View detailed statistics about your photo and video collection
            </p>
          </div>
        </div>
      </div>
    `;
  }

}
