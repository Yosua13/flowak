import { describe, expect, it } from 'vitest';
import { generateOpenApi, generateRedactedCurl, redactSensitive } from './exportService';
import type { Module } from '../domain/types';

const module: Module = { id: 'm1', name: 'Checkout', nodes: [{ id: 'n1', type: 'process', label: 'Bayar', x: 0, y: 0, doc: {}, roles: { backend: { method: 'POST', endpoint: '/payments?debug=true', auth: 'bearer', request: '{"email":"a@b.test","token":"private"}', response: '{"access_token":"private"}' } } }], edges: [], schemaVersion: 7 };

describe('safe exports', () => {
  it('redacts sensitive values recursively', () => expect(redactSensitive({ token: 'private', child: { password: 'secret', safe: 'ok' } })).toEqual({ token: '{{REDACTED}}', child: { password: '{{REDACTED}}', safe: 'ok' } }));
  it('removes legacy auth and nested JSON secrets from canonical exports', () => {
    const cleaned = JSON.stringify(redactSensitive({ auth: 'Bearer private', endpoint: '/items?api_key=private', request: '{"token":"private","safe":"ok"}' }));
    expect(cleaned).not.toContain('private');
    expect(cleaned).toContain('{{API_TOKEN}}');
    expect(cleaned).toContain('safe');
  });
  it('exports typed OpenAPI without response examples and cURL variables', () => {
    const openapi = generateOpenApi(module);
    expect(openapi).not.toContain('private');
    expect(openapi).toContain('"type": "object"');
    expect(generateRedactedCurl(module)).toContain('{{API_TOKEN}}');
  });
});
