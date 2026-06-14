import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { I18nextProvider } from 'react-i18next';
import i18n from '@/i18n';
import { WizardShell, StepIndicator } from '@/components/onboarding/WizardShell';
import { RecommendationCard } from '@/components/onboarding/RecommendationCard';

function wrap(ui: React.ReactNode): React.ReactElement {
  return <I18nextProvider i18n={i18n}>{ui}</I18nextProvider>;
}

describe('onboarding components', () => {
  it('renders wizard shell with step indicator', () => {
    render(wrap(<WizardShell currentStep="language">Step content</WizardShell>));
    expect(screen.getByText(/welcome to somra/i)).toBeInTheDocument();
    expect(screen.getByTestId('wizard-step-language')).toBeInTheDocument();
    expect(screen.getByText('Step content')).toBeInTheDocument();
  });

  it('renders step indicator for defaults phase', () => {
    render(wrap(<StepIndicator currentStep="defaults" />));
    expect(screen.getByTestId('wizard-step-defaults')).toHaveAttribute('aria-current', 'step');
  });

  it('renders recommendation card', () => {
    render(
      wrap(
        <RecommendationCard title="2 concurrent transcodes" description="Based on CPU" applied />,
      ),
    );
    expect(screen.getByTestId('recommendation-card')).toBeInTheDocument();
    expect(screen.getByText('2 concurrent transcodes')).toBeInTheDocument();
  });
});
