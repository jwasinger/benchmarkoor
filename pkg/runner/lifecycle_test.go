package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestApplyArgsOverride verifies the shared helper used to fold
// extra_args and benchmark_extra_args into a base command. The key
// behaviors are:
//
//   - empty / nil overrides leave cmd untouched
//   - args without "=" are appended without any eviction
//   - args with "--flag=value" evict any base entry that has the
//     same "--flag=" prefix, so later sources win without producing
//     duplicate flags
//   - arg ordering within args is preserved on append
func TestApplyArgsOverride(t *testing.T) {
	tests := []struct {
		name string
		cmd  []string
		args []string
		want []string
	}{
		{
			name: "nil args returns cmd unchanged",
			cmd:  []string{"geth", "--verbosity=3"},
			args: nil,
			want: []string{"geth", "--verbosity=3"},
		},
		{
			name: "empty args returns cmd unchanged",
			cmd:  []string{"geth", "--verbosity=3"},
			args: []string{},
			want: []string{"geth", "--verbosity=3"},
		},
		{
			name: "args without = are appended verbatim",
			cmd:  []string{"geth", "--verbosity=3"},
			args: []string{"--debug", "--pprof"},
			want: []string{"geth", "--verbosity=3", "--debug", "--pprof"},
		},
		{
			name: "args with = override matching base flags",
			cmd:  []string{"geth", "--verbosity=3", "--http"},
			args: []string{"--verbosity=5"},
			want: []string{"geth", "--http", "--verbosity=5"},
		},
		{
			name: "args with = leave non-matching base flags alone",
			cmd:  []string{"geth", "--verbosity=3", "--datadir=/data"},
			args: []string{"--miner.gaslimit=1000000000"},
			want: []string{"geth", "--verbosity=3", "--datadir=/data", "--miner.gaslimit=1000000000"},
		},
		{
			name: "mix of overriding and appending args",
			cmd:  []string{"geth", "--verbosity=3", "--http"},
			args: []string{"--verbosity=5", "--pprof"},
			want: []string{"geth", "--http", "--verbosity=5", "--pprof"},
		},
		{
			name: "override on a cmd that has no matching prefix is just an append",
			cmd:  []string{"geth"},
			args: []string{"--profile=cpu"},
			want: []string{"geth", "--profile=cpu"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyArgsOverride(tt.cmd, tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestApplyArgsOverride_BenchmarkLayeredOverExtra verifies the layering used
// by runContainerLifecycle: the base command is first composed with
// extra_args, then benchmark_extra_args is applied on top. The latter must
// win when both target the same flag, since benchmark_extra_args is intended
// to be the final, benchmark-phase-only word on client flags.
func TestApplyArgsOverride_BenchmarkLayeredOverExtra(t *testing.T) {
	base := []string{"geth", "--verbosity=2", "--http"}
	extra := []string{"--verbosity=4", "--datadir=/data"}
	benchmarkOnly := []string{"--verbosity=5", "--pprof"}

	withExtra := applyArgsOverride(base, extra)
	final := applyArgsOverride(withExtra, benchmarkOnly)

	// --verbosity from benchmarkOnly must win; --http and --datadir survive;
	// --pprof is a bare flag and is appended.
	assert.Equal(t,
		[]string{"geth", "--http", "--datadir=/data", "--verbosity=5", "--pprof"},
		final,
	)
}
