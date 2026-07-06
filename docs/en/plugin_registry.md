# Plugin Registry Authoring Guide

This document explains how to create and publish a Songloft plugin registry, so that other users can subscribe to your registry URL and browse and install your plugins from the in-app "Plugin Store".

---

## What Is a Plugin Registry

A plugin registry is a JSON file containing a list of `plugin.json` URLs. Once a user adds your JSON URL under "Settings → JS Plugin Management → Plugin Store → Manage Subscriptions", they can see all the plugins in your registry and install them with one click.

The backend automatically fetches each plugin's name, version, description, and other metadata from its `plugin.json` URL, so there's no need to duplicate that information in the registry file.

Plugin registries support **nested references** (`includes`), letting you compose multiple standalone registries for decentralized plugin distribution.

---

## JSON Format Specification

```json
{
  "name": "My Plugin Registry",
  "includes": [
    "https://example.com/other-registry.json"
  ],
  "plugins": [
    "https://raw.githubusercontent.com/you/example-plugin/main/plugin.json",
    "https://raw.githubusercontent.com/you/another-plugin/main/plugin.json"
  ]
}
```

### Field Reference

| Field | Type | Required | Description |
|------|------|------|------|
| `name` | string | No | Registry name, shown in the UI |
| `includes` | string[] | No | Array of nested registry URLs to reference |
| `plugins` | string[] | Yes | Array of URLs to each plugin's `plugin.json` |

### Automatic Resolution

Each URL in `plugins` points to a `plugin.json` in a plugin repository. The backend automatically fetches and parses the following fields:

- `name` — plugin name
- `entryPath` — unique plugin identifier
- `version` — version number
- `description` — description
- `author` — author
- `homepage` — project homepage
- `download_url` — ZIP package download URL
- `updateUrl` — update check URL
- `minHostVersion` — minimum host version

If `plugin.json` does not contain a `download_url` (which is the case for most plugins), the backend automatically retrieves it from the `manifest.json` referenced by `updateUrl`.

---

## Complete Examples

### Minimal Example

```json
{
  "plugins": [
    "https://raw.githubusercontent.com/you/my-plugin/main/plugin.json"
  ]
}
```

### Full Example with Nesting

```json
{
  "name": "Songloft Community Aggregate Registry",
  "includes": [
    "https://raw.githubusercontent.com/alice/songloft-plugins/main/registry.json",
    "https://raw.githubusercontent.com/bob/my-plugin-registry/main/registry.json"
  ],
  "plugins": [
    "https://raw.githubusercontent.com/songloft-org/songloft-plugin-miot/main/plugin.json",
    "https://raw.githubusercontent.com/songloft-org/songloft-plugin-subsonic/main/plugin.json"
  ]
}
```

---

## Publishing to GitHub

### Option 1: In-Repo Raw URL (Recommended)

1. Create `registry.json` in the root of your GitHub repository
2. Push to the `main` branch
3. The registry URL is:
   ```
   https://raw.githubusercontent.com/{username}/{repo}/main/registry.json
   ```

**Hosting the plugin ZIP**: GitHub Releases are recommended:
1. Create a Release in the plugin repository
2. Upload the `.jsplugin.zip` as a Release Asset
3. In `plugin.json`, point `updateUrl` at `manifest.json`, and set `download_url` in `manifest.json`:
   ```
   https://github.com/{username}/{repo}/releases/download/v{version}/{entry_path}.jsplugin.zip
   ```

### Option 2: GitHub Pages

1. Enable GitHub Pages for the repository
2. Place `registry.json` at the Pages root
3. The registry URL is:
   ```
   https://{username}.github.io/{repo}/registry.json
   ```

### Example Repository Structure

```
my-plugin-registry/
├── registry.json          ← Registry JSON (just lists plugin.json URLs)
└── README.md
```

The plugins themselves are maintained in their own repositories, for example:

```
my-songloft-plugin/
├── plugin.json            ← Plugin metadata (name, version, entryPath, etc.)
├── manifest.json          ← Update manifest (version + download_url)
├── main.js                ← Plugin entry point
└── ...
```

---

## Publishing to a JS CDN

### jsDelivr (via npm)

1. Create an npm package containing `registry.json`
2. Publish to npm:
   ```bash
   npm publish
   ```
3. The registry URL is:
   ```
   https://cdn.jsdelivr.net/npm/{package}@latest/registry.json
   ```

### unpkg (via npm)

