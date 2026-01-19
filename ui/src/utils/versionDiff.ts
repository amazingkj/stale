export type VersionDiffType = 'major' | 'minor' | 'patch' | 'unknown';

/**
 * Parse a semver-like version string into parts
 */
function parseVersion(version: string): [number, number, number] | null {
  // Remove common prefixes like 'v', '^', '~', etc.
  const cleaned = version.replace(/^[v^~>=<]+/, '').trim();

  // Match semver pattern (with optional pre-release/build metadata)
  const match = cleaned.match(/^(\d+)(?:\.(\d+))?(?:\.(\d+))?/);
  if (!match) return null;

  return [
    parseInt(match[1], 10) || 0,
    parseInt(match[2], 10) || 0,
    parseInt(match[3], 10) || 0,
  ];
}

/**
 * Compare two versions and return the diff type
 */
export function getVersionDiff(current: string, latest: string): VersionDiffType {
  const currentParts = parseVersion(current);
  const latestParts = parseVersion(latest);

  if (!currentParts || !latestParts) {
    return 'unknown';
  }

  const [currMajor, currMinor, currPatch] = currentParts;
  const [latMajor, latMinor, latPatch] = latestParts;

  if (latMajor > currMajor) {
    return 'major';
  }
  if (latMinor > currMinor) {
    return 'minor';
  }
  if (latPatch > currPatch) {
    return 'patch';
  }

  return 'unknown';
}

/**
 * Get display info for version diff
 */
export function getVersionDiffInfo(type: VersionDiffType): { label: string; color: string; bgColor: string } {
  switch (type) {
    case 'major':
      return { label: 'Major', color: '#dc2626', bgColor: '#fef2f2' };
    case 'minor':
      return { label: 'Minor', color: '#d97706', bgColor: '#fffbeb' };
    case 'patch':
      return { label: 'Patch', color: '#059669', bgColor: '#ecfdf5' };
    default:
      return { label: '', color: '#6b7280', bgColor: '#f3f4f6' };
  }
}
