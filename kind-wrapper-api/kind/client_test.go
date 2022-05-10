package kind

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParsingGetClustersOutput(t *testing.T) {
	emptyOutput := parseClusterNamesFromCommandOutput([]string{})
	require.Len(t, emptyOutput, 0)

	output := parseClusterNamesFromCommandOutput([]string{"kind"})
	require.Len(t, output, 1)
	_, ok := output["kind"]
	require.True(t, ok)

	output = parseClusterNamesFromCommandOutput([]string{"kind", "kind-2", "kind-3"})
	require.Len(t, output, 3)
	_, ok = output["kind"]
	require.True(t, ok)
	_, ok = output["kind-2"]
	require.True(t, ok)
	_, ok = output["kind-3"]
	require.True(t, ok)
}