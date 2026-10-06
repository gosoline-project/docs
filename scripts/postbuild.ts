// Finishes dist/ after `blume build`.
//
// 1. Removes the `[!code highlight]` notation that scripts/code-excerpts.ts
//    writes for highlighted lines. Shiki strips it from the HTML, but the
//    Markdown mirrors, llms-full.txt and the JSON docs API copy code verbatim.
// 2. Writes blog/atom.xml, the Atom feed the Docusaurus site published, from
//    Blume's blog/rss.xml and each post's rendered article, keeping the old
//    entry ids so subscribers see no duplicate posts. blog/rss.xml gets the
//    old trailing-slash guids back for the same reason.
import { readdirSync, readFileSync, writeFileSync } from "node:fs";
import { extname, join } from "node:path";

const dist = "dist";

function* files(dir: string): Generator<string> {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) yield* files(path);
    else yield path;
  }
}

for (const file of files(dist)) {
  if (![".md", ".mdx", ".txt", ".json"].includes(extname(file))) continue;
  const text = readFileSync(file, "utf8");
  const clean = text.replaceAll(/ (?:\/\/|#|--) \[!code highlight\]/g, "");
  if (clean !== text) writeFileSync(file, clean);
}

const rss = readFileSync(join(dist, "blog/rss.xml"), "utf8").replaceAll(
  /(<guid[^>]*>)([^<]*?)\/?(<\/guid>)/g,
  "$1$2/$3",
);
writeFileSync(join(dist, "blog/rss.xml"), rss);
const tag = (xml: string, name: string) => xml.match(new RegExp(`<${name}>([\\s\\S]*?)</${name}>`))?.[1] ?? "";
const entities: Record<string, string> = { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" };
const escape = (text: string) => text.replaceAll(/[&<>"]/g, (char) => entities[char]);
const cdata = (text: string) => `<![CDATA[${text.replaceAll("]]>", "]]]]><![CDATA[>")}]]>`;
const home = new URL(tag(rss, "link"));
const base = home.pathname.replace(/\/$/, "");

const entries = [...rss.matchAll(/<item>([\s\S]*?)<\/item>/g)].map(([, item]) => {
  const link = tag(item, "link");
  const path = new URL(link).pathname.slice(base.length);
  const page = readFileSync(join(dist, path, "index.html"), "utf8");
  const start = page.indexOf("<article");
  const end = page.indexOf("</article>", start);
  if (start < 0 || end < 0) throw new Error(`postbuild: no <article> in ${path}/index.html`);
  const article = page.slice(page.indexOf(">", start) + 1, end);
  // Blume renders the title and description above the body; the feed carries them separately.
  const content = article.replace(/^\s*<h1>[\s\S]*?<\/h1>(\s*<p class="text-lg[^"]*">[\s\S]*?<\/p>)?/, "");
  const front = readFileSync(join(dist, `${path}.md`), "utf8").split("\n---\n")[0];
  const value = (text = "") => text.trim().replace(/^(["'])(.*)\1$/, "$2");
  const authors = [...front.matchAll(/- name: (.+)(?:\n\s+url: (.+))?/g)].map(([, name, uri]) => {
    const link = uri ? `<uri>${escape(value(uri))}</uri>` : "";
    return `<author><name>${escape(value(name))}</name>${link}</author>`;
  });
  const tagList = front.match(/tags: \[(.*)\]/)?.[1] ?? "";
  const tags = tagList.split(",").map(value).filter(Boolean);
  const updated = new Date(tag(item, "pubDate")).toISOString();
  return { updated, xml: `  <entry>
    <title type="html">${cdata(tag(item, "title"))}</title>
    <id>${link}/</id>
    <link href="${link}/"/>
    <updated>${updated}</updated>
    <summary type="html">${cdata(tag(item, "description"))}</summary>
    <content type="html" xml:base="${link}/">${cdata(content)}</content>
    ${authors.join("\n    ")}
    ${tags.map((t) => `<category label="${escape(t)}" term="${escape(t)}"/>`).join("\n    ")}
  </entry>` };
});

const blog = `${tag(rss, "link")}/blog/`;
writeFileSync(
  join(dist, "blog/atom.xml"),
  `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <id>${blog}</id>
  <title>${tag(rss, "title")}</title>
  <updated>${entries.map((e) => e.updated).sort().at(-1)}</updated>
  <link rel="alternate" href="${blog}"/>
  <link rel="self" href="${blog}atom.xml"/>
  <subtitle>${tag(rss, "description")}</subtitle>
${entries.map((e) => e.xml).join("\n")}
</feed>
`,
);
console.log(`postbuild: wrote ${join(dist, "blog/atom.xml")} with ${entries.length} entries`);
