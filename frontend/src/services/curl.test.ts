import { describe, expect, it } from 'vitest';
import { generateCurl } from './curl';
describe('generateCurl', () => {
  it('uses a variable placeholder instead of an authorization value', () => {
    const command = generateCurl({ method: 'GET', endpoint: '/v1/items?api_key=private', auth: 'Bearer actual-secret' });
    expect(command).toContain('{{API_TOKEN}}');
    expect(command).not.toContain('actual-secret');
    expect(command).not.toContain('private');
  });
  it('redacts sensitive request example fields', () => {
    const command = generateCurl({ method: 'POST', endpoint: '/v1/items', auth: '{{BEARER_TOKEN}}', request: '{"name":"safe","token":"private"}' });
    expect(command).toContain('{{BEARER_TOKEN}}');
    expect(command).toContain('"name":"safe"');
    expect(command).not.toContain('private');
  });
});
