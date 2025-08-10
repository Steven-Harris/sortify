/**
 * Utility functions for generating unique IDs with fallbacks
 */

/**
 * Generate a UUID v4 with crypto.randomUUID() fallback
 * Handles environments without secure context or crypto support
 */
export function generateUUID(): string {
  // Try crypto.randomUUID() first (most secure, requires secure context)
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    try {
      return crypto.randomUUID();
    } catch (error) {
      console.warn('crypto.randomUUID() failed, falling back to manual generation:', error);
    }
  }

  // Fallback: Use crypto.getRandomValues() if available
  if (typeof crypto !== 'undefined' && typeof crypto.getRandomValues === 'function') {
    try {
      return generateUUIDFromRandomValues();
    } catch (error) {
      console.warn('crypto.getRandomValues() failed, falling back to Math.random():', error);
    }
  }

  // Final fallback: Use Math.random() (less secure but works everywhere)
  return generateUUIDFromMathRandom();
}

/**
 * Generate UUID using crypto.getRandomValues()
 */
function generateUUIDFromRandomValues(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  
  // Set version (4) and variant bits
  bytes[6] = (bytes[6] & 0x0f) | 0x40; // Version 4
  bytes[8] = (bytes[8] & 0x3f) | 0x80; // Variant 10

  // Convert to hex string with hyphens
  const hex = Array.from(bytes)
    .map(b => b.toString(16).padStart(2, '0'))
    .join('');
  
  return [
    hex.slice(0, 8),
    hex.slice(8, 12),
    hex.slice(12, 16),
    hex.slice(16, 20),
    hex.slice(20, 32)
  ].join('-');
}

/**
 * Generate UUID using Math.random() (fallback)
 */
function generateUUIDFromMathRandom(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

/**
 * Generate a simple unique ID for non-critical use cases
 */
export function generateSimpleId(): string {
  return Date.now().toString(36) + Math.random().toString(36).substr(2);
}
