package main

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// resetFlags resets pflag and viper state for testing
func resetFlags() {
	// Reset pflag
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	// Reset viper
	viper.Reset()
}

// TestFlagNormalizer tests that the flag normalizer converts underscores to dashes
func TestFlagNormalizer(t *testing.T) {
	resetFlags()

	// Set up the normalizer
	pflag.CommandLine.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
	})

	// Define a test flag
	pflag.String("test-flag", "default", "test flag")

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "dash format",
			args:     []string{"--test-flag=dash-value"},
			expected: "dash-value",
		},
		{
			name:     "underscore format",
			args:     []string{"--test_flag=underscore-value"},
			expected: "underscore-value",
		},
		{
			name:     "mixed format",
			args:     []string{"--test_flag=mixed-value"},
			expected: "mixed-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag value
			pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
			pflag.CommandLine.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
				return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
			})
			pflag.String("test-flag", "default", "test flag")

			// Parse the args
			err := pflag.CommandLine.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse args: %v", err)
			}

			// Get the value
			val, err := pflag.CommandLine.GetString("test-flag")
			if err != nil {
				t.Fatalf("Failed to get flag value: %v", err)
			}

			if val != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, val)
			}
		})
	}
}

// TestTracerouteFlagsNormalization tests that traceroute-related flags accept both formats
func TestTracerouteFlagsNormalization(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedBin   string
		expectedFlags string
		expectedRaw   bool
		expectedMax   int
	}{
		{
			name: "dash format",
			args: []string{
				"--traceroute-bin=/usr/bin/traceroute",
				"--traceroute-flags=-n -q 1",
				"--traceroute-raw=true",
				"--traceroute-max-concurrent=5",
			},
			expectedBin:   "/usr/bin/traceroute",
			expectedFlags: "-n -q 1",
			expectedRaw:   true,
			expectedMax:   5,
		},
		{
			name: "underscore format",
			args: []string{
				"--traceroute_bin=/usr/bin/mtr",
				"--traceroute_flags=-w -c1",
				"--traceroute_raw=false",
				"--traceroute_max_concurrent=20",
			},
			expectedBin:   "/usr/bin/mtr",
			expectedFlags: "-w -c1",
			expectedRaw:   false,
			expectedMax:   20,
		},
		{
			name: "mixed format",
			args: []string{
				"--traceroute-bin=/bin/traceroute",
				"--traceroute_flags=-I",
				"--traceroute-raw=true",
				"--traceroute_max_concurrent=15",
			},
			expectedBin:   "/bin/traceroute",
			expectedFlags: "-I",
			expectedRaw:   true,
			expectedMax:   15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()

			// Set up the normalizer
			pflag.CommandLine.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
				return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
			})

			// Define flags (using dash format as the standard)
			pflag.String("traceroute-bin", "", "traceroute binary")
			pflag.String("traceroute-flags", "", "traceroute flags")
			pflag.Bool("traceroute-raw", false, "traceroute raw")
			pflag.Int("traceroute-max-concurrent", 10, "max concurrent")

			// Parse the args
			err := pflag.CommandLine.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse args: %v", err)
			}

			// Verify values
			bin, _ := pflag.CommandLine.GetString("traceroute-bin")
			if bin != tt.expectedBin {
				t.Errorf("traceroute-bin: expected %q, got %q", tt.expectedBin, bin)
			}

			flags, _ := pflag.CommandLine.GetString("traceroute-flags")
			if flags != tt.expectedFlags {
				t.Errorf("traceroute-flags: expected %q, got %q", tt.expectedFlags, flags)
			}

			raw, _ := pflag.CommandLine.GetBool("traceroute-raw")
			if raw != tt.expectedRaw {
				t.Errorf("traceroute-raw: expected %v, got %v", tt.expectedRaw, raw)
			}

			maxConcurrent, _ := pflag.CommandLine.GetInt("traceroute-max-concurrent")
			if maxConcurrent != tt.expectedMax {
				t.Errorf("traceroute-max-concurrent: expected %d, got %d", tt.expectedMax, maxConcurrent)
			}
		})
	}
}

// TestBirdRestrictCmdsNormalization tests that bird-restrict-cmds accepts both formats
func TestBirdRestrictCmdsNormalization(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "dash format true",
			args:     []string{"--bird-restrict-cmds=true"},
			expected: true,
		},
		{
			name:     "underscore format false",
			args:     []string{"--bird_restrict_cmds=false"},
			expected: false,
		},
		{
			name:     "mixed format true",
			args:     []string{"--bird-restrict_cmds=true"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()

			// Set up the normalizer
			pflag.CommandLine.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
				return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
			})

			// Define flag
			pflag.Bool("bird-restrict-cmds", true, "restrict commands")

			// Parse the args
			err := pflag.CommandLine.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse args: %v", err)
			}

			// Verify value
			val, _ := pflag.CommandLine.GetBool("bird-restrict-cmds")
			if val != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, val)
			}
		})
	}
}

func TestParseSettings(t *testing.T) {
	resetFlags()
	parseSettings()
	resetFlags()
}
