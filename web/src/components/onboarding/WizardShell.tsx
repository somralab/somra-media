import { type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/cn';

export type OnboardingStepId = 'language' | 'admin' | 'library' | 'defaults' | 'scan' | 'complete';

const STEP_ORDER: OnboardingStepId[] = [
  'language',
  'admin',
  'library',
  'defaults',
  'scan',
  'complete',
];

interface WizardShellProps {
  children: ReactNode;
  currentStep: OnboardingStepId;
}

export function WizardShell({ children, currentStep }: WizardShellProps): ReactNode {
  const { t } = useTranslation('onboarding');

  return (
    <div className="mx-auto flex min-h-[70vh] max-w-2xl flex-col gap-8 p-6">
      <header className="flex flex-col gap-1 text-center">
        <h1 className="text-2xl font-semibold">{t('wizard.title')}</h1>
        <p className="text-sm text-muted">{t('wizard.subtitle')}</p>
      </header>
      <StepIndicator currentStep={currentStep} />
      <div className="flex-1">{children}</div>
    </div>
  );
}

interface StepIndicatorProps {
  currentStep: OnboardingStepId;
}

export function StepIndicator({ currentStep }: StepIndicatorProps): ReactNode {
  const { t } = useTranslation('onboarding');
  const currentIdx = STEP_ORDER.indexOf(currentStep);

  return (
    <nav aria-label={t('wizard.title')} className="flex justify-center gap-2">
      {STEP_ORDER.filter((s) => s !== 'complete').map((step, idx) => {
        const isActive = step === currentStep;
        const isDone = idx < currentIdx;
        return (
          <div
            key={step}
            className={cn(
              'flex flex-col items-center gap-1 text-xs',
              isActive ? 'text-primary' : isDone ? 'text-text' : 'text-muted',
            )}
            data-testid={`wizard-step-${step}`}
            aria-current={isActive ? 'step' : undefined}
          >
            <span
              className={cn(
                'flex h-7 w-7 items-center justify-center rounded-full border text-xs font-medium',
                isActive && 'border-primary bg-primary/15',
                isDone && 'text-primary-foreground border-primary bg-primary',
              )}
            >
              {idx + 1}
            </span>
            <span className="hidden sm:block">{t(`steps.${step}`)}</span>
          </div>
        );
      })}
    </nav>
  );
}
