# Browser console and Cloudflare

## Fixed in this repository

- The Permissions Policy now sends a conservative set of broadly supported directives instead of experimental/removed directives that Chrome warns about.
- CSP allows the Cloudflare Web Analytics beacon script and its metrics connection without enabling `unsafe-inline`.
- i18next declares English as the only bundled language, so a German browser no longer requests missing `/locales/de-DE/translation.json` and `/locales/de/translation.json` files. Debug logging is development-only.

## Required Cloudflare dashboard rule

Disable Rocket Loader for the Portainer hostname (for example `portainer.themodcraft.net`). Portainer uses a strict CSP and a complex SPA bundle; Rocket Loader injects/transforms scripts and its inline bootstrap is correctly blocked by that CSP. Do not add `unsafe-inline` to make Rocket Loader work.

Keep Web Analytics enabled if desired. The application CSP permits `https://static.cloudflareinsights.com` for scripts and `https://cloudflareinsights.com` for connections.

`redirectionChainSiteScript.js`, `contentScript.js`, and `content-safety.js` are browser-extension scripts, not Portainer assets. Confirm by testing a private browser window with extensions disabled. Their attempts to redefine `window.location` or use `unload` cannot be repaired in this repository.

References: [Cloudflare Web Analytics CSP](https://developers.cloudflare.com/web-analytics/faq/), [Cloudflare Rocket Loader troubleshooting](https://developers.cloudflare.com/speed/optimization/content/rocket-loader/), and [MDN Permissions Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/Permissions_Policy).
