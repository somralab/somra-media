import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  advanceOnboardingStep,
  completeOnboarding,
  detectSystem,
  getOnboardingStatus,
  type OnboardingPhase,
} from '@/api/endpoints/settings';
import { getSetupStatus } from '@/api/endpoints/auth';

export function useSetupStatus() {
  return useQuery({
    queryKey: ['setup-status'],
    queryFn: getSetupStatus,
    staleTime: 30_000,
  });
}

export function useOnboardingStatus() {
  return useQuery({
    queryKey: ['onboarding-status'],
    queryFn: getOnboardingStatus,
    staleTime: 10_000,
  });
}

export function useSystemDetect(paths?: string) {
  return useQuery({
    queryKey: ['system-detect', paths],
    queryFn: () => detectSystem(paths),
    staleTime: 60_000,
  });
}

export function useAdvanceOnboarding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: advanceOnboardingStep,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['onboarding-status'] });
      void qc.invalidateQueries({ queryKey: ['setup-status'] });
    },
  });
}

export function useCompleteOnboarding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: completeOnboarding,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['onboarding-status'] });
      void qc.invalidateQueries({ queryKey: ['setup-status'] });
    },
  });
}

export function phaseIndex(phase: OnboardingPhase): number {
  const order: OnboardingPhase[] = ['language', 'admin', 'library', 'defaults', 'scan', 'complete'];
  return order.indexOf(phase);
}
