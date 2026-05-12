import * as tui from "./tui.ts";
import * as git from "./git.ts";
import * as app from "./app.ts";

const VERSION = "0.1.0";

function help(): void {
  console.log(`ee ${VERSION}

Usage: ee

A TUI that lists tracked git changes (staged + unstaged) and opens the
selected file in your editor.

Environment:
  VISUAL, EDITOR   Editor command (VISUAL takes precedence)
  NO_COLOR         Set to disable colored output

Keys:
  ↑/↓ or j/k       Move cursor (wraps)
  1-9              Open file at that position
  Enter            Open file at cursor
  q, Esc, Ctrl-C   Quit`);
}

function installSignalHandlers(): void {
  const cleanup = () => {
    tui.exit();
    Deno.exit(130);
  };
  Deno.addSignalListener("SIGINT", cleanup);
  Deno.addSignalListener("SIGTERM", cleanup);
}

async function main(): Promise<number> {
  const args = Deno.args;
  if (args.includes("--version")) {
    console.log(VERSION);
    return 0;
  }
  if (args.includes("--help") || args.includes("-h")) {
    help();
    return 0;
  }

  let repoRoot: string;
  try {
    repoRoot = await git.findRepoRoot();
  } catch {
    console.error("ee: not a git repository");
    return 1;
  }

  const initial = await git.listChanges(repoRoot);
  if (initial.length === 0) {
    console.log("Nothing to edit");
    return 0;
  }

  installSignalHandlers();

  try {
    tui.enter();
    await app.run(repoRoot, initial);
  } finally {
    tui.exit();
  }
  return 0;
}

if (import.meta.main) {
  try {
    const code = await main();
    Deno.exit(code);
  } catch (err) {
    tui.exit();
    console.error("ee:", err instanceof Error ? err.message : String(err));
    Deno.exit(1);
  }
}
