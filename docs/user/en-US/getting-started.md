# Getting Started

1. **Install** Somra — see [installation.md](./installation.md).
2. **Open the web UI** at `http://your-host:8080`.
3. **Setup wizard:** choose language (en-US / tr-TR), create the admin account, add your first library path.
4. **Scan:** Somra indexes media and fetches metadata automatically.
5. **Browse** the home page and open a title to play or request missing content.

## First library

1. Go to **Libraries** → **Add library**.
2. Choose type: Movies, TV, or Music.
3. Add the folder path where files are stored (must be readable by the container).
4. Trigger **Scan** — progress appears in the UI and via SSE events.

## Profiles and users

- Admin can create users under **Settings → Users**.
- Each user has profiles with separate watch history and locale preferences.

## Locale

Switch language from the header dropdown. User docs are available in English and Turkish under `docs/user/`.

## Next steps

- [Library guide](./library.md)
- [Playback](./playback.md)
- [Requests & automation](./requests-automation.md)
- [FAQ](./faq-troubleshooting.md)
