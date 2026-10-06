# Gosoline Docs

Documentation and articles for the Gosoline Go application framework, built with [Blume](https://useblume.dev/).

## Development

Install the project tools and dependencies:

```sh
mise install
mise exec -- bun install --frozen-lockfile
mise exec -- bun run dev
```

Node and Bun versions are pinned in `mise.toml`. Blume serves the site under `/docs/`, matching GitHub Pages. Use `DOCS_BASE_URL=/` for a preview mounted at the root.

## Verification

```sh
mise exec -- bun run typecheck
mise exec -- bun run validate
mise exec -- bun run build
mise exec -- bun run preview
```

`validate` checks links, anchors, assets, and includes in strict mode. `build` checks configuration, frontmatter, and routes, and writes the static site to `dist/`.

## Content

Pages live in `docs/`; blog posts live in `docs/blog/`. Frontmatter supplies each page's title and sidebar order. Folder `meta.ts` files supply category labels and generated index cards. Navigation follows the filesystem.

Reuse whole code files with Blume includes:

```mdx
<include lang="go" meta='title="main.go" lineNumbers'>./src/example/main.go</include>
```

Blume does not select named snippet regions. Those examples are inlined as fenced code; when changing a source example, update its corresponding named snippets in the documentation too. Migration instruction files remain in `docs/migrations/src/`, with downloadable copies in `public/downloads/`; keep those copies in sync.

## Deployment

GitHub Actions validates and builds pushes to `main`, then replaces the published site on `gh-pages` while preserving existing PR preview directories. This removes obsolete Docusaurus output on deployment. Pull requests from this repository build previews under `/docs/pr-<number>/`; closing a PR removes its preview. The workflows use the pinned tools from `mise.toml` and the frozen Bun lockfile.

Search uses Blume's built-in local index in production and previews. Markdown page mirrors, `llms.txt`, and `llms-full.txt` are generated with the site. The static GitHub Pages deployment supports these files; request-time services require a server host.

## Migration from Docusaurus

The migration preserves the 45 documentation pages, five blog articles, standalone Markdown sample page, blog slugs, assets, repository links, code themes, and light/dark accent colors. Eight former `/category/…` URLs redirect to folder indexes. Blog posts retain their dates, authors, and tags as search metadata, with a new card index.

Docusaurus, React/MUI wrappers, raw-loader, Algolia configuration, and the separate llms plugin were replaced by Blume. Overview cards use native components; migration instructions use code-block copy controls. Named snippets and examples with Docusaurus highlight markers were converted to fences. Two obsolete log-context snippet names now show the corresponding `CreateTodo` and `UpdateTodo` methods. Unused starter components and illustrations were removed.

Blume supplies its own layout, generated social cards, and URLs without trailing slashes in links. The old global social-card image remains available as an asset. Docusaurus blog author/tag/archive pages, automatic reading-time labels, footer column headings/copyright, and custom Infima/MUI/blog-width CSS were dropped. Article content and existing page URLs remain available, including URLs requested with trailing slashes on GitHub Pages.
