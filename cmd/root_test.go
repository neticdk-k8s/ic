package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/stretchr/testify/assert"
)

func Test_Wiring(t *testing.T) {
	testcases := []string{
		"get",
		"get cluster",
		"get clusters",
		"get regions",
		"get partitions",
		"create cluster",
		"delete cluster",
		"login",
		"logout",
	}

	for _, tc := range testcases {
		t.Run(tc, func(t *testing.T) {
			osargs := strings.Split(tc, " ")
			ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
			ac := ic.NewContext()
			ac.EC = ec
			cmd := newRootCmd(ac)
			cmd, _, err := cmd.Find(osargs)
			assert.NoError(t, err)
			assert.Equal(t, osargs[len(osargs)-1], cmd.Name())
		})
	}
}

func Test_UnknownCommand(t *testing.T) {
	ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
	ac := ic.NewContext()
	ac.EC = ec
	cmd := newRootCmd(ac)
	cmd.SetArgs([]string{"unknown"})
	err := cmd.ExecuteContext(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

func Test_HelpCommand(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		got := new(bytes.Buffer)
		ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
		ec.Stderr = got
		ec.Stdout = got
		ac := ic.NewContext()
		ac.EC = ec
		cmd := newRootCmd(ac)
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), "Usage:")
	})

	t.Run("help", func(t *testing.T) {
		got := new(bytes.Buffer)
		ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
		ec.Stderr = got
		ec.Stdout = got
		ac := ic.NewContext()
		ac.EC = ec
		cmd := newRootCmd(ac)
		cmd.SetArgs([]string{"help"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), "Usage:")
	})

	t.Run("--help", func(t *testing.T) {
		got := new(bytes.Buffer)
		ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
		ec.Stderr = got
		ec.Stdout = got
		ac := ic.NewContext()
		ac.EC = ec
		cmd := newRootCmd(ac)
		cmd.SetArgs([]string{"--help"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), "Usage:")
	})
}

func Test_VersionFlag(t *testing.T) {
	got := new(bytes.Buffer)
	ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
	ec.Stderr = got
	ec.Stdout = got
	ac := ic.NewContext()
	ac.EC = ec
	cmd := newRootCmd(ac)
	cmd.SetArgs([]string{"--version"})
	err := cmd.ExecuteContext(context.Background())
	assert.NoError(t, err)
	assert.Contains(t, got.String(), "ic version test")
}

func Test_OIDCFlags(t *testing.T) {
	t.Run("--oidc-token-cache-dir", func(t *testing.T) {
		got := new(bytes.Buffer)
		ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
		ec.Stderr = got
		ec.Stdout = got
		ac := ic.NewContext()
		ac.EC = ec
		cmd := newRootCmd(ac)
		cmd.SetArgs([]string{"--oidc-token-cache-dir", "/tmp"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, ac.OIDC.TokenCacheDir, "/tmp")
		assert.NotNil(t, ac.TokenCache)
	})
}
