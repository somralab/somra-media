import { type ReactNode, useRef } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { cn } from '@/lib/cn';
import { PosterCard, type PosterCardItem } from './PosterCard';

export type ViewMode = 'grid' | 'list';

export interface MediaGridProps {
  items: PosterCardItem[];
  viewMode?: ViewMode;
  className?: string;
}

export function MediaGrid({ items, viewMode = 'grid', className }: MediaGridProps): ReactNode {
  const parentRef = useRef<HTMLDivElement>(null);

  const columnCount = viewMode === 'grid' ? 4 : 1;
  const rowCount = Math.ceil(items.length / columnCount);

  const rowVirtualizer = useVirtualizer({
    count: rowCount,
    getScrollElement: () => parentRef.current,
    estimateSize: () => (viewMode === 'grid' ? 280 : 96),
    overscan: 3,
  });

  if (items.length === 0) {
    return null;
  }

  return (
    <div ref={parentRef} className={cn('h-[70vh] overflow-auto', className)}>
      <div
        style={{ height: `${rowVirtualizer.getTotalSize()}px`, position: 'relative' }}
        role="list"
      >
        {rowVirtualizer.getVirtualItems().map((virtualRow) => {
          const startIndex = virtualRow.index * columnCount;
          const rowItems = items.slice(startIndex, startIndex + columnCount);
          return (
            <div
              key={virtualRow.key}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                transform: `translateY(${virtualRow.start}px)`,
              }}
              className={cn(
                'grid gap-4 px-1',
                viewMode === 'grid' ? 'grid-cols-2 sm:grid-cols-3 lg:grid-cols-4' : 'grid-cols-1',
              )}
            >
              {rowItems.map((item) =>
                viewMode === 'grid' ? (
                  <div key={item.id} role="listitem">
                    <PosterCard item={item} />
                  </div>
                ) : (
                  <div
                    key={item.id}
                    role="listitem"
                    className="flex items-center gap-4 rounded-md border border-border p-3"
                  >
                    <div className="w-16 shrink-0">
                      <PosterCard item={item} className="w-full" />
                    </div>
                    <div className="min-w-0">
                      <p className="font-medium">{item.title}</p>
                      {item.year ? <p className="text-sm text-muted">{item.year}</p> : null}
                    </div>
                  </div>
                ),
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
