import { describe, expect, it } from 'vitest';

import { defaultLibraryPath } from '@/lib/libraryPaths';

describe('defaultLibraryPath', () => {
  it('returns deploy media path in dev', () => {
    expect(defaultLibraryPath()).toBe('./deploy/media');
  });
});
