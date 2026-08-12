import { LitElement, css, html } from 'lit'
import { customElement, property, state } from 'lit/decorators.js'
import { registerSW } from 'virtual:pwa-register'

/**
 * PWA Badge element.
 */
@customElement('pwa-badge')
export class PwaBadge extends LitElement {
    @property()
    private _period = 60 * 60 * 1000 // check for updates every hour
    @property()
    private _swActivated = false
    @state()
    private _offlineReady = false
    @state()
    private _needRefresh = false
    @property()
    private _updateServiceWorker: undefined | ((reloadPage?: boolean) => Promise<void>)

    firstUpdated() {
        this._updateServiceWorker = registerSW({
            immediate: true,
            
            onNeedRefresh: () => (this._needRefresh = true),
            onRegisteredSW: this._onRegisteredSW
        })
    }

    render() {
        const classes: string[] = []
        if (this._offlineReady)
            classes.push('show')
        else if (this._needRefresh) {
            classes.push('show', 'refresh')
        }
        const message = this._offlineReady
            ? 'App ready to work offline'
            : this._needRefresh
                ? 'New content available, click on reload button to update'
                : ''
        return html`
            <div
                id="pwa-toast"
                role="alert"
                aria-labelledby="toast-message"
                class=${classes.join(' ')}
            >
                <div class="message">
                    <span id="toast-message">${message}</span>
                </div>
                <div class="buttons">
                    <button id="pwa-refresh" type="button" @click=${this._refreshApp}>
                        Reload
                    </button>
                    <button id="pwa-close" type="button" @click=${this._closeBadge}>
                        Close
                    </button>
                </div>
            </div>
    `
    }

    private _refreshApp() {
        if (this._updateServiceWorker && this._needRefresh)
            this._updateServiceWorker()
    }

    private _closeBadge() {
        this._offlineReady = false
        this._needRefresh = false
    }

    private _onRegisteredSW(swUrl: string, r?: ServiceWorkerRegistration) {
        if (this._period <= 0) return
        if (r?.active?.state === 'activated') {
            this._swActivated = true
            this._registerPeriodicSync(swUrl, r)
        }
        else if (r?.installing) {
            r.installing.addEventListener('statechange', (e) => {
                const sw = e.target as ServiceWorker
                this._swActivated = sw.state === 'activated'
                if (this._swActivated)
                    this._registerPeriodicSync(swUrl, r)
            })
        }
    }

    private _registerPeriodicSync(swUrl: string, r: ServiceWorkerRegistration) {
        if (this._period <= 0) return

        setInterval(async () => {
            if ('onLine' in navigator && !navigator.onLine)
                return

            const resp = await fetch(swUrl, {
                cache: 'no-store',
                headers: {
                    'cache': 'no-store',
                    'cache-control': 'no-cache',
                },
            })

            if (resp?.status === 200)
                await r.update()
        }, this._period)
    }

    static styles = css`
    :host {
      max-width: 0;
      margin: 0;
      padding: 0;
    }

    #pwa-toast {
        visibility: hidden;
        opacity: 0;
        transform: translateY(10px);
        position: fixed;
        right: 0;
        bottom: 0;
        margin: 16px;
        padding: 14px 16px;
        max-width: min(22rem, calc(100vw - 32px));
        border: 1px solid var(--stroke-strong, rgba(255, 255, 255, 0.16));
        border-radius: var(--radius-md, 0.875rem);
        background: rgba(12, 17, 32, 0.86);
        backdrop-filter: blur(16px) saturate(160%);
        -webkit-backdrop-filter: blur(16px) saturate(160%);
        color: var(--text-primary, #f2f6ff);
        font-family: var(--font-body, system-ui, sans-serif);
        font-size: 0.875rem;
        z-index: 100;
        text-align: left;
        box-shadow: var(--shadow-lg, 0 30px 60px -25px rgba(2, 4, 12, 1));
        display: grid;
        transition: opacity 0.25s var(--ease, ease), transform 0.25s var(--ease, ease),
          visibility 0.25s var(--ease, ease);
    }
    #pwa-toast .message {
        margin-bottom: 12px;
        color: var(--text-secondary, #a8b6d1);
        line-height: 1.5;
    }
    #pwa-toast .buttons {
        display: flex;
        gap: 8px;
    }
    #pwa-toast button {
        font-family: inherit;
        font-size: 0.8125rem;
        font-weight: 600;
        border-radius: 999px;
        padding: 6px 14px;
        cursor: pointer;
        border: 1px solid var(--stroke, rgba(255, 255, 255, 0.09));
        background: var(--surface-1, rgba(255, 255, 255, 0.05));
        color: var(--text-secondary, #a8b6d1);
        outline: none;
        transition: background 0.2s var(--ease, ease), color 0.2s var(--ease, ease);
    }
    #pwa-toast button:hover {
        background: var(--surface-2, rgba(255, 255, 255, 0.08));
        color: var(--text-primary, #f2f6ff);
    }
    #pwa-toast.show {
        visibility: visible;
        opacity: 1;
        transform: none;
    }
    button#pwa-refresh {
        display: none;
        background: var(--accent-gradient, linear-gradient(120deg, #22d3ee, #6366f1));
        color: #06070f;
        border-color: transparent;
    }
    #pwa-toast.show.refresh button#pwa-refresh {
        display: block;
    }
  `
}

declare global {
    interface HTMLElementTagNameMap {
        'pwa-badge': PwaBadge
    }
}
