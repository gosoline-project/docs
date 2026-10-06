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

Pages show example code with Blume includes of `docs/_code/`, which `scripts/code-excerpts.ts` regenerates before `dev`, `build`, and `validate`:

```mdx
<include lang="go" meta='title="main.go" lineNumbers'>/_code/how-to/x/src/main.go</include>
<include lang="go" meta='title="main.go"'>/_code/how-to/x/src/main.input-output.go</include>
```

The first is the whole `docs/how-to/x/src/main.go`; the second is its `// snippet-start: input output` region. Files under `examples/` appear under `/_code/examples/`. `highlight-next-line`, `highlight-start`, and `highlight-end` comments in a source highlight its lines. Edit the source file, never `docs/_code/`. Restart `bun run dev` after editing an example source.

## Deployment

GitHub Actions validates and builds pushes to `main`, then replaces the published site on `gh-pages` while preserving existing PR preview directories. Pull requests from this repository build previews under `/docs/pr-<number>/`; closing a PR removes its preview. The workflows use the pinned tools from `mise.toml` and the frozen Bun lockfile.

Search uses Blume's built-in local index in production and previews. Markdown page mirrors, `llms.txt`, and `llms-full.txt` are generated with the site. The static GitHub Pages deployment supports these files; request-time services require a server host.
