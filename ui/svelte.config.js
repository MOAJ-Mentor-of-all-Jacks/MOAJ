// MOAJ (Mentor of all Jacks) - SvelteKit Static Adapter Configuration Scaffold
// Implementation Target: Phase 4 (First usable workflow slice) & Phase 8 (Embedded single binary)
// Reference: docs/adr/ADR-01-go-sqlc-single-binary.md (NFR-01)

/** @type {import('@sveltejs/kit').Config} */
const config = {
  kit: {
    // Configured with @sveltejs/adapter-static in Phase 4 for Go binary embedding (embed.FS)
  }
};

export default config;
