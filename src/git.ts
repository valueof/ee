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
