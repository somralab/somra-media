import { type ReactNode, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { PosterCard, type PosterCardItem } from './PosterCard';

export interface MediaRowProps {
  titleKey: string;
  items: PosterCardItem[];
  titleNs?: string;
}

export function MediaRow({ titleKey, items, titleNs = 'discover' }: MediaRowProps): ReactNode {
  const { t } = useTranslation(titleNs);
  const scrollRef = useRef<HTMLDivElement>(null);

  if (items.length === 0) {
    return null;
  }

  return (
    <section aria-labelledby={`row-${titleKey}`} className="flex flex-col gap-3">
      <h2 id={`row-${titleKey}`} className="text-lg font-semibold text-text">
        {t(titleKey)}
      </h2>
      <div
        ref={scrollRef}
        className="flex gap-4 overflow-x-auto scroll-smooth pb-2 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
        role="list"
      >
        {items.map((item) => (
          <div key={item.id} role="listitem" className="w-36 sm:w-40">
            <PosterCard item={item} />
          </div>
        ))}
      </div>
    </section>
  );
}
