package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/somralab/somra-media/internal/platform/db"
)

// OnboardingState is the wizard status returned to clients.
type OnboardingState struct {
	Phase         string         `json:"phase"`
	Completed     bool           `json:"completed"`
	SetupRequired bool           `json:"setupRequired"`
	SmartDefaults *SmartDefaults `json:"smartDefaults,omitempty"`
}

// StepRequest carries wizard step transitions.
type StepRequest struct {
	Phase         string
	Locale        string
	ApplyDefaults bool
	LibraryID     int64
}

// SetupChecker reports whether first-admin setup is still required.
type SetupChecker interface {
	SetupRequired(ctx context.Context) (bool, error)
}

// Onboarding manages the first-run wizard state machine.
type Onboarding struct {
	settings *Service
	repo     *db.SettingsRepo
	setup    SetupChecker
}

// NewOnboarding returns an onboarding service.
func NewOnboarding(repo *db.SettingsRepo, settings *Service, setup SetupChecker) *Onboarding {
	return &Onboarding{repo: repo, settings: settings, setup: setup}
}

// Status returns the current onboarding state.
func (o *Onboarding) Status(ctx context.Context) (OnboardingState, error) {
	setupRequired, err := o.setup.SetupRequired(ctx)
	if err != nil {
		return OnboardingState{}, err
	}
	completed, err := o.isCompleted(ctx)
	if err != nil {
		return OnboardingState{}, err
	}
	phase, err := o.currentPhase(ctx, setupRequired, completed)
	if err != nil {
		return OnboardingState{}, err
	}
	state := OnboardingState{
		Phase:         phase,
		Completed:     completed,
		SetupRequired: setupRequired,
	}
	if phase == PhaseDefaults {
		locale, _ := o.settings.GetString(ctx, KeyDefaultLocale, "en-US")
		profile := DetectSystem(nil)
		def := RecommendDefaultsWithProfile(detectCPUCores(), locale, profile)
		state.SmartDefaults = &def
	}
	return state, nil
}

// AdvanceStep processes a wizard step transition.
func (o *Onboarding) AdvanceStep(ctx context.Context, req StepRequest) (OnboardingState, error) {
	switch req.Phase {
	case PhaseLanguage:
		if req.Locale != "en-US" && req.Locale != "tr-TR" {
			return OnboardingState{}, fmt.Errorf("onboarding: invalid locale %q", req.Locale)
		}
		if err := o.repo.Set(ctx, KeyDefaultLocale, req.Locale); err != nil {
			return OnboardingState{}, err
		}
		if err := o.repo.Set(ctx, KeyOnboardingPhase, PhaseAdmin); err != nil {
			return OnboardingState{}, err
		}
	case PhaseDefaults:
		if req.ApplyDefaults {
			locale, _ := o.settings.GetString(ctx, KeyDefaultLocale, "en-US")
			profile := DetectSystem(nil)
			def := RecommendDefaultsWithProfile(detectCPUCores(), locale, profile)
			if err := o.settings.ApplySmartDefaults(ctx, def); err != nil {
				return OnboardingState{}, err
			}
		}
		if err := o.repo.Set(ctx, KeyOnboardingPhase, PhaseScan); err != nil {
			return OnboardingState{}, err
		}
	case PhaseScan:
		if err := o.repo.Set(ctx, KeyOnboardingPhase, PhaseComplete); err != nil {
			return OnboardingState{}, err
		}
	default:
		return OnboardingState{}, fmt.Errorf("onboarding: unsupported step %q", req.Phase)
	}
	return o.Status(ctx)
}

// Complete marks onboarding as finished.
func (o *Onboarding) Complete(ctx context.Context) error {
	if err := o.repo.Set(ctx, KeyOnboardingCompleted, "true"); err != nil {
		return err
	}
	return o.repo.Set(ctx, KeyOnboardingPhase, PhaseComplete)
}

// AfterAdminCreated advances phase from admin to library after setup/admin.
func (o *Onboarding) AfterAdminCreated(ctx context.Context) error {
	completed, err := o.isCompleted(ctx)
	if err != nil || completed {
		return err
	}
	phase, _ := o.repo.Get(ctx, KeyOnboardingPhase)
	if phase == "" || phase == PhaseLanguage || phase == PhaseAdmin {
		return o.repo.Set(ctx, KeyOnboardingPhase, PhaseLibrary)
	}
	return nil
}

// AfterLibraryCreated advances phase from library to defaults.
func (o *Onboarding) AfterLibraryCreated(ctx context.Context) error {
	completed, err := o.isCompleted(ctx)
	if err != nil || completed {
		return err
	}
	return o.repo.Set(ctx, KeyOnboardingPhase, PhaseDefaults)
}

func (o *Onboarding) isCompleted(ctx context.Context) (bool, error) {
	v, err := o.repo.Get(ctx, KeyOnboardingCompleted)
	if errors.Is(err, db.ErrSettingNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return v == "true" || v == "1", nil
}

func (o *Onboarding) currentPhase(ctx context.Context, setupRequired, completed bool) (string, error) {
	if completed {
		return PhaseComplete, nil
	}
	phase, err := o.repo.Get(ctx, KeyOnboardingPhase)
	if errors.Is(err, db.ErrSettingNotFound) {
		phase = PhaseLanguage
	} else if err != nil {
		return "", err
	}
	if phase == "" {
		phase = PhaseLanguage
	}
	if setupRequired && phase != PhaseLanguage && phase != PhaseAdmin {
		// Admin not created yet — stay on admin step.
		return PhaseAdmin, nil
	}
	if !setupRequired && phase == PhaseLanguage {
		return PhaseLibrary, nil
	}
	if !setupRequired && phase == PhaseAdmin {
		return PhaseLibrary, nil
	}
	return phase, nil
}

func detectCPUCores() int {
	return DetectSystem(nil).CPUCores
}
