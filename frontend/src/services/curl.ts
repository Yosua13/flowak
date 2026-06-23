/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { BackendFacet } from '../domain/types';

/**
 * Generates a standard cURL command based on the BackendFacet parameters.
 * INV: normalizes request body whitespace into a single line for -d flag.
 */
export function generateCurl(backendFacet: BackendFacet, baseUrl: string = 'https://api.flowak.com'): string {
  const method = backendFacet.method || 'GET';
  const endpoint = backendFacet.endpoint || '/api/endpoint';
  const fullUrl = endpoint.startsWith('http') ? endpoint : `${baseUrl}${endpoint}`;
  const authHeader = backendFacet.auth ? backendFacet.auth : 'Bearer YOUR_TOKEN_HERE';

  let curlParts = [
    `curl -X ${method} "${fullUrl}"`,
    `  -H "Authorization: ${authHeader}"`,
    `  -H "Content-Type: application/json"`,
  ];

  if (method !== 'GET' && backendFacet.request) {
    // Normalize whitespace of JSON payload into a clean single line for cURL compatibility
    try {
      const parsed = JSON.parse(backendFacet.request);
      const flattened = JSON.stringify(parsed);
      curlParts.push(`  -d '${flattened}'`);
    } catch (e) {
      // If parsing fails, just strip newlines and duplicate spaces
      const cleaned = backendFacet.request.replace(/\s+/g, ' ').trim();
      curlParts.push(`  -d '${cleaned}'`);
    }
  }

  return curlParts.join(' \\\n');
}
