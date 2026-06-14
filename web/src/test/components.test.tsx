import { describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from '@/components/ui/Card';
import { LanguageSwitcher } from '@/components/LanguageSwitcher';
import { ThemeSwitcher } from '@/components/ThemeSwitcher';
import { useUIStore } from '@/stores/ui';
import { createQueryClient } from '@/lib/queryClient';
import { ThemeProvider } from '@/theme/ThemeProvider';
import i18n from '@/i18n';

describe('<Button />', () => {
  it('renders a button with type="button" by default', () => {
    render(<Button>Save</Button>);
    const el = screen.getByRole('button', { name: 'Save' });
    expect(el.getAttribute('type')).toBe('button');
  });

  it('applies size and variant classes', () => {
    const { rerender } = render(
      <Button variant="danger" size="lg">
        Delete
      </Button>,
    );
    expect(screen.getByRole('button')).toHaveClass('bg-danger');
    rerender(
      <Button variant="ghost" size="sm">
        Cancel
      </Button>,
    );
    expect(screen.getByRole('button')).toHaveClass('bg-transparent');
  });

  it('supports asChild composition', () => {
    render(
      <Button asChild>
        <a href="#somewhere">link</a>
      </Button>,
    );
    const anchor = screen.getByRole('link', { name: 'link' });
    expect(anchor.tagName).toBe('A');
  });

  it('passes a click handler through', () => {
    const onClick = vi.fn();
    render(<Button onClick={onClick}>Click</Button>);
    fireEvent.click(screen.getByRole('button'));
    expect(onClick).toHaveBeenCalled();
  });
});

describe('<Input />', () => {
  it('renders with type="text" by default', () => {
    render(<Input aria-label="name" />);
    const el = screen.getByLabelText('name');
    expect(el.getAttribute('type')).toBe('text');
  });

  it('marks the input invalid via aria-invalid when invalid', () => {
    render(<Input aria-label="email" invalid />);
    expect(screen.getByLabelText('email').getAttribute('aria-invalid')).toBe('true');
  });

  it('respects custom type and forwards extra props', () => {
    render(<Input aria-label="pw" type="password" placeholder="••" />);
    const el = screen.getByLabelText('pw');
    expect(el.getAttribute('type')).toBe('password');
    expect(el.getAttribute('placeholder')).toBe('••');
  });
});

describe('<Card />', () => {
  it('renders the composition with title and footer', () => {
    render(
      <Card data-testid="card">
        <CardHeader>
          <CardTitle>Title</CardTitle>
          <CardDescription>Description</CardDescription>
        </CardHeader>
        <CardContent>Body</CardContent>
        <CardFooter>Footer</CardFooter>
      </Card>,
    );
    expect(screen.getByText('Title')).toBeInTheDocument();
    expect(screen.getByText('Description')).toBeInTheDocument();
    expect(screen.getByText('Body')).toBeInTheDocument();
    expect(screen.getByText('Footer')).toBeInTheDocument();
  });
});

describe('<LanguageSwitcher />', () => {
  it('changes the active language when an option is selected', async () => {
    await i18n.changeLanguage('en-US');
    render(
      <I18nextProvider i18n={i18n}>
        <LanguageSwitcher />
      </I18nextProvider>,
    );
    const select = screen.getByRole('combobox') as HTMLSelectElement;
    expect(select.value).toBe('en-US');
    fireEvent.change(select, { target: { value: 'tr-TR' } });
    expect(i18n.language).toBe('tr-TR');
    await i18n.changeLanguage('en-US');
  });
});

describe('<ThemeSwitcher />', () => {
  it('renders the theme options', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <ThemeProvider>
          <ThemeSwitcher />
        </ThemeProvider>
      </I18nextProvider>,
    );
    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });
});

describe('useUIStore', () => {
  it('toggles the sidebar', () => {
    const before = useUIStore.getState().sidebarOpen;
    useUIStore.getState().toggleSidebar();
    expect(useUIStore.getState().sidebarOpen).toBe(!before);
    useUIStore.getState().setSidebarOpen(false);
    expect(useUIStore.getState().sidebarOpen).toBe(false);
  });
});

describe('createQueryClient', () => {
  it('returns a configured client', () => {
    const c = createQueryClient();
    const opts = c.getDefaultOptions();
    expect(opts.queries?.retry).toBe(2);
    expect(opts.mutations?.retry).toBe(0);
  });
});
