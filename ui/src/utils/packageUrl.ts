/**
 * Get the URL for a package on its respective registry
 */
export function getPackageUrl(name: string, ecosystem: string): string {
  switch (ecosystem) {
    case 'npm':
      return `https://www.npmjs.com/package/${name}`;
    case 'maven':
    case 'gradle': {
      const parts = name.split(':');
      if (parts.length === 2) {
        return `https://mvnrepository.com/artifact/${parts[0]}/${parts[1]}`;
      }
      return `https://mvnrepository.com/search?q=${encodeURIComponent(name)}`;
    }
    case 'go':
      return `https://pkg.go.dev/${name}`;
    default:
      return '#';
  }
}
