import {
  forwardRef,
  type HTMLAttributes,
  type ReactNode,
  type ComponentPropsWithoutRef,
  type ElementRef,
} from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { cn } from '@/lib/cn';

export const Modal = Dialog.Root;
export const ModalTrigger = Dialog.Trigger;
export const ModalPortal = Dialog.Portal;
export const ModalClose = Dialog.Close;

export const ModalOverlay = forwardRef<
  ElementRef<typeof Dialog.Overlay>,
  ComponentPropsWithoutRef<typeof Dialog.Overlay>
>(function ModalOverlay({ className, ...props }, ref) {
  return (
    <Dialog.Overlay
      ref={ref}
      className={cn(
        'fixed inset-0 z-50 bg-black/60 backdrop-blur-sm',
        'data-[state=open]:animate-in data-[state=closed]:animate-out',
        className,
      )}
      {...props}
    />
  );
});

export interface ModalContentProps extends ComponentPropsWithoutRef<typeof Dialog.Content> {
  children: ReactNode;
}

export const ModalContent = forwardRef<ElementRef<typeof Dialog.Content>, ModalContentProps>(
  function ModalContent({ className, children, ...props }, ref) {
    return (
      <ModalPortal>
        <ModalOverlay />
        <Dialog.Content
          ref={ref}
          className={cn(
            'fixed left-1/2 top-1/2 z-50 w-full max-w-lg -translate-x-1/2 -translate-y-1/2',
            'rounded-lg border border-border bg-surface p-6 shadow-xl',
            className,
          )}
          {...props}
        >
          {children}
        </Dialog.Content>
      </ModalPortal>
    );
  },
);

export function ModalHeader({ className, ...props }: HTMLAttributes<HTMLDivElement>): ReactNode {
  return <div className={cn('mb-4 flex flex-col gap-1', className)} {...props} />;
}

export const ModalTitle = forwardRef<
  ElementRef<typeof Dialog.Title>,
  ComponentPropsWithoutRef<typeof Dialog.Title>
>(function ModalTitle({ className, ...props }, ref) {
  return <Dialog.Title ref={ref} className={cn('text-lg font-semibold', className)} {...props} />;
});

export const ModalDescription = forwardRef<
  ElementRef<typeof Dialog.Description>,
  ComponentPropsWithoutRef<typeof Dialog.Description>
>(function ModalDescription({ className, ...props }, ref) {
  return (
    <Dialog.Description ref={ref} className={cn('text-sm text-muted', className)} {...props} />
  );
});
