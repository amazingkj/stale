interface Props {
  current: string;
  latest: string;
  isOutdated: boolean;
}

export function VersionBadge({ current, latest, isOutdated }: Props) {
  if (!isOutdated) {
    return (
      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
        {current}
      </span>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
        {current}
      </span>
      <span className="text-gray-400">→</span>
      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
        {latest}
      </span>
    </div>
  );
}
