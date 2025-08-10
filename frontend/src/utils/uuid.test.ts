import { describe, it, expect } from 'vitest';
import { generateUUID, generateSimpleId } from '../utils/uuid.js';

describe('UUID Utils', () => {
  describe('generateUUID', () => {
    it('should generate a valid UUID v4 format', () => {
      const uuid = generateUUID();
      
      // Check UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
      const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
      expect(uuid).toMatch(uuidRegex);
    });

    it('should generate unique UUIDs', () => {
      const uuid1 = generateUUID();
      const uuid2 = generateUUID();
      
      expect(uuid1).not.toBe(uuid2);
    });

    it('should generate UUIDs of correct length', () => {
      const uuid = generateUUID();
      expect(uuid).toHaveLength(36); // 32 chars + 4 hyphens
    });
  });

  describe('generateSimpleId', () => {
    it('should generate a string', () => {
      const id = generateSimpleId();
      expect(typeof id).toBe('string');
      expect(id.length).toBeGreaterThan(0);
    });

    it('should generate unique IDs', () => {
      const id1 = generateSimpleId();
      const id2 = generateSimpleId();
      
      expect(id1).not.toBe(id2);
    });
  });
});
