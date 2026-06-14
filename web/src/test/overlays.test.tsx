import { describe, expect, it, vi } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import {
  Modal,
  ModalTrigger,
  ModalContent,
  ModalHeader,
  ModalTitle,
  ModalDescription,
} from '@/components/ui/Modal';
import { ToastProvider, useToast } from '@/components/ui/Toast';

describe('<Modal />', () => {
  it('renders trigger and shows content when open', () => {
    render(
      <Modal open>
        <ModalTrigger>open</ModalTrigger>
        <ModalContent>
          <ModalHeader>
            <ModalTitle>Title</ModalTitle>
            <ModalDescription>Description</ModalDescription>
          </ModalHeader>
        </ModalContent>
      </Modal>,
    );
    expect(screen.getByText('Title')).toBeInTheDocument();
    expect(screen.getByText('Description')).toBeInTheDocument();
  });
});

describe('<ToastProvider />', () => {
  it('publishes a toast when the hook is called', () => {
    const publishSpy = vi.fn();

    function PublishHarness(): null {
      const { publish } = useToast();
      publishSpy(publish);
      return null;
    }

    render(
      <ToastProvider>
        <PublishHarness />
      </ToastProvider>,
    );

    expect(publishSpy).toHaveBeenCalled();
    const firstCall = publishSpy.mock.calls[0] ?? [];
    const publish = firstCall[0] as ReturnType<typeof useToast>['publish'];
    act(() => {
      publish({ title: 'Hello', variant: 'success' });
    });
    expect(screen.getByText('Hello')).toBeInTheDocument();
  });

  it('throws when useToast is used outside provider', () => {
    function Naked(): null {
      useToast();
      return null;
    }
    expect(() => render(<Naked />)).toThrow();
  });
});
