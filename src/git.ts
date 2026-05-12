export type Entry = {
  indexStatus: string;
  worktreeStatus: string;
  path: string;
  renamedFrom?: string;
};

const EDITABLE_CODES = new Set(["M", "A", "D", "R", "C", "T", "U", "?"]);

export function filterEditable(entries: Entry[]): Entry[] {
  return entries.filter(
    (e) =>
      EDITABLE_CODES.has(e.indexStatus) ||
      EDITABLE_CODES.has(e.worktreeStatus),
  );
}

export async function findRepoRoot(cwd: string = Deno.cwd()): Promise<string> {
  const cmd = new Deno.Command("git", {
    args: ["rev-parse", "--show-toplevel"],
    cwd,
    stdout: "piped",
    stderr: "piped",
  });
  const { code, stdout, stderr } = await cmd.output();
  if (code !== 0) {
    throw new Error(
      `not a git repository: ${new TextDecoder().decode(stderr).trim()}`,
    );
  }
  return new TextDecoder().decode(stdout).trimEnd();
}

export async function listChanges(repoRoot: string): Promise<Entry[]> {
  const cmd = new Deno.Command("git", {
    args: ["status", "--porcelain=v1", "-z"],
    cwd: repoRoot,
    stdout: "piped",
    stderr: "piped",
  });
  const { code, stdout, stderr } = await cmd.output();
  if (code !== 0) {
    throw new Error(
      `git status failed: ${new TextDecoder().decode(stderr).trim()}`,
    );
  }
  return filterEditable(parsePorcelain(stdout));
}

export function parsePorcelain(bytes: Uint8Array): Entry[] {
  const text = new TextDecoder().decode(bytes);
  if (text.length === 0) return [];

  const records = text.split("\0");
  if (records.at(-1) === "") records.pop();

  const entries: Entry[] = [];
  let i = 0;
  while (i < records.length) {
    const rec = records[i];
    if (rec.length < 3) {
      i++;
      continue;
    }
    const indexStatus = rec[0];
    const worktreeStatus = rec[1];
    const path = rec.slice(3);
    const isRenameOrCopy = indexStatus === "R" || indexStatus === "C" ||
      worktreeStatus === "R" || worktreeStatus === "C";
    if (isRenameOrCopy && i + 1 < records.length) {
      entries.push({
        indexStatus,
        worktreeStatus,
        path,
        renamedFrom: records[i + 1],
      });
      i += 2;
    } else {
      entries.push({ indexStatus, worktreeStatus, path });
      i += 1;
    }
  }
  return entries;
}
