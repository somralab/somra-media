import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { I18nextProvider } from 'react-i18next';
import NotificationSettingsPage from '@/pages/NotificationSettingsPage';
import i18n from '@/i18n';
import { TestProviders } from './testUtils';

const patchMutate = vi.fn();

vi.mock('@/api/hooks/useNotificationPrefs', () => ({
  useNotificationPrefs: () => ({
    data: {
      preferences: [
        {
          id: 1,
          userId: 'u1',
          eventType: 'request.approved',
          channelId: 2,
          enabled: true,
          debounceSeconds: 60,
        },
      ],
    },
    isLoading: false,
    isError: false,
  }),
  usePatchNotificationPrefs: () => ({
    mutate: patchMutate,
    isPending: false,
    isSuccess: false,
  }),
}));

function renderPage(): ReturnType<typeof render> {
  return render(
    <TestProviders>
      <MemoryRouter>
        <I18nextProvider i18n={i18n}>
          <NotificationSettingsPage />
        </I18nextProvider>
      </MemoryRouter>
    </TestProviders>,
  );
}

describe('<NotificationSettingsPage />', () => {
  it('renders preferences and toggles enabled state', async () => {
    await i18n.changeLanguage('en-US');
    renderPage();
    expect(screen.getByRole('heading', { name: /notification preferences/i })).toBeInTheDocument();
    expect(screen.getByText(/request approved/i)).toBeInTheDocument();

    const checkbox = screen.getByRole('checkbox');
    fireEvent.click(checkbox);
    expect(patchMutate).toHaveBeenCalledWith({
      preferences: [
        expect.objectContaining({
          eventType: 'request.approved',
          channelId: 2,
          enabled: false,
        }),
      ],
    });
  });

  it('updates debounce seconds on blur', async () => {
    await i18n.changeLanguage('en-US');
    renderPage();
    const debounceInput = screen.getByRole('spinbutton');
    fireEvent.change(debounceInput, { target: { value: '30' } });
    fireEvent.blur(debounceInput);
    expect(patchMutate).toHaveBeenCalledWith({
      preferences: [
        expect.objectContaining({
          debounceSeconds: 30,
        }),
      ],
    });
  });
});
