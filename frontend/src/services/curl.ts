/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { BackendFacet } from '../domain/types';
import { redactSensitive } from './exportService';

/**
 * Generates a standard cURL command based on the BackendFacet parameters.
 * INV: normalizes request body whitespace into a single line for -d flag.
 */
export function generateCurl(backendFacet: BackendFacet, baseUrl: string = 'https://api.flowak.com'): string {
  const method = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].includes(backendFacet.method || '') ? backendFacet.method : 'GET';
  const endpoint = backendFacet.endpoint || '/api/endpoint';
  const fullUrl = (endpoint.startsWith('http') ? endpoint : `${baseUrl}${endpoint}`).split('?')[0].split('#')[0];
  const authHeader = /^\{\{[A-Z][A-Z0-9_]*\}\}$/.test(backendFacet.auth || '') ? backendFacet.auth : '{{API_TOKEN}}';
  const quote = (value: string) => `'${value.replaceAll("'", "'\\''")}'`;

  let curlParts = [
    `curl -X ${method} ${quote(fullUrl)}`,
    `  -H ${quote(`Authorization: ${authHeader}`)}`,
    `  -H ${quote('Content-Type: application/json')}`,
  ];

  if (method !== 'GET' && backendFacet.request) {
    // Normalize whitespace of JSON payload into a clean single line for cURL compatibility
    try {
      const parsed = JSON.parse(backendFacet.request);
      const flattened = JSON.stringify(redactSensitive(parsed));
      curlParts.push(`  -d ${quote(flattened)}`);
    } catch (e) {
      // If parsing fails, just strip newlines and duplicate spaces
      curlParts.push(`  -d ${quote('{{REQUEST_BODY}}')}`);
    }
  }

  return curlParts.join(' \\\n');
}
