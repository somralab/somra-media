/**
 * Suggested default library root for the onboarding wizard.
 * Local Vite dev serves from repo root; production Docker images mount at /media.
 */
export function defaultLibraryPath(): string {
  return import.meta.env.DEV ? './deploy/media' : '/media';
}
