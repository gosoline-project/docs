import { defineConfig } from "blume";

export default defineConfig({
  title: "Gosoline Docs",
  description: "Documentation and articles for the Gosoline Go application framework.",
  logo: { image: "/img/logo-transparent.png" },
  theme: {
    accent: { light: "#2e8555", dark: "#25c2a0" },
  },
  navigation: {
    tabs: [
      { label: "Docs", path: "/" },
      { label: "Blog", path: "/blog" },
    ],
    sidebar: { display: "group" },
  },
  footer: {
    links: [
      { label: "Overview", href: "/" },
      { label: "Getting started", href: "/getting-started" },
      { label: "How-to guides", href: "/how-to" },
      { label: "Reference", href: "/reference" },
      { label: "Blog", href: "/blog" },
    ],
  },
  github: { owner: "gosoline-project", repo: "docs", branch: "main" },
  markdown: {
    code: { theme: { light: "github-light", dark: "dracula" } },
  },
  deployment: {
    site: "https://gosoline-project.github.io",
    base: process.env.DOCS_BASE_URL || "/docs/",
  },
  redirects: [
    { from: "/category/fundamentals", to: "/fundamentals" },
    { from: "/category/getting-started", to: "/getting-started" },
    { from: "/category/testing", to: "/getting-started/testing" },
    { from: "/category/how-to-guides", to: "/how-to" },
    { from: "/category/http-server", to: "/how-to/http-server" },
    { from: "/category/streaming-applications", to: "/how-to/streaming-applications" },
    { from: "/category/reference", to: "/reference" },
    { from: "/category/migrations", to: "/migrations" },
  ],
});