Same as jsDelivr, just swap the domain:
```
https://unpkg.com/{package}@latest/registry.json
```

### jsDelivr (via GitHub)

You can also use jsDelivr to accelerate GitHub files directly without publishing an npm package:
```
https://cdn.jsdelivr.net/gh/{username}/{repo}@{branch}/registry.json
```

---

## Use Cases for Nested Registries

The `includes` field lets one registry reference others, recursively downloading and merging all their plugins.

### Typical Patterns

**Aggregator registry**: collect several independent authors' registries so users can subscribe to them all at once:
```json
{
  "name": "Community Aggregate",
  "includes": [
    "https://raw.githubusercontent.com/alice/plugins/main/registry.json",
    "https://raw.githubusercontent.com/bob/plugins/main/registry.json",
    "https://raw.githubusercontent.com/charlie/plugins/main/registry.json"
  ],
  "plugins": []
}
```

**Official + community**: an official registry with core plugins that also pulls in community contributions:
```json
{
  "name": "Official Registry",
  "includes": [
    "https://community.example.com/registry.json"
  ],
  "plugins": [
    "https://raw.githubusercontent.com/songloft-org/songloft-plugin-miot/main/plugin.json"
  ]
}
```

### Deduplication Rules

When the same `entryPath` appears in multiple registries (including nested ones), the entry with the **higher version number** is kept. Versions are compared segment by segment numerically after splitting on `.`.

---

## Private Registry Authentication

Plugin registries support **Bearer Token authentication**, used to distribute closed-source plugins or to access plugins in private GitHub repositories.

### How to Configure

When adding or editing a registry in the "Manage Subscriptions" dialog, simply fill in the Token field. Registries with a configured Token show a lock icon in the list.

The token is sent as an `Authorization: Bearer <token>` header. Scoping rules:

- **The registry JSON itself**, **all URLs in the plugins array**, and **ZIP downloads** — always carry the token
- **Sub-registries referenced by includes** — carry the token only when they share the **same host** as the registry URL; cross-host includes do not carry the token (to prevent leaks)

As a result, a private registry can safely include public official or third-party registries without leaking the token to them. Meanwhile, private sub-registries on the same server (same-host includes) can still authenticate normally.

### Private GitHub Repositories

1. Create a [Fine-grained Personal Access Token](https://github.com/settings/tokens?type=beta) on GitHub
2. Permissions: only **Contents: Read-only** on the target repository is needed
3. Use the `raw.githubusercontent.com` format for the registry URL, same as public repositories:
   ```
   https://raw.githubusercontent.com/{username}/{private-repo}/main/registry.json
   ```
4. Fill the PAT into the Token field under "Manage Subscriptions"

> A GitHub PAT works for both `raw.githubusercontent.com` (registry/plugin.json) and `github.com` (Release downloads). The Go HTTP client automatically strips the Authorization header on cross-host redirects, and the S3 signed URL that Release downloads redirect to does not require a token, so the behavior is correct.

### Self-Hosted Private Registries

Simply validate the `Authorization: Bearer <token>` header on your server. Using nginx as an example:

```nginx
location /registry/ {
    if ($http_authorization != "Bearer your-secret-token") {
        return 401;
    }
    root /var/www/plugins;
}
```

### Security Considerations

- The token is stored in plaintext in the server's config table (this is a self-hosted app, within the user's control)
- Sub-registries referenced by `includes` carry the token only on the **same host**; cross-host includes never leak the token
- The token is sent via the `Authorization` request header and never appears in the URL, so it won't leak into logs or the Referrer

---

## Notes and Limits

| Limit | Value | Description |
|------|-----|------|
| Maximum total plugins | 500 | Excess entries are truncated and logged as a warning |
| Maximum recursion depth | 20 levels | `includes` beyond 20 levels are skipped |
| Maximum size per JSON | 2 MB | Larger files are rejected |
| Request timeout per URL | 15 seconds | On timeout, the URL is skipped (logged as a warning) |
| Circular references | Auto-detected | A → B → A won't loop forever; already-visited URLs are automatically skipped |

- URLs in `plugins` must point to valid `plugin.json` files
- If the registry file is hosted on GitHub, users in mainland China can select a GitHub mirror accelerator (preset or custom) in the Plugin Store, or configure a general proxy (e.g. `http://192.168.1.1:7890`) under "Settings → System → HTTP Proxy" to speed up access
- A failed fetch of one `includes` or `plugins` entry does not affect the loading of other registries and plugins
