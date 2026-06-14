package diagnostics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/platform/diagnostics"
)

type fakeProvider struct {
	name     string
	status   diagnostics.Status
	detail   string
	critical bool
	calls    int
}

func (f *fakeProvider) Name() string   { return f.name }
func (f *fakeProvider) Critical() bool { return f.critical }
func (f *fakeProvider) Check(_ context.Context) diagnostics.Check {
	f.calls++
	return diagnostics.Check{Name: f.name, Status: f.status, Detail: f.detail}
}

func TestRegistryAggregatesStatuses(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		critical bool
		statuses []diagnostics.Status
		critFlag []bool
		want     diagnostics.Status
	}{
		{"all ok", false, []diagnostics.Status{diagnostics.StatusOK, diagnostics.StatusOK}, []bool{true, false}, diagnostics.StatusOK},
		{"degraded non-critical", false, []diagnostics.Status{diagnostics.StatusOK, diagnostics.StatusDegraded}, []bool{true, false}, diagnostics.StatusDegraded},
		{"critical down", false, []diagnostics.Status{diagnostics.StatusOK, diagnostics.StatusDown}, []bool{false, true}, diagnostics.StatusDown},
		{"non-critical down", false, []diagnostics.Status{diagnostics.StatusOK, diagnostics.StatusDown}, []bool{true, false}, diagnostics.StatusDegraded},
		{"critical degraded", false, []diagnostics.Status{diagnostics.StatusDegraded}, []bool{true}, diagnostics.StatusDegraded},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			reg := diagnostics.NewRegistry()
			for i, s := range tc.statuses {
				reg.Register(&fakeProvider{
					name:     "p" + string(rune('a'+i)),
					status:   s,
					critical: tc.critFlag[i],
				})
			}
			snap := reg.Run(context.Background())
			require.Equal(t, tc.want, snap.Overall)
			require.Len(t, snap.Checks, len(tc.statuses))
		})
	}
}

func TestRegistryEmpty(t *testing.T) {
	t.Parallel()
	reg := diagnostics.NewRegistry()
	snap := reg.Run(context.Background())
	require.Equal(t, diagnostics.StatusOK, snap.Overall)
	require.Empty(t, snap.Checks)
}

func TestRegistryRegisterDeduplicates(t *testing.T) {
	t.Parallel()
	reg := diagnostics.NewRegistry()
	p := &fakeProvider{name: "x", status: diagnostics.StatusOK}
	reg.Register(p)
	reg.Register(p)
	reg.Register(nil)
	snap := reg.Run(context.Background())
	require.Len(t, snap.Checks, 1)
}

func TestUptimeProvider(t *testing.T) {
	t.Parallel()
	p := diagnostics.NewUptimeProvider()
	time.Sleep(10 * time.Millisecond)
	check := p.Check(context.Background())
	require.Equal(t, diagnostics.StatusOK, check.Status)
	require.Equal(t, "uptime", check.Name)
	require.False(t, p.StartedAt().IsZero())
	require.False(t, p.Critical())
}

func TestSchedulerProviderReflectsState(t *testing.T) {
	t.Parallel()
	s := jobs.New(nil)
	p := diagnostics.NewSchedulerProvider(s)

	check := p.Check(context.Background())
	require.Equal(t, diagnostics.StatusDegraded, check.Status)
	require.Equal(t, "scheduler", p.Name())
	require.False(t, p.Critical())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)
	defer func() { _ = s.Stop(context.Background()) }()

	check = p.Check(context.Background())
	require.Equal(t, diagnostics.StatusOK, check.Status)
}

func TestSchedulerProviderNil(t *testing.T) {
	t.Parallel()
	p := diagnostics.NewSchedulerProvider(nil)
	check := p.Check(context.Background())
	require.Equal(t, diagnostics.StatusDown, check.Status)
}

type fakePinger struct{ err error }

func (f *fakePinger) Ping(_ context.Context) error { return f.err }

func TestDBProvider(t *testing.T) {
	t.Parallel()
	p := diagnostics.NewDBProvider(&fakePinger{})
	require.True(t, p.Critical())
	require.Equal(t, "database", p.Name())
	check := p.Check(context.Background())
	require.Equal(t, diagnostics.StatusOK, check.Status)

	p = diagnostics.NewDBProvider(&fakePinger{err: errors.New("nope")})
	check = p.Check(context.Background())
	require.Equal(t, diagnostics.StatusDown, check.Status)
	require.Contains(t, check.Detail, "nope")

	p = diagnostics.NewDBProvider(nil)
	check = p.Check(context.Background())
	require.Equal(t, diagnostics.StatusDown, check.Status)
}
