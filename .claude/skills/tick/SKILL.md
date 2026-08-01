---
name: tick
description: Run exactly one Loop Engine v2 tick through the trusted TypeScript runner. Use when the user invokes /tick, including under /loop.
---

# Run one tick

Execute exactly:

```bash
"${LOOP_INSTALL_PREFIX:-$HOME/.local}/bin/loop-runner" tick
```

Present the runner's final JSON result to the user.

Do not interpret or reproduce the state machine. Do not call roles, `loopctl`,
or platform tools yourself. Do not retry, repair, or continue a failed tick;
report the runner's error and stop.
