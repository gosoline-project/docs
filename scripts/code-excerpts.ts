// Writes the example code that pages include into docs/_code.
//
// Blume's <include> embeds a whole file from inside the content root. The
// examples are runnable Go modules, some outside docs/, and pages show marked
// regions of them, so this copies each example file and each
// `// snippet-start: <name>` ... `// snippet-end: <name>` region (`#` in YAML)
// into docs/_code before every build:
//
//   docs/how-to/x/src/main.go          -> docs/_code/how-to/x/src/main.go
//   region "input output" of that file -> docs/_code/how-to/x/src/main.input-output.go
//   examples/fixtures/simple/main.go   -> docs/_code/examples/fixtures/simple/main.go
//
// Snippet markers are removed, and `highlight-next-line`, `highlight-start`
// and `highlight-end` comments become Shiki's `[!code highlight]` notation.
// A region also loses its outer blank lines and common indentation, and its
// leading tabs become two spaces.
import { mkdirSync, readdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { dirname, extname, join, relative } from "node:path";

const out = "docs/_code";
const sources = [
  { root: "docs", prefix: "" },
  { root: "examples", prefix: "examples" },
];
const snippet = /(?:\/\/|#) snippet-(start|end): (.+?)\s*$/;
const highlight = /^\s*(?:\/\/|#|--) highlight-(next-line|start|end)\s*$/;

function* files(dir: string): Generator<string> {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) {
      if (path !== out) yield* files(path);
    } else if (dir.split("/").includes("src") || root(path) === "examples") {
      yield path;
    }
  }
}

function root(path: string) {
  return path.split("/")[0];
}

// Shiki reads the notation only inside a comment of the block's language.
function comment(file: string) {
  const ext = extname(file);
  if ([".yml", ".yaml", ".sh"].includes(ext)) return "#";
  return ext === ".sql" ? "--" : "//";
}

// Drops marker lines and tags each highlighted line with a trailing comment.
function render(lines: string[], tag: string): string[] {
  const result: string[] = [];
  let next = false;
  let open = false;
  for (const line of lines) {
    const [, kind] = line.match(highlight) ?? [];
    if (kind === "next-line") next = true;
    else if (kind === "start") open = true;
    else if (kind === "end") open = false;
    else if (!snippet.test(line)) {
      // A tag on a blank line would count as indentation when a region is dedented.
      result.push((next || open) && line.trim() ? `${line} ${tag} [!code highlight]` : line);
      next = false;
    }
  }
  return result;
}

function region(lines: string[], tag: string): string {
  const kept = render(lines, tag);
  while (kept.length && !kept[0].trim()) kept.shift();
  while (kept.length && !kept.at(-1)!.trim()) kept.pop();
  const spaced = kept.map((line) => line.replace(/^\t+/, (tabs) => "  ".repeat(tabs.length)));
  const indents = spaced.filter((line) => line.trim()).map((line) => line.length - line.trimStart().length);
  const indent = indents.length ? Math.min(...indents) : 0;
  return `${spaced.map((line) => line.slice(indent)).join("\n")}\n`;
}

function write(path: string, text: string) {
  mkdirSync(dirname(path), { recursive: true });
  writeFileSync(path, text);
}

rmSync(out, { recursive: true, force: true });
let count = 0;
for (const { root: dir, prefix } of sources) {
  for (const file of files(dir)) {
    const target = join(out, prefix, relative(dir, file));
    const lines = readFileSync(file, "utf8").split("\n");
    write(target, render(lines, comment(file)).join("\n"));
    count++;

    const open = new Map<string, number>();
    lines.forEach((line, index) => {
      const [, kind, name] = line.match(snippet) ?? [];
      if (kind === "start") open.set(name, index);
      if (kind !== "end") return;
      const start = open.get(name);
      if (start === undefined) throw new Error(`${file}:${index + 1}: snippet-end without start: ${name}`);
      open.delete(name);
      const ext = extname(file);
      const slug = name.replaceAll(/\s+/g, "-");
      write(`${ext ? target.slice(0, -ext.length) : target}.${slug}${ext}`, region(lines.slice(start + 1, index), comment(file)));
      count++;
    });
    for (const name of open.keys()) throw new Error(`${file}: snippet-start without end: ${name}`);
  }
}
console.log(`code-excerpts: wrote ${count} files to ${out}`);
