/**
 * Progressive-enhancement light-dismiss fallback for `<dialog closedby="any">`
 * on WebKit/Safari where `HTMLDialogElement.prototype.closedBy` is not yet supported.
 * Tests pointer coordinates (not event.target) on both pointerdown and pointerup
 * so dragging a text selection onto the backdrop never dismisses the dialog.
 */
export function attachLightDismissFallback(dialog: HTMLDialogElement): () => void {
  if ('closedBy' in HTMLDialogElement.prototype) {
    return () => {};
  }

  const isOutside = (e: PointerEvent) => {
    const r = dialog.getBoundingClientRect();
    return (
      e.clientX < r.left ||
      e.clientX > r.right ||
      e.clientY < r.top ||
      e.clientY > r.bottom
    );
  };

  let pressedOutside = false;
  const onDown = (e: PointerEvent) => {
    pressedOutside = isOutside(e);
  };
  const onUp = (e: PointerEvent) => {
    if (pressedOutside && isOutside(e)) {
      dialog.close();
    }
    pressedOutside = false;
  };

  dialog.addEventListener('pointerdown', onDown);
  dialog.addEventListener('pointerup', onUp);

  return () => {
    dialog.removeEventListener('pointerdown', onDown);
    dialog.removeEventListener('pointerup', onUp);
  };
}

export function columnGlyph(colId: string): string {
  switch (colId.toLowerCase()) {
    case 'backlog':
      return '❀';
    case 'ready':
      return '◈';
    case 'active':
      return '❧';
    case 'done':
      return '❖';
    default:
      return '✦';
  }
}
