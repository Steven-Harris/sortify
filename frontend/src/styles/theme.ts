/**
 * Theme System for Sortify
 * Spotify-inspired color palette and design tokens
 */

import { css } from 'lit';

export const theme = css`
  :host {
    /* Color Palette */
    --color-primary: #164773;
    --color-secondary: #0B2B40;
    --color-accent-1: #1E5959;
    --color-accent-2: #3B8C6E;
    --color-success: #89D99D;
    
    /* Semantic Colors */
    --color-background: var(--color-secondary);
    --color-surface: #1a1a1a;
    --color-surface-elevated: #2a2a2a;
    --color-text-primary: #ffffff;
    --color-text-secondary: #b3b3b3;
    --color-text-muted: #6a6a6a;
    --color-border: #333333;
    --color-error: #e22134;
    --color-warning: #ffa500;
    
    /* Spacing */
    --spacing-xs: 4px;
    --spacing-sm: 8px;
    --spacing-md: 16px;
    --spacing-lg: 24px;
    --spacing-xl: 32px;
    --spacing-xxl: 48px;
    
    /* Border Radius */
    --border-radius-sm: 4px;
    --border-radius-md: 8px;
    --border-radius-lg: 12px;
    --border-radius-xl: 16px;
    
    /* Typography */
    --font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    --font-size-xs: 0.75rem;
    --font-size-sm: 0.875rem;
    --font-size-base: 1rem;
    --font-size-lg: 1.125rem;
    --font-size-xl: 1.25rem;
    --font-size-2xl: 1.5rem;
    --font-size-3xl: 2rem;
    
    --font-weight-normal: 400;
    --font-weight-medium: 500;
    --font-weight-semibold: 600;
    --font-weight-bold: 700;
    
    /* Shadows */
    --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.12);
    --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.16);
    --shadow-lg: 0 8px 16px rgba(0, 0, 0, 0.24);
    
    /* Transitions */
    --transition-fast: 150ms ease;
    --transition-normal: 250ms ease;
    --transition-slow: 350ms ease;
    
    /* Z-indexes */
    --z-dropdown: 100;
    --z-modal: 200;
    --z-tooltip: 300;
  }
`;

export const globalStyles = css`
  * {
    box-sizing: border-box;
  }
  
  body {
    margin: 0;
    padding: 0;
    font-family: var(--font-family);
    background-color: var(--color-background);
    color: var(--color-text-primary);
    overflow-x: hidden;
  }
  
  /* Utility Classes */
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
  
  .container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 var(--spacing-md);
  }
  
  .flex {
    display: flex;
  }
  
  .flex-col {
    flex-direction: column;
  }
  
  .items-center {
    align-items: center;
  }
  
  .justify-center {
    justify-content: center;
  }
  
  .justify-between {
    justify-content: space-between;
  }
  
  .gap-sm {
    gap: var(--spacing-sm);
  }
  
  .gap-md {
    gap: var(--spacing-md);
  }
  
  .gap-lg {
    gap: var(--spacing-lg);
  }
  
  .text-center {
    text-align: center;
  }
  
  .text-sm {
    font-size: var(--font-size-sm);
  }
  
  .text-lg {
    font-size: var(--font-size-lg);
  }
  
  .font-medium {
    font-weight: var(--font-weight-medium);
  }
  
  .font-semibold {
    font-weight: var(--font-weight-semibold);
  }
  
  .rounded {
    border-radius: var(--border-radius-md);
  }
  
  .rounded-lg {
    border-radius: var(--border-radius-lg);
  }
`;

/* Component-specific styles for reuse */
export const buttonStyles = css`
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--spacing-sm);
    padding: var(--spacing-sm) var(--spacing-md);
    border: none;
    border-radius: var(--border-radius-md);
    font-family: inherit;
    font-size: var(--font-size-base);
    font-weight: var(--font-weight-medium);
    text-decoration: none;
    cursor: pointer;
    transition: all var(--transition-fast);
    user-select: none;
  }
  
  .btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
  
  .btn-primary {
    background-color: var(--color-primary);
    color: white;
  }
  
  .btn-primary:hover:not(:disabled) {
    background-color: #1a5380;
    transform: translateY(-1px);
  }
  
  .btn-secondary {
    background-color: var(--color-surface-elevated);
    color: var(--color-text-primary);
    border: 1px solid var(--color-border);
  }
  
  .btn-secondary:hover:not(:disabled) {
    background-color: var(--color-surface);
    border-color: var(--color-accent-1);
  }
  
  .btn-success {
    background-color: var(--color-success);
    color: var(--color-secondary);
  }
  
  .btn-success:hover:not(:disabled) {
    background-color: #7bc98a;
  }
  
  .btn-sm {
    padding: 6px 12px;
    font-size: var(--font-size-sm);
  }
  
  .btn-lg {
    padding: 12px 24px;
    font-size: var(--font-size-lg);
  }
`;

export const cardStyles = css`
  .card {
    background-color: var(--color-surface-elevated);
    border: 1px solid var(--color-border);
    border-radius: var(--border-radius-lg);
    padding: var(--spacing-lg);
    box-shadow: var(--shadow-sm);
    transition: all var(--transition-normal);
  }
  
  .card:hover {
    border-color: var(--color-accent-1);
    box-shadow: var(--shadow-md);
  }
  
  .card-header {
    margin-bottom: var(--spacing-md);
  }
  
  .card-title {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-semibold);
    margin: 0 0 var(--spacing-sm) 0;
    color: var(--color-text-primary);
  }
  
  .card-subtitle {
    font-size: var(--font-size-sm);
    color: var(--color-text-secondary);
    margin: 0;
  }
`;
