import { describe, expect, it, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { I18nextProvider } from 'react-i18next';
import OnboardingWizardPage from '@/pages/OnboardingWizardPage';
import i18n from '@/i18n';
import { TestProviders, createTestQueryClient } from './testUtils';
import { ToastProvider } from '@/components/ui/Toast';
import { setupAdmin } from '@/api/endpoints/auth';
import { createLibrary } from '@/api/endpoints/library';

const advanceMutate = vi.fn().mockResolvedValue({ phase: 'admin' });
const { mockPhase } = vi.hoisted(() => ({ mockPhase: { value: 'language' as string } }));

vi.mock('@/api/endpoints/auth', () => ({
  setupAdmin: vi.fn(),
}));

vi.mock('@/api/endpoints/library', () => ({
  createLibrary: vi.fn(),
  listLibraries: vi.fn().mockResolvedValue([]),
  triggerScan: vi.fn(),
}));

vi.mock('@/api/hooks/useOnboarding', () => ({
  useOnboardingStatus: () => ({
    data: {
      phase: mockPhase.value,
      completed: false,
      setupRequired: true,
      smartDefaults: { maxConcurrentTranscodes: 2, scanCron: '0 3 * * *' },
    },
    isLoading: false,
  }),
  useSystemDetect: () => ({
    data: {
      cpuCores: 4,
      memoryBytes: 8_000_000_000,
      gpuPresent: true,
      accelerators: [
        {
          id: 'qsv',
          available: true,
          devicePresent: true,
          encodeCodecs: ['h264_qsv'],
          decodeCodecs: [],
        },
      ],
      recommendedAccelerator: 'qsv',
      paths: [],
    },
  }),
  useAdvanceOnboarding: () => ({ mutateAsync: advanceMutate, isPending: false }),
  useCompleteOnboarding: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
}));

function renderWizard(client = createTestQueryClient()): ReturnType<typeof render> {
  return render(
    <TestProviders client={client}>
      <I18nextProvider i18n={i18n}>
        <ToastProvider>
          <MemoryRouter>
            <OnboardingWizardPage />
          </MemoryRouter>
        </ToastProvider>
      </I18nextProvider>
    </TestProviders>,
  );
}

describe('OnboardingWizardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders language step and advances on continue', async () => {
    mockPhase.value = 'language';
    await i18n.changeLanguage('en-US');
    renderWizard();

    expect(screen.getByText(/welcome to somra/i)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /continue/i }));
    await waitFor(() => expect(advanceMutate).toHaveBeenCalled());
  });

  it('renders defaults step with recommendations', async () => {
    mockPhase.value = 'defaults';
    await i18n.changeLanguage('en-US');
    renderWizard();

    expect(screen.getByText(/recommended settings/i)).toBeInTheDocument();
    expect(screen.getByText(/4 CPU cores/i)).toBeInTheDocument();
  });

  it('renders scan step progress area', async () => {
    mockPhase.value = 'scan';
    await i18n.changeLanguage('en-US');
    renderWizard();

    expect(screen.getByText(/scanning your library/i)).toBeInTheDocument();
  });

  it('renders admin step fields', async () => {
    mockPhase.value = 'admin';
    await i18n.changeLanguage('en-US');
    renderWizard();

    expect(screen.getByText(/create admin account/i)).toBeInTheDocument();
  });

  it('renders complete step actions', async () => {
    mockPhase.value = 'complete';
    await i18n.changeLanguage('en-US');
    renderWizard();

    expect(screen.getByText(/you're all set/i)).toBeInTheDocument();
  });

  it('prefills library path for local dev', async () => {
    mockPhase.value = 'library';
    await i18n.changeLanguage('en-US');
    renderWizard();

    expect(screen.getByDisplayValue('./deploy/media')).toBeInTheDocument();
  });

  it('invalidates onboarding queries after admin setup', async () => {
    mockPhase.value = 'admin';
    await i18n.changeLanguage('en-US');
    const client = createTestQueryClient();
    const invalidateSpy = vi.spyOn(client, 'invalidateQueries');
    vi.mocked(setupAdmin).mockResolvedValue({
      accessToken: 'tok',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      user: { id: '1', username: 'admin', roles: ['admin'], disabled: false },
    });

    renderWizard(client);

    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'admin' } });
    const passwordInput = document.querySelector('input[type="password"]');
    expect(passwordInput).not.toBeNull();
    fireEvent.change(passwordInput!, { target: { value: 'Admin1234!' } });
    fireEvent.click(screen.getByRole('button', { name: /continue/i }));

    await waitFor(() => {
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['onboarding-status'] });
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['setup-status'] });
    });
  });

  it('shows error toast when admin setup fails', async () => {
    mockPhase.value = 'admin';
    await i18n.changeLanguage('en-US');
    vi.mocked(setupAdmin).mockRejectedValue(new Error('conflict'));

    renderWizard();

    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'admin' } });
    const passwordInput = document.querySelector('input[type="password"]');
    expect(passwordInput).not.toBeNull();
    fireEvent.change(passwordInput!, { target: { value: 'short' } });
    fireEvent.click(screen.getByRole('button', { name: /continue/i }));

    await waitFor(() => {
      expect(screen.getByText(/could not create the admin account/i)).toBeInTheDocument();
    });
  });

  it('invalidates onboarding queries after library create without advance step', async () => {
    mockPhase.value = 'library';
    await i18n.changeLanguage('en-US');
    const client = createTestQueryClient();
    const invalidateSpy = vi.spyOn(client, 'invalidateQueries');
    vi.mocked(createLibrary).mockResolvedValue({
      id: 1,
      name: 'Test Library',
      kind: 'movie',
      paths: ['./deploy/media'],
      watchEnabled: true,
    });

    renderWizard(client);

    fireEvent.click(screen.getByRole('button', { name: /continue/i }));

    await waitFor(() => {
      expect(createLibrary).toHaveBeenCalled();
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['onboarding-status'] });
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['setup-status'] });
    });
    expect(advanceMutate).not.toHaveBeenCalled();
  });
});
