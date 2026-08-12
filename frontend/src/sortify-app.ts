import { LitElement, css, html, svg } from 'lit';
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
      font-family: var(--font-body, system-ui, sans-serif);
    }

    .app-layout {
      width: 100vw;
      height: 100vh;
      overflow: hidden;
      background: transparent;
    }

    .main-content {
      height: 100vh;
      overflow-y: auto;
      overflow-x: hidden;
    }

    /* ---------- Top bar ---------- */
    .header {
      position: sticky;
      top: 0;
      z-index: 20;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 1rem;
      padding: 0.875rem clamp(1rem, 4vw, 2.5rem);
      background: rgba(7, 11, 22, 0.6);
      border-bottom: 1px solid var(--stroke, rgba(255, 255, 255, 0.09));
      backdrop-filter: blur(18px) saturate(160%);
      -webkit-backdrop-filter: blur(18px) saturate(160%);
    }

    .brand {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      min-width: 0;
    }

    .brand-mark {
      width: 2.25rem;
      height: 2.25rem;
      border-radius: 0.75rem;
      display: grid;
      place-items: center;
      color: #06070f;
      background: var(--accent-gradient, linear-gradient(120deg, #22d3ee, #6366f1 45%, #c084fc));
      box-shadow: 0 10px 24px -10px rgba(99, 102, 241, 0.9);
      flex-shrink: 0;
    }

    .brand-mark svg {
      width: 1.15rem;
      height: 1.15rem;
    }

    .brand-text {
      display: flex;
      flex-direction: column;
      line-height: 1.15;
      min-width: 0;
    }

    .header-title {
      font-family: var(--font-display, system-ui, sans-serif);
      font-size: 1.375rem;
      font-weight: 700;
      letter-spacing: -0.03em;
      margin: 0;
      color: var(--text-primary, #f2f6ff);
    }

    .brand-tagline {
      font-size: 0.75rem;
      color: var(--text-muted, #6f7f9c);
      letter-spacing: 0.02em;
      white-space: nowrap;
    }

    .header-meta {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .status-pill {
      display: inline-flex;
      align-items: center;
      gap: 0.45rem;
      padding: 0.35rem 0.75rem;
      font-size: 0.75rem;
      font-weight: 500;
      color: var(--text-secondary, #a8b6d1);
      background: var(--surface-1, rgba(255, 255, 255, 0.05));
      border: 1px solid var(--stroke, rgba(255, 255, 255, 0.09));
      border-radius: 999px;
      white-space: nowrap;
    }

    .status-dot {
      width: 0.45rem;
      height: 0.45rem;
      border-radius: 999px;
      background: var(--success, #34d399);
      box-shadow: 0 0 0 3px rgba(52, 211, 153, 0.16);
      animation: breathe 2.4s ease-in-out infinite;
    }

    @keyframes breathe {
      0%, 100% { opacity: 1; transform: scale(1); }
      50% { opacity: 0.55; transform: scale(0.85); }
    }

    /* ---------- Hero ---------- */
    .content-area {
      padding: clamp(1.5rem, 4vw, 3.5rem) clamp(1rem, 4vw, 2.5rem) 4rem;
      min-height: calc(100vh - 4.5rem);
    }

    .welcome-section {
      max-width: 64rem;
      margin: 0 auto;
      display: flex;
      flex-direction: column;
      gap: 2rem;
    }

    .hero {
      text-align: center;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 1rem;
      animation: rise 0.6s var(--ease, ease) both;
    }

    .eyebrow {
      display: inline-flex;
      align-items: center;
      gap: 0.5rem;
      padding: 0.3rem 0.85rem 0.3rem 0.35rem;
      border-radius: 999px;
      background: var(--surface-1, rgba(255, 255, 255, 0.05));
      border: 1px solid var(--stroke, rgba(255, 255, 255, 0.09));
      font-size: 0.75rem;
      font-weight: 500;
      color: var(--text-secondary, #a8b6d1);
    }

    .eyebrow-badge {
      padding: 0.15rem 0.5rem;
      border-radius: 999px;
      background: var(--accent-gradient, linear-gradient(120deg, #22d3ee, #6366f1));
      color: #06070f;
      font-weight: 700;
      font-size: 0.6875rem;
      letter-spacing: 0.04em;
      text-transform: uppercase;
    }

    .hero-title {
      font-family: var(--font-display, system-ui, sans-serif);
      font-size: clamp(2.25rem, 6vw, 3.75rem);
      line-height: 1.02;
      letter-spacing: -0.04em;
      font-weight: 800;
      margin: 0;
      color: var(--text-primary, #f2f6ff);
      max-width: 18ch;
    }

    .hero-title em {
      font-style: normal;
      color: var(--brand-cyan, #22d3ee);
    }

    .hero-subtitle {
      margin: 0;
      font-size: clamp(0.9375rem, 1.5vw, 1.125rem);
      color: var(--text-secondary, #a8b6d1);
      max-width: 52ch;
      line-height: 1.6;
    }

    .feature-strip {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(13rem, 1fr));
      gap: 0.75rem;
      animation: rise 0.6s var(--ease, ease) 0.08s both;
    }

    .feature {
      display: flex;
      align-items: flex-start;
      gap: 0.75rem;
      padding: 0.9rem 1rem;
      border-radius: var(--radius-md, 0.875rem);
      background: var(--surface-0, rgba(255, 255, 255, 0.03));
      border: 1px solid var(--stroke, rgba(255, 255, 255, 0.09));
      transition: border-color 0.25s var(--ease, ease), background 0.25s var(--ease, ease),
        transform 0.25s var(--ease, ease);
    }

    .feature:hover {
      border-color: var(--stroke-strong, rgba(255, 255, 255, 0.16));
      background: var(--surface-1, rgba(255, 255, 255, 0.05));
      transform: translateY(-2px);
    }

    .feature-icon {
      width: 1.75rem;
      height: 1.75rem;
      flex-shrink: 0;
      border-radius: 0.5rem;
      display: grid;
      place-items: center;
      background: var(--accent-gradient-soft, rgba(99, 102, 241, 0.18));
      color: var(--brand-cyan, #22d3ee);
    }

    .feature-icon svg {
      width: 0.95rem;
      height: 0.95rem;
    }

    .feature-title {
      font-size: 0.8125rem;
      font-weight: 600;
      color: var(--text-primary, #f2f6ff);
      margin: 0 0 0.15rem;
    }

    .feature-copy {
      font-size: 0.75rem;
      color: var(--text-muted, #6f7f9c);
      margin: 0;
      line-height: 1.45;
    }

    .upload-slot {
      animation: rise 0.6s var(--ease, ease) 0.16s both;
    }

    @keyframes rise {
      from { opacity: 0; transform: translateY(14px); }
      to { opacity: 1; transform: none; }
    }

    @media (max-width: 1024px) {
      .brand-tagline {
        display: none;
      }
    }

    @media (max-width: 640px) {
      .header {
        padding: 0.75rem 1rem;
      }
      .header-title {
        font-size: 1.125rem;
      }
      .status-pill .status-label {
        display: none;
      }
      .content-area {
        padding: 1.5rem 1rem 3rem;
      }
    }

    @media (prefers-reduced-motion: reduce) {
      .hero,
      .feature-strip,
      .upload-slot {
        animation: none;
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
            <div class="brand">
              <div class="brand-mark" aria-hidden="true">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4"
                  stroke-linecap="round" stroke-linejoin="round">
                  <path d="M4 6h16" />
                  <path d="M7 12h10" />
                  <path d="M10 18h4" />
                </svg>
              </div>
              <div class="brand-text">
                <h1 class="header-title">Sortify</h1>
                <span class="brand-tagline">Photo &amp; video library organizer</span>
              </div>
            </div>

            <div class="header-meta">
              <span class="status-pill">
                <span class="status-dot" aria-hidden="true"></span>
                <span class="status-label">Ready to organize</span>
              </span>
            </div>
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
        <section class="hero">
          <span class="eyebrow">
            <span class="eyebrow-badge">New</span>
            Queued uploads with duplicate detection
          </span>
          <h2 class="hero-title">Drop the chaos in. Get a <em>sorted library</em> back.</h2>
          <p class="hero-subtitle">
            Sortify reads the metadata your camera already wrote, skips exact duplicates, and files
            every photo and video into a clean, date-based structure.
          </p>
        </section>

        <div class="feature-strip">
          ${this.renderFeature(
            svg`<path d="M4 7V5a1 1 0 0 1 1-1h3l2 2h9a1 1 0 0 1 1 1v2" /><path d="M3 10h18l-1.5 8.2a2 2 0 0 1-2 1.8H6.5a2 2 0 0 1-2-1.8Z" />`,
            'Metadata first',
            'Trusts EXIF capture dates before falling back to filenames.',
          )}
          ${this.renderFeature(
            svg`<path d="M20 6 9 17l-5-5" />`,
            'Duplicate aware',
            'Checksums every file so the same shot is never stored twice.',
          )}
          ${this.renderFeature(
            svg`<path d="M4 6h10" /><path d="M4 12h16" /><path d="M4 18h7" /><circle cx="18" cy="6" r="2" /><circle cx="15" cy="18" r="2" />`,
            'Batch ready',
            'Queue thousands of files, then pause or resume any time.',
          )}
        </div>

        <div class="upload-slot">
          <sortify-upload></sortify-upload>
        </div>
      </div>
    `;
  }

  private renderFeature(icon: unknown, title: string, copy: string) {
    return html`
      <div class="feature">
        <div class="feature-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round">
            ${icon}
          </svg>
        </div>
        <div>
          <p class="feature-title">${title}</p>
          <p class="feature-copy">${copy}</p>
        </div>
      </div>
    `;
  }
}
