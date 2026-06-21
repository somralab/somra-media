# Requests & Automation

## Content requests

Users discover titles and submit **requests**. Admins approve or reject from the admin queue. Approved requests can trigger automation when plugins are configured.

## Plugins (optional)

Somra works fully without plugins. To automate acquisition:

1. **Settings → Plugins** — enable stub or real indexer/download client plugins.
2. Configure indexer instances for search.
3. Configure download clients for handoff.

## Automation hub

Admins manage quality profiles, series monitors, and active downloads from **Automation**.

## Indexers

Search indexers from the automation UI. Results depend on configured plugin instances (torrent/usenet).

## Downloads

Download status refreshes via polling (5 s interval in 1.0). Completed imports trigger a library scan on the target path.
