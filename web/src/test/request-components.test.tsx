import { type ReactElement } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '@/i18n';
import { QualitySelector } from '@/components/requests/QualitySelector';
import { RequestStatusBadge } from '@/components/requests/RequestStatusBadge';

function renderWithI18n(ui: ReactElement): ReturnType<typeof render> {
  return render(<I18nextProvider i18n={i18n}>{ui}</I18nextProvider>);
}

describe('request components', () => {
  it('renders status badge with localized label', async () => {
    await i18n.changeLanguage('en-US');
    renderWithI18n(<RequestStatusBadge status="pending" />);
    expect(screen.getByText('Pending')).toBeInTheDocument();
  });

  it('renders all quality options', async () => {
    await i18n.changeLanguage('en-US');
    const onChange = vi.fn();
    renderWithI18n(<QualitySelector value="1080p" onChange={onChange} />);
    expect(screen.getByText('1080p')).toBeInTheDocument();
    expect(screen.getByText('720p')).toBeInTheDocument();
    expect(screen.getByText('Any')).toBeInTheDocument();
  });

  it('calls onChange when a different quality is selected', async () => {
    await i18n.changeLanguage('en-US');
    const onChange = vi.fn();
    renderWithI18n(<QualitySelector value="1080p" onChange={onChange} />);
    fireEvent.click(screen.getByRole('radio', { name: '720p' }));
    expect(onChange).toHaveBeenCalledWith('720p');
  });
});
