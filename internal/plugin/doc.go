// Package plugin defines versioned contracts for content acquisition adapters
// (indexers and download clients). The Somra core depends only on these
// interfaces; concrete Torznab, Newznab, qBittorrent, and similar adapters
// implement them in isolated packages so acquisition capabilities remain
// optional and legally separable from the core product.
//
// Contract versioning: bump [ContractVersion] when breaking interface or DTO
// changes require adapters to recompile. The registry (Sprint 09 A2) should
// call [ValidateContract] before accepting a plugin.
package plugin
