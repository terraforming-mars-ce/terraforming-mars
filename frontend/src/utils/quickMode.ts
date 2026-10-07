import { useLayoutEffect } from "react";
import { useLocation } from "react-router-dom";

/**
 * Testing switch: load any page with `?quick` to skip slow-loading extras (currently
 * the menu's 3D space background). The flag is read once at boot and kept in the URL
 * across in-app navigation by `useKeepQuickParam`.
 */
export const QUICK_PARAM = "quick";

const initialValue = new URLSearchParams(window.location.search).get(QUICK_PARAM);

export const QUICK_MODE = initialValue !== null;

/**
 * Re-appends `?quick` whenever an in-app navigation drops it. Rewrites the address bar
 * directly instead of navigating, so the router sees no new location and route effects
 * (game bootstrapping keyed on `location.state`) don't run a second time.
 */
export function useKeepQuickParam() {
  const location = useLocation();

  useLayoutEffect(() => {
    if (!QUICK_MODE) {
      return;
    }
    const params = new URLSearchParams(location.search);
    if (params.has(QUICK_PARAM)) {
      return;
    }
    // URLSearchParams would write a bare flag as "quick=", so append it by hand.
    const flag = initialValue ? `${QUICK_PARAM}=${encodeURIComponent(initialValue)}` : QUICK_PARAM;
    const rest = params.toString();
    const search = rest ? `${rest}&${flag}` : flag;
    window.history.replaceState(
      window.history.state,
      "",
      `${location.pathname}?${search}${location.hash}`,
    );
  }, [location]);
}
