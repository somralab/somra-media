import { describe, expect, it, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import LoginPage from '@/pages/LoginPage';
import ProfilePage from '@/pages/ProfilePage';
import AdminUsersPage from '@/pages/AdminUsersPage';
import { ThemeProvider } from '@/theme/ThemeProvider';
import { TestProviders } from './testUtils';
import * as authApi from '@/api/endpoints/auth';

vi.mock('@/api/endpoints/auth', () => ({
  getSetupStatus: vi.fn(),
  setupAdmin: vi.fn(),
  login: vi.fn(),
  updateProfile: vi.fn(),
  getProfile: vi.fn(),
  listUsers: vi.fn(),
  createUser: vi.fn(),
  updateUser: vi.fn(),
}));

const mockedAuth = vi.mocked(authApi);

describe('auth pages', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders login form when setup is complete', async () => {
    mockedAuth.getSetupStatus.mockResolvedValue({ setupRequired: false });

    render(
      <TestProviders>
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      </TestProviders>,
    );

    expect(await screen.findByRole('heading', { name: /sign in/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
  });

  it('renders setup form when setup is required', async () => {
    mockedAuth.getSetupStatus.mockResolvedValue({ setupRequired: true });

    render(
      <TestProviders>
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      </TestProviders>,
    );

    expect(await screen.findByRole('heading', { name: /create admin account/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /create admin/i })).toBeInTheDocument();
  });

  it('submits login credentials', async () => {
    mockedAuth.getSetupStatus.mockResolvedValue({ setupRequired: false });
    mockedAuth.login.mockResolvedValue({
      accessToken: 'tok',
      expiresAt: new Date().toISOString(),
      user: { id: '1', username: 'admin', roles: ['admin'], disabled: false },
    });

    render(
      <TestProviders>
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      </TestProviders>,
    );

    await screen.findByRole('heading', { name: /sign in/i });
    fireEvent.change(screen.getByLabelText(/username/i), { target: { value: 'admin' } });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'AdminPass1' } });
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(mockedAuth.login).toHaveBeenCalledWith('admin', 'AdminPass1');
    });
  });

  it('renders profile preferences', async () => {
    mockedAuth.getProfile.mockResolvedValue({
      userId: 'user-1',
      locale: 'en-US',
      theme: 'cinematic',
      isChild: false,
    });

    render(
      <TestProviders>
        <ThemeProvider>
          <ProfilePage />
        </ThemeProvider>
      </TestProviders>,
    );

    expect(await screen.findByRole('heading', { name: /profile/i })).toBeInTheDocument();
    expect(screen.getByText('user-1')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /save preferences/i })).toBeInTheDocument();
  });

  it('saves profile preferences', async () => {
    mockedAuth.getProfile.mockResolvedValue({
      userId: 'user-1',
      locale: 'en-US',
      theme: 'cinematic',
      maxContentRating: null,
      isChild: false,
    });
    mockedAuth.updateProfile.mockResolvedValue({
      userId: 'user-1',
      locale: 'en-US',
      theme: 'cinematic',
      maxContentRating: null,
      isChild: false,
    });

    render(
      <TestProviders>
        <ThemeProvider>
          <ProfilePage />
        </ThemeProvider>
      </TestProviders>,
    );

    await screen.findByRole('heading', { name: /profile/i });
    fireEvent.click(screen.getByRole('button', { name: /save preferences/i }));

    await waitFor(() => {
      expect(mockedAuth.updateProfile).toHaveBeenCalled();
    });
  });

  it('renders admin users table and create form', async () => {
    mockedAuth.listUsers.mockResolvedValue([
      { id: '1', username: 'admin', roles: ['admin'], disabled: false },
      { id: '2', username: 'member', roles: ['user'], disabled: true },
    ]);

    render(
      <TestProviders>
        <AdminUsersPage />
      </TestProviders>,
    );

    expect(await screen.findByRole('heading', { name: /user management/i })).toBeInTheDocument();
    expect(screen.getAllByText('admin').length).toBeGreaterThan(0);
    expect(screen.getByText('member')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^create$/i })).toBeInTheDocument();
  });

  it('toggles user disabled state', async () => {
    mockedAuth.listUsers.mockResolvedValue([
      { id: '2', username: 'member', roles: ['user'], disabled: false },
    ]);
    mockedAuth.updateUser.mockResolvedValue({
      id: '2',
      username: 'member',
      roles: ['user'],
      disabled: true,
    });

    render(
      <TestProviders>
        <AdminUsersPage />
      </TestProviders>,
    );

    await screen.findByText('member');
    fireEvent.click(screen.getByRole('button', { name: /disable/i }));

    await waitFor(() => {
      expect(mockedAuth.updateUser).toHaveBeenCalledWith('2', { disabled: true });
    });
  });
});
