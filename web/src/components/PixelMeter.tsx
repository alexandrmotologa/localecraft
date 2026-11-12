import React from 'react';

interface PixelMeterProps {
  renderedWidth: number;
  availableWidth: number;
  maxWidth: number;
  isOverflow: boolean;
  overflowPX?: number;
  overflowPCT?: number;
}

export const PixelMeter: React.FC<PixelMeterProps> = ({
  renderedWidth,
  availableWidth,
  maxWidth,
  isOverflow,
  overflowPX = 0,
  overflowPCT = 0,
}) => {
  const percentage = availableWidth > 0 ? (renderedWidth / availableWidth) * 100 : 0;
  const barWidth = Math.min(percentage, 100);

  let barColor = 'bg-emerald-500';
  let badgeColor = 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20';

  if (isOverflow) {
    barColor = 'bg-rose-500 animate-pulse';
    badgeColor = 'text-rose-400 bg-rose-500/10 border-rose-500/30';
  } else if (percentage > 85) {
    barColor = 'bg-amber-500';
    badgeColor = 'text-amber-400 bg-amber-500/10 border-amber-500/20';
  }

  return (
    <div className="space-y-1.5 w-full">
      <div className="flex items-center justify-between text-xs">
        <span className="text-slate-400 font-mono">
          <strong className="text-slate-200">{renderedWidth.toFixed(1)}px</strong> / {maxWidth.toFixed(0)}px
        </span>
        <span className={`text-[11px] font-mono px-1.5 py-0.5 rounded border ${badgeColor}`}>
          {isOverflow ? `+${overflowPX.toFixed(1)}px (+${overflowPCT.toFixed(0)}%)` : `${percentage.toFixed(0)}% budget`}
        </span>
      </div>

      <div className="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden flex">
        <div
          className={`h-full transition-all duration-300 rounded-full ${barColor}`}
          style={{ width: `${barWidth}%` }}
        />
      </div>
    </div>
  );
};
