import { describe, expect, it, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import App from '@/App';
import { TestProviders } from './testUtils';
import * as authApi from '@/api/endpoints/auth';
import { useAuthStore } from '@/stores/auth';

vi.mock('@/api/endpoints/auth', async (importOriginal) => ({
  ...(await importOriginal<typeof authApi>()),
  getSetupStatus: vi.fn(),
  logout: vi.fn(),
}));

const mockedSetupStatus = vi.mocked(authApi.getSetupStatus);
const mockedLogout = vi.mocked(authApi.logout);

function signIn(roles: string[]): void {
  useAuthStore.getState().setSession('tok', new Date(Date.now() + 60_000).toISOString(), {
    id: 'u1',
    username: 'alice',
    roles,
    disabled: false,
  });
}

// "/nowhere" matches no route, so only the shell renders and no page fetches data.
function renderApp(path = '/nowhere'): void {
  render(
    <TestProviders>
      <MemoryRouter initialEntries={[path]}>
        <App />
      </MemoryRouter>
    </TestProviders>,
  );
}

function primaryNav(): HTMLElement {
  return screen.getByRole('navigation', { name: 'primary' });
}

function navLinkNames(): string[] {
  return within(primaryNav())
    .getAllByRole('link')
    .map((link) => link.textContent ?? '');
}

describe('App shell', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.getState().clearSession();
    mockedSetupStatus.mockResolvedValue({
      setupRequired: false,
      completed: true,
      phase: 'complete',
    });
    mockedLogout.mockResolvedValue(undefined);
  });

  it('shows only public links and a sign-in link when signed out', async () => {
    renderApp();

    await waitFor(() => expect(navLinkNames()).toEqual(['Status', 'Settings', 'Sign in']));
    expect(screen.queryByRole('button', { name: 'Sign out' })).not.toBeInTheDocument();
    expect(screen.queryByRole('searchbox')).not.toBeInTheDocument();
  });

  it('redirects a signed-out visitor from a protected page to sign-in', async () => {
    renderApp('/');

    expect(await screen.findByText('Username')).toBeInTheDocument();
  });

  it('shows the library links but no admin links to a regular user', async () => {
    signIn(['user']);
    renderApp();

    await waitFor(() =>
      expect(navLinkNames()).toEqual([
        'Home',
        'Libraries',
        'Requests',
        'Status',
        'Settings',
        'Profile',
      ]),
    );
    expect(screen.getByRole('button', { name: 'Sign out' })).toBeInTheDocument();
  });

  it('adds user and request administration links for an admin', async () => {
    signIn(['admin']);
    renderApp();

    await waitFor(() => expect(navLinkNames()).toContain('Users'));
    expect(navLinkNames()).toContain('Approve requests');
  });

  it('hides navigation while onboarding is incomplete', async () => {
    mockedSetupStatus.mockResolvedValue({
      setupRequired: false,
      completed: false,
      phase: 'libraries',
    });
    signIn(['admin']);
    renderApp();

    await waitFor(() => expect(within(primaryNav()).queryAllByRole('link')).toHaveLength(0));
    expect(screen.queryByRole('button', { name: 'Sign out' })).not.toBeInTheDocument();
  });

  it('signs out, clears the session and returns to sign-in', async () => {
    signIn(['user']);
    renderApp();

    fireEvent.click(await screen.findByRole('button', { name: 'Sign out' }));

    await waitFor(() => expect(mockedLogout).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(useAuthStore.getState().accessToken).toBeNull());
    expect(await screen.findByText('Username')).toBeInTheDocument();
    expect(navLinkNames()).toContain('Sign in');
  });

  it('clears the session even when the logout request fails', async () => {
    mockedLogout.mockRejectedValue(new Error('network down'));
    signIn(['user']);
    renderApp();

    fireEvent.click(await screen.findByRole('button', { name: 'Sign out' }));

    await waitFor(() => expect(useAuthStore.getState().accessToken).toBeNull());
    expect(await screen.findByText('Username')).toBeInTheDocument();
  });
});
