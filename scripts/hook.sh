#!/usr/bin/env bash
# scripts/hook.sh — the PreToolUse forwarder.
#
# It carries no policy. The event arrives on stdin, `loopctl hook check` rules
# on it, and whatever loopctl says goes back out on stdout. The rules live in
# loopctl because the two lists they are made of — .loop/policy/blocklist.yaml
# and .loop/policy/test-deps.yaml — belong to the trusted base, and because a
# ruling that is wrong here does not fail, it silently lets a write through
# (which is why it has to be pinned by a test matrix, not reviewed by eye).
#
# Fail-closed: if loopctl cannot be found, cannot run, or returns nothing, the
# write is refused. The one thing decided before loopctl is consulted is the
# absence of LOOP_STAGE — that is not the loop running, and a person editing a
# file in their own terminal must not be blocked because the trusted base has
# not been built yet. It is the same reading the write policy gives the
# blocklist: 人类改护栏不在 LOOP_STAGE 下发生.
#
# Installed by `loopctl init` and hash-checked by `loopctl config validate`:
# drift here is silent, which is the test for what gets pinned.

set -o nounset

# No stage, no ruling. Exit 0 with no output is the protocol's "carry on".
[ -n "${LOOP_STAGE:-}" ] || exit 0

deny() {
	# The reason reaches the calling agent verbatim, so it says what to do
	# instead rather than only what happened.
	printf '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"%s"}}\n' "$1"
	exit 0
}

payload=$(cat)

ctl=${LOOPCTL:-}
if [ -z "$ctl" ]; then
	if command -v loopctl >/dev/null 2>&1; then
		ctl=$(command -v loopctl)
	elif [ -x "${CLAUDE_PROJECT_DIR:-.}/tools/loopctl/loopctl" ]; then
		ctl="${CLAUDE_PROJECT_DIR:-.}/tools/loopctl/loopctl"
	else
		deny "loopctl is not on PATH, so the write policy cannot be read; install loopctl or set LOOPCTL. Refusing rather than guessing."
	fi
fi

out=$(printf '%s' "$payload" | "$ctl" hook check 2>/dev/null)
status=$?
if [ "$status" -ne 0 ] || [ -z "$out" ]; then
	deny "loopctl hook check produced no ruling, so the write policy could not be applied. Refusing rather than guessing."
fi

printf '%s\n' "$out"
exit 0
