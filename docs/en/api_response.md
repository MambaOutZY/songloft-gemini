# API Response Format Specification

This project adopts a **RESTful direct-return style** and does not use a unified `{code, data, message}` envelope.

---

## Success Responses

### Single Entity — Return the Model Directly

```json
// GET /api/v1/songs/1
{"id":1, "title":"Sample Track", "artist":"Sample Artist", ...}
```

### Paginated List — Collection Name + Pagination Metadata

```json
// GET /api/v1/songs?limit=20&offset=0
{"songs":[...], "total":100, "limit":20, "offset":0}
```

### Operation Result — `{"message": "..."}`

```json
// DELETE /api/v1/songs/1
{"message": "Song deleted"}
```

You may use either `models.SuccessResponse` or `map[string]string{"message": "..."}`.

---

## Error Responses

All errors are returned uniformly via `respondError`, with a fixed format:

```json
{"error": "Human-readable error message", "detail": "Optional technical details"}
```

- `error`: Always present; a short, user-facing description
- `detail`: Optional; only emitted when a non-nil `err` is passed in, containing the underlying error message

Corresponding struct: `models.ErrorResponse`.

The **middleware layer** must likewise return errors in JSON format (using the local `respondAuthError`); returning plain text via `http.Error()` is **prohibited**.

### Exception: Binary Stream Endpoints

For binary stream endpoints such as playback (`/songs/{id}/play`), proxying (`/proxy`), and static files, errors may use `http.Error()` / `http.NotFound()`, since the client does not expect a JSON body.

---

## Prohibitions

| Prohibited | Reason |
|------|------|
| `{code, data, message}` envelope | `code` semantically duplicates the HTTP status code, adding parsing burden on the client |
| Custom error field names | You must use `error` + `detail`; substitutes such as `message`, `msg`, or `reason` are not allowed |
| Returning plain-text errors from API endpoints | The frontend's `response.json()` parsing will throw (except for binary stream endpoints) |

---

## Implementation Notes

| Scenario | Usage |
|------|----------|
| Return data | `respondJSON(w, status, data)` — `data` is serialized directly as top-level JSON |
| Return an error | `respondError(w, status, message, err)` — automatically builds `{error, detail}` |
| Middleware error | `respondAuthError(w, status, message, err)` — same format as `respondError` |
| jsplugin error | `writePluginUnavailable` — also uses the `{error, detail}` fields |
