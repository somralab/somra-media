package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPlugin struct {
	id      string
	typ     PluginType
	version string
}

func (m mockPlugin) ID() string              { return m.id }
func (m mockPlugin) Type() PluginType        { return m.typ }
func (m mockPlugin) ContractVersion() string { return m.version }

func TestValidateContract_success(t *testing.T) {
	t.Parallel()
	p := mockPlugin{id: "idx-1", typ: PluginTypeIndexer, version: ContractVersion}
	require.NoError(t, ValidateContract(p))
}

func TestValidateContract_incompatibleVersion(t *testing.T) {
	t.Parallel()
	p := mockPlugin{id: "idx-1", typ: PluginTypeIndexer, version: "0"}
	err := ValidateContract(p)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIncompatibleContract)
}

func TestValidateContract_nilPlugin(t *testing.T) {
	t.Parallel()
	err := ValidateContract(nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIncompatibleContract)
}

func TestContractVersionConstant(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "1", ContractVersion)
}

func TestPluginTypeConstants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, PluginType("indexer"), PluginTypeIndexer)
	assert.Equal(t, PluginType("download_client"), PluginTypeDownloadClient)
}
