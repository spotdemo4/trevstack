export function normalizePath(path: string) {
  return path.toLowerCase().replace(/\/+$/, "") || "/";
}

export function isPublicPath(path: string) {
  const normalized = normalizePath(path);
  return normalized === "/auth" || normalized === "/403";
}

export function getSignInPath(returnTo: string) {
  return `/auth?returnTo=${encodeURIComponent(returnTo)}`;
}

export function getReturnPath(search: string, origin: string) {
  const returnTo = new URLSearchParams(search).get("returnTo");
  if (!returnTo || !returnTo.startsWith("/")) {
    return "/";
  }

  try {
    const target = new URL(returnTo, origin);
    if (target.origin !== origin || normalizePath(target.pathname) === "/auth") {
      return "/";
    }
    return `${target.pathname}${target.search}${target.hash}`;
  } catch {
    return "/";
  }
}
