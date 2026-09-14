import { describe, expect, it } from 'vitest';
import { generateCurl } from './curl';
describe('generateCurl', () => {
  it('uses a variable placeholder instead of an authorization value', () => {
    expect(generateCurl({ method: 'GET', endpoint: '/v1/items', auth: 'Bearer actual-secret' })).toContain('{{BEARER_ACTUAL_SECRET_SECRET}}');
  });
});
