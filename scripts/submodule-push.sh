#!/bin/bash
#
# Push all submodules that have unpushed commits to their remote.
# Before pushing, lists uncommitted changes and committer info for review.

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"

submodule_paths=$(git submodule status --recursive | awk '{print $2}')

if [ -z "$submodule_paths" ]; then
    echo "No submodules found."
    exit 0
fi

# Phase 1: collect submodules that have something to report
needs_push=()
has_dirty=()

for submodule_path in $submodule_paths; do
    full_path="$REPO_ROOT/$submodule_path"
    [ -d "$full_path" ] || continue

    dirty=$(git -C "$full_path" status --short 2>/dev/null)
    unpushed=$(git -C "$full_path" log --oneline '@{upstream}..HEAD' 2>/dev/null || true)

    if [ -n "$dirty" ] || [ -n "$unpushed" ]; then
        if [ -n "$dirty" ]; then
            has_dirty+=("$submodule_path")
        fi
        if [ -n "$unpushed" ]; then
            needs_push+=("$submodule_path")
        fi
    fi
done

if [ ${#has_dirty[@]} -eq 0 ] && [ ${#needs_push[@]} -eq 0 ]; then
    echo "All submodules are clean and up-to-date with remote."
    exit 0
fi

# Phase 2: show status
for submodule_path in $submodule_paths; do
    full_path="$REPO_ROOT/$submodule_path"
    [ -d "$full_path" ] || continue

    dirty=$(git -C "$full_path" status --short 2>/dev/null)
    unpushed=$(git -C "$full_path" log --oneline '@{upstream}..HEAD' 2>/dev/null || true)

    [ -n "$dirty" ] || [ -n "$unpushed" ] || continue

    echo "━━━ $submodule_path ━━━"

    # committer info
    name=$(git -C "$full_path" config user.name 2>/dev/null || echo "(unset)")
    email=$(git -C "$full_path" config user.email 2>/dev/null || echo "(unset)")
    echo "  committer: $name <$email>"

    if [ -n "$dirty" ]; then
        echo "  uncommitted changes:"
        echo "$dirty" | sed 's/^/    /'
    fi

    if [ -n "$unpushed" ]; then
        echo "  unpushed commits:"
        echo "$unpushed" | sed 's/^/    /'
    fi
    echo
done

if [ ${#has_dirty[@]} -gt 0 ]; then
    echo "⚠  ${#has_dirty[@]} submodule(s) have uncommitted changes (listed above)."
    echo "   Please commit or stash them before pushing."
    echo
fi

if [ ${#needs_push[@]} -eq 0 ]; then
    echo "No unpushed commits to push."
    exit 0
fi

# Phase 3: confirm and push
echo "${#needs_push[@]} submodule(s) ready to push: ${needs_push[*]}"
read -r -p "Push now? [y/N] " answer
if [[ ! "$answer" =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 0
fi

failed=()
for submodule_path in "${needs_push[@]}"; do
    full_path="$REPO_ROOT/$submodule_path"
    echo "==> Pushing $submodule_path..."
    if git -C "$full_path" push; then
        echo "    ✓ pushed"
    else
        echo "    ✗ failed"
        failed+=("$submodule_path")
    fi
done

echo
if [ ${#failed[@]} -eq 0 ]; then
    echo "All submodules pushed successfully."
else
    echo "Failed to push:"
    for m in "${failed[@]}"; do
        echo "  - $m"
    done
    exit 1
fi
