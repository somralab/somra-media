package plugin

import "fmt"

// ValidateContract checks that p implements the current core contract version.
func ValidateContract(p Plugin) error {
	if p == nil {
		return fmt.Errorf("validate plugin contract: %w", ErrIncompatibleContract)
	}
	if p.ContractVersion() != ContractVersion {
		return fmt.Errorf("plugin %q contract %q: %w", p.ID(), p.ContractVersion(), ErrIncompatibleContract)
	}
	return nil
}
