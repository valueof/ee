import { assertEquals, assertRejects } from "@std/assert";
import { findRepoRoot, listChanges } from "../src/git.ts";

async function run(cwd: string, ...args: string[]) {
  const cmd = new Deno.Command("git", { args, cwd, stdout: "piped", stderr: "piped" });
  const { code, stderr } = await cmd.output();
  if (code !== 0) {
    throw new Error(`git ${args.join(" ")} failed: ${new TextDecoder().decode(stderr)}`);
  }
}

async function setupRepo(): Promise<string> {
  const dir = await Deno.makeTempDir({ prefix: "git-ed-test-" });
  await run(dir, "init", "-b", "main");
  await run(dir, "config", "user.email", "test@example.com");
  await run(dir, "config", "user.name", "Test");
  return dir;
}

Deno.test("findRepoRoot: returns toplevel from subdirectory", async () => {
  const dir = await setupRepo();
  try {
    await Deno.mkdir(`${dir}/sub`);
    const root = await findRepoRoot(`${dir}/sub`);
    assertEquals(await Deno.realPath(root), await Deno.realPath(dir));
  } finally {
    await Deno.remove(dir, { recursive: true });
  }
});

Deno.test("findRepoRoot: rejects when not in a repo", async () => {
  const dir = await Deno.makeTempDir({ prefix: "git-ed-norepo-" });
  try {
    await assertRejects(() => findRepoRoot(dir));
  } finally {
    await Deno.remove(dir, { recursive: true });
  }
});

Deno.test("listChanges: returns staged + unstaged, skips untracked", async () => {
  const dir = await setupRepo();
  try {
    await Deno.writeTextFile(`${dir}/committed.txt`, "v1\n");
    await run(dir, "add", "committed.txt");
    await run(dir, "commit", "-m", "init");

    await Deno.writeTextFile(`${dir}/committed.txt`, "v2\n");
    await Deno.writeTextFile(`${dir}/staged-new.txt`, "hi\n");
    await run(dir, "add", "staged-new.txt");
    await Deno.writeTextFile(`${dir}/untracked.txt`, "u\n");

    const entries = await listChanges(dir);
    const paths = entries.map((e) => e.path).sort();
    assertEquals(paths, ["committed.txt", "staged-new.txt"]);
  } finally {
    await Deno.remove(dir, { recursive: true });
  }
});
