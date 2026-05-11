export type Entry = {
  indexStatus: string;
  worktreeStatus: string;
  path: string;
  renamedFrom?: string;
};

export function parsePorcelain(bytes: Uint8Array): Entry[] {
  const text = new TextDecoder().decode(bytes);
  if (text.length === 0) return [];

  const records = text.split("\0");
  if (records.at(-1) === "") records.pop();

  const entries: Entry[] = [];
  for (let i = 0; i < records.length; i++) {
    const rec = records[i];
    if (rec.length < 3) continue;
    const indexStatus = rec[0];
    const worktreeStatus = rec[1];
    const path = rec.slice(3);
    entries.push({ indexStatus, worktreeStatus, path });
  }
  return entries;
}
