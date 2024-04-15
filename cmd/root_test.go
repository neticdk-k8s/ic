package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Wiring(t *testing.T) {
	testcases := []string{
		"get",
		"get cluster",
		"get clusters",
		"login",
		"logout",
	}

	for _, tc := range testcases {
		t.Run(tc, func(t *testing.T) {
			osargs := strings.Split(tc, " ")
			cmd := NewRootCmd(NewExecutionContext(ExecutionContextInput{}))
			cmd, _, err := cmd.Find(osargs)
			assert.NoError(t, err)
			assert.Equal(t, osargs[len(osargs)-1], cmd.Name())
		})
	}
}

func Test_UnknownCommand(t *testing.T) {
	in := ExecutionContextInput{
		Version: "testing",
	}
	ec := NewExecutionContext(in)
	cmd := NewRootCmd(ec)
	cmd.SetArgs([]string{"unknown"})
	err := cmd.ExecuteContext(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

func Test_HelpCommand(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		got := new(bytes.Buffer)
		in := ExecutionContextInput{
			Stdout:  got,
			Stderr:  got,
			Version: "testing",
		}
		ec := NewExecutionContext(in)
		cmd := NewRootCmd(ec)
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), "Usage:")
	})

	t.Run("help", func(t *testing.T) {
		got := new(bytes.Buffer)
		in := ExecutionContextInput{
			Stdout:  got,
			Stderr:  got,
			Version: "testing",
		}
		ec := NewExecutionContext(in)
		cmd := NewRootCmd(ec)
		cmd.SetArgs([]string{"help"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), "Usage:")
	})

	t.Run("--help", func(t *testing.T) {
		got := new(bytes.Buffer)
		in := ExecutionContextInput{
			Stdout:  got,
			Stderr:  got,
			Version: "testing",
		}
		ec := NewExecutionContext(in)
		cmd := NewRootCmd(ec)
		cmd.SetArgs([]string{"--help"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), "Usage:")
	})
}

func Test_VersionFlag(t *testing.T) {
	got := new(bytes.Buffer)
	in := ExecutionContextInput{
		Stdout:  got,
		Stderr:  got,
		Version: "testing",
	}
	ec := NewExecutionContext(in)
	cmd := NewRootCmd(ec)
	cmd.SetArgs([]string{"--version"})
	err := cmd.ExecuteContext(context.Background())
	assert.NoError(t, err)
	assert.Contains(t, got.String(), "ic version testing")
}

func Test_OIDCFlags(t *testing.T) {
	t.Run("--oidc-token-cache-dir", func(t *testing.T) {
		got := new(bytes.Buffer)
		in := ExecutionContextInput{
			Stdout:  got,
			Stderr:  got,
			Version: "testing",
		}
		ec := NewExecutionContext(in)
		cmd := NewRootCmd(ec)
		cmd.SetArgs([]string{"--oidc-token-cache-dir", "/tmp"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, ec.OIDC.TokenCacheDir, "/tmp")
		assert.NotNil(t, ec.TokenCache)
	})
}
