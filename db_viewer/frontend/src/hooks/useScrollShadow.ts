import { useEffect, useRef, useState, useCallback } from 'react';

export function useScrollShadow() {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [shadowRight, setShadowRight] = useState(false);
  const [shadowBottom, setShadowBottom] = useState(false);

  const update = useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    setShadowRight(el.scrollLeft + el.clientWidth < el.scrollWidth - 1);
    setShadowBottom(el.scrollTop + el.clientHeight < el.scrollHeight - 1);
  }, []);

  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    el.addEventListener('scroll', update);
    const ro = new ResizeObserver(update);
    ro.observe(el);
    // Also observe content children so shadow updates when table data changes
    Array.from(el.children).forEach(child => ro.observe(child));
    const mo = new MutationObserver(() => {
      // Re-observe new children and recalculate
      ro.disconnect();
      ro.observe(el);
      Array.from(el.children).forEach(child => ro.observe(child));
      update();
    });
    mo.observe(el, { childList: true });
    update();
    return () => {
      el.removeEventListener('scroll', update);
      ro.disconnect();
      mo.disconnect();
    };
  }, [update]);

  const wrapperClass = `data-view-wrapper${shadowRight ? ' shadow-right' : ''}${shadowBottom ? ' shadow-bottom' : ''}`;

  return { scrollRef, wrapperClass, updateShadow: update };
}
