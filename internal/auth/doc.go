// Implementation roadmap:
//
//   - Sprint 03 Paket: implement TokenService with HS256/RS256 JWT signing
//     and short-lived access tokens (ExpiresAt <= 15m).
//   - Sprint 03 Paket: implement RefreshTokenStore over the SQLite layer
//     introduced in Sprint 01 Paket 4. Hash secrets with argon2id; never
//     store the raw value.
//
// Until then the API gateway treats every request as anonymous; only health
// and version endpoints are exposed.
package auth
