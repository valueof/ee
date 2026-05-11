export function resolveEditor(env: Record<string, string | undefined>): string {
  const visual = (env.VISUAL ?? "").trim();
  if (visual) return visual;
  const editor = (env.EDITOR ?? "").trim();
  if (editor) return editor;
  throw new Error("Set $EDITOR or $VISUAL to use git-ed");
}
