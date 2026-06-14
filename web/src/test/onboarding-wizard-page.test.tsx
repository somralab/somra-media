import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { I18nextProvider } from 'react-i18next';
import OnboardingWizardPage from '@/pages/OnboardingWizardPage';
import i18n from '@/i18n';
import { TestProviders } from './testUtils';

const advanceMutate = vi.fn().mockResolvedValue({ phase: 'admin' });

vi.mock('@/api/hooks/useOnboarding', () => ({
  useOnboardingStatus: () => ({
    data: { phase: 'language', completed: false, setupRequired: true },
    isLoading: false,
  }),
  useSystemDetect: () => ({
    data: { cpuCores: 4, memoryBytes: 8_000_000_000, gpuPresent: false, paths: [] },
  }),
  useAdvanceOnboarding: () => ({ mutateAsync: advanceMutate, isPending: false }),
  useCompleteOnboarding: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
}));

describe('OnboardingWizardPage', () => {
  it('renders language step and advances on continue', async () => {
    await i18n.changeLanguage('en-US');
    render(
      <TestProviders>
        <I18nextProvider i18n={i18n}>
          <MemoryRouter>
            <OnboardingWizardPage />
          </MemoryRouter>
        </I18nextProvider>
      </TestProviders>,
    );

    expect(screen.getByText(/welcome to somra/i)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /continue/i }));
    await waitFor(() => expect(advanceMutate).toHaveBeenCalled());
  });
});
