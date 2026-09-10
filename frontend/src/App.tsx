/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import AppShell from './components/AppShell';
import { QueryClientProvider } from '@tanstack/react-query';
import { queryClient } from './services/queryClient';

export default function App() {
  return <QueryClientProvider client={queryClient}><AppShell /></QueryClientProvider>;
}
