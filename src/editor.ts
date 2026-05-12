export function resolveEditor(env: Record<string, string | undefined>): string {
  const visual = (env.VISUAL ?? "").trim();
  if (visual) return visual;
  const editor = (env.EDITOR ?? "").trim();
  if (editor) return editor;
  throw new Error("Set $EDITOR or $VISUAL to use ee");
}

export type TuiHandle = {
  exit: () => void;
  enter: () => void;
};

export async function runEditor(
  editorCmd: string,
  filePath: string,
  tui: TuiHandle,
): Promise<void> {
  const parts = editorCmd.split(/\s+/).filter(Boolean);
  if (parts.length === 0) throw new Error("empty editor command");
  const [bin, ...args] = parts;

  tui.exit();
  try {
    const cmd = new Deno.Command(bin, {
      args: [...args, filePath],
      stdin: "inherit",
      stdout: "inherit",
      stderr: "inherit",
    });
    const child = cmd.spawn();
    await child.status;
  } finally {
    tui.enter();
  }
}
