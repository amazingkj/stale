interface Props {
  title: string;
  value: number;
  subtitle?: string;
  color: 'blue' | 'red' | 'green' | 'purple';
}

const colorClasses = {
  blue: 'bg-blue-50 text-blue-700',
  red: 'bg-red-50 text-red-700',
  green: 'bg-green-50 text-green-700',
  purple: 'bg-purple-50 text-purple-700',
};

export function StatCard({ title, value, subtitle, color }: Props) {
  return (
    <div className={`rounded-lg p-6 ${colorClasses[color]}`}>
      <p className="text-sm font-medium opacity-75">{title}</p>
      <p className="text-3xl font-bold mt-2">{value.toLocaleString()}</p>
      {subtitle && <p className="text-sm mt-1">{subtitle}</p>}
    </div>
  );
}
