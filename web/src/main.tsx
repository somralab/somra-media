import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { I18nextProvider } from 'react-i18next';
import { BrowserRouter } from 'react-router-dom';

import App from './App';
import i18n from './i18n';
import { AuthBootstrap } from './components/AuthBootstrap';
import { ThemeProvider } from './theme/ThemeProvider';
import { ToastProvider } from './components/ui/Toast';
import { createQueryClient } from './lib/queryClient';
import './styles/globals.css';

const queryClient = createQueryClient();

const container = document.getElementById('root');
if (!container) {
  throw new Error('Root container "#root" was not found in the document.');
}

createRoot(container).render(
  <StrictMode>
    <I18nextProvider i18n={i18n}>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider>
          <ToastProvider>
            <BrowserRouter>
              <AuthBootstrap>
                <App />
              </AuthBootstrap>
            </BrowserRouter>
          </ToastProvider>
        </ThemeProvider>
        {import.meta.env.DEV ? <ReactQueryDevtools initialIsOpen={false} /> : null}
      </QueryClientProvider>
    </I18nextProvider>
  </StrictMode>,
);
