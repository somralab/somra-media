import { type ReactNode } from 'react';
import { cn } from '@/lib/cn';

export interface SkeletonProps {
  className?: string;
  lines?: number;
}

export function Skeleton({ className, lines = 1 }: SkeletonProps): ReactNode {
  return (
    <div className={cn('animate-pulse', className)} aria-hidden="true">
      {Array.from({ length: lines }).map((_, i) => (
        <div key={i} className="mb-2 h-4 rounded bg-border/60 last:mb-0" />
      ))}
    </div>
  );
}

export function PosterSkeleton(): ReactNode {
  return (
    <div className="flex flex-col gap-2" aria-hidden="true">
      <div className="aspect-[2/3] animate-pulse rounded-md bg-border/60" />
      <div className="h-4 w-3/4 animate-pulse rounded bg-border/60" />
    </div>
  );
}

export function MediaRowSkeleton(): ReactNode {
  return (
    <div className="flex flex-col gap-3" aria-busy="true">
      <div className="h-6 w-48 animate-pulse rounded bg-border/60" />
      <div className="flex gap-4 overflow-hidden">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="w-36 shrink-0">
            <PosterSkeleton />
          </div>
        ))}
      </div>
    </div>
  );
}
