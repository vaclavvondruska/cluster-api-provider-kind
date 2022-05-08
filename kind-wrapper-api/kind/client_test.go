package kind

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"testing"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") != "1" {
		return
	}
	isFailureCase := os.Getenv("FAILURE_CASE") == "1"
	code := 0
	switch os.Getenv("TEST_CASE") {
	case "create":
		if isFailureCase {
			_, _ = fmt.Fprintf(os.Stderr, "Cluster falied")
			code = 1
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "Cluster ready")
		}
	case "delete":
		if isFailureCase {
			_, _ = fmt.Fprintf(os.Stderr, "Cluster deletion failed")
			code = 1
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "Cluster deleted")
		}
	case "get":
		if isFailureCase {
			_, _ = fmt.Fprintf(os.Stderr, "Get clusters falied")
			code = 1
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "kind\nkind-2\nkind-3")
		}
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Command not found")
		code = 127
	}
	os.Exit(code)
}

func fakeExecCommandSuccess(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_TEST_PROCESS=1"}
	if command == "bash" {
		cmd.Env = append(cmd.Env, "TEST_CASE=create")
	} else if command == "kind" && len(args) > 0 && args[0] == "delete" {
		cmd.Env = append(cmd.Env, "TEST_CASE=delete")
	} else if command == "kind" && len(args) > 0 && args[0] == "get" {
		cmd.Env = append(cmd.Env, "TEST_CASE=get")
	}
	return cmd
}

func fakeExecCommandFailure(command string, args ...string) *exec.Cmd {
	cmd := fakeExecCommandSuccess(command, args...)
	cmd.Env = append(cmd.Env, "FAILURE_CASE=1")
	return cmd
}

func TestClusterCreate(t *testing.T) {
	kindClient := &CLIClient{cmdContext: fakeExecCommandSuccess}
	require.Error(t, kindClient.CreateCluster(""))
	require.NoError(t, kindClient.CreateCluster("---"))

	kindClient = &CLIClient{cmdContext: fakeExecCommandFailure}
	require.Error(t, kindClient.CreateCluster("---"))
}

func TestClusterDelete(t *testing.T) {
	kindClient := &CLIClient{cmdContext: fakeExecCommandSuccess}
	require.Error(t, kindClient.DeleteCluster(""))
	require.NoError(t, kindClient.CreateCluster("cluster"))

	kindClient = &CLIClient{cmdContext: fakeExecCommandFailure}
	require.Error(t, kindClient.CreateCluster("cluster"))
}

func TestGetClusters(t *testing.T) {
	kindClient := &CLIClient{cmdContext: fakeExecCommandSuccess}
	clusterNames, err := kindClient.GetClusters()
	require.NoError(t, err)
	require.Len(t, clusterNames, 3)

	kindClient = &CLIClient{cmdContext: fakeExecCommandFailure}
	_, err = kindClient.GetClusters()
	require.Error(t, err)
}

func TestParsingGetClustersOutput(t *testing.T) {
	emptyOutput := parseClusterNamesFromCommandOutput([]byte{})
	require.Len(t, emptyOutput, 0)

	output := parseClusterNamesFromCommandOutput([]byte("kind"))
	require.Len(t, output, 1)
	_, ok := output["kind"]
	require.True(t, ok)

	output = parseClusterNamesFromCommandOutput([]byte("kind\nkind-2\nkind-3"))
	require.Len(t, output, 3)
	_, ok = output["kind"]
	require.True(t, ok)
	_, ok = output["kind-2"]
	require.True(t, ok)
	_, ok = output["kind-3"]
	require.True(t, ok)
}