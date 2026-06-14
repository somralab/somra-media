import {
  createContext,
  forwardRef,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ComponentPropsWithoutRef,
  type ElementRef,
  type ReactNode,
} from 'react';
import * as ToastPrimitive from '@radix-ui/react-toast';
import { cn } from '@/lib/cn';

export interface ToastItem {
  id: string;
  title: string;
  description?: string;
  variant?: 'default' | 'success' | 'danger';
}

interface ToastContextValue {
  publish: (toast: Omit<ToastItem, 'id'>) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext);
  if (!ctx) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return ctx;
}

const VARIANT_CLASSES: Record<NonNullable<ToastItem['variant']>, string> = {
  default: 'border-border bg-surface text-text',
  success: 'border-primary bg-surface text-text',
  danger: 'border-danger bg-surface text-text',
};

export const ToastRoot = forwardRef<
  ElementRef<typeof ToastPrimitive.Root>,
  ComponentPropsWithoutRef<typeof ToastPrimitive.Root> & { variant?: ToastItem['variant'] }
>(function ToastRoot({ className, variant = 'default', ...props }, ref) {
  return (
    <ToastPrimitive.Root
      ref={ref}
      className={cn(
        'pointer-events-auto flex w-full max-w-sm items-start gap-3 rounded-md border p-4 shadow-lg',
        VARIANT_CLASSES[variant ?? 'default'],
        className,
      )}
      {...props}
    />
  );
});

export interface ToastProviderProps {
  children: ReactNode;
}

export function ToastProvider({ children }: ToastProviderProps): ReactNode {
  const [items, setItems] = useState<ToastItem[]>([]);

  const publish = useCallback<ToastContextValue['publish']>((toast) => {
    const id =
      typeof crypto !== 'undefined' && 'randomUUID' in crypto
        ? crypto.randomUUID()
        : Math.random().toString(36).slice(2);
    setItems((prev) => [...prev, { id, ...toast }]);
  }, []);

  const handleOpenChange = useCallback((id: string, open: boolean) => {
    if (!open) {
      setItems((prev) => prev.filter((item) => item.id !== id));
    }
  }, []);

  const value = useMemo<ToastContextValue>(() => ({ publish }), [publish]);

  return (
    <ToastContext.Provider value={value}>
      <ToastPrimitive.Provider swipeDirection="right">
        {children}
        {items.map((item) => (
          <ToastRoot
            key={item.id}
            variant={item.variant ?? 'default'}
            onOpenChange={(open) => handleOpenChange(item.id, open)}
          >
            <div className="flex flex-col gap-1">
              <ToastPrimitive.Title className="text-sm font-semibold">
                {item.title}
              </ToastPrimitive.Title>
              {item.description ? (
                <ToastPrimitive.Description className="text-sm text-muted">
                  {item.description}
                </ToastPrimitive.Description>
              ) : null}
            </div>
          </ToastRoot>
        ))}
        <ToastPrimitive.Viewport className="fixed bottom-0 right-0 z-50 m-4 flex w-full max-w-sm flex-col gap-2 outline-none" />
      </ToastPrimitive.Provider>
    </ToastContext.Provider>
  );
}
