/** @type {import('semantic-release').GlobalConfig} */
module.exports = {
  branches: [
    "main",
    {
      name: "release/+([0-9])?(.{+([0-9]),x}).x",
      range: "${name.replace(/^release\\//g, '')}",
    },
    "next",
    { name: "alpha", prerelease: true },
    { name: "beta", prerelease: true },
    { name: "rc", prerelease: true },
  ],
  plugins: [
    [
      "@semantic-release/commit-analyzer",
      {
        preset: "conventionalcommits",
      },
    ],
    [
      "@semantic-release/release-notes-generator",
      {
        preset: "conventionalcommits",
        presetConfig: {
          preset: {
            name: "conventionalchangelog",
          },
        },
        writerOpts: {
          // Collect unique contributors from all commits in the release
          // and expose them to the Handlebars template as `context.contributors`.
          finalizeContext(context) {
            const seen = new Map();
            for (const group of context.commitGroups ?? []) {
              for (const commit of group.commits ?? []) {
                const name = commit.author?.name;
                const login = commit.author?.login; // present when using GitHub API
                if (name && !seen.has(name)) {
                  seen.set(name, { name, login: login ?? null });
                }
              }
            }
            context.contributors = [...seen.values()];
            return context;
          },

          // Append a Contributors section at the end of the release notes.
          // If a contributor has a GitHub login, link their profile; otherwise
          // fall back to displaying only their name.
          footerPartial:
            "{{#if noteGroups}}" +
            "{{#each noteGroups}}" +
            "\n\n### {{title}}\n\n" +
            "{{#each notes}}" +
            "* {{#if commit.scope}}**{{commit.scope}}:** {{/if}}{{text}}\n" +
            "{{/each}}" +
            "{{/each}}" +
            "{{/if}}" +
            "{{#if contributors}}" +
            "\n\n### Contributors\n\n" +
            "{{#each contributors}}" +
            "{{#if login}}" +
            "* [@{{login}}](https://github.com/{{login}})\n" +
            "{{else}}" +
            "* {{name}}\n" +
            "{{/if}}" +
            "{{/each}}" +
            "{{/if}}\n",
        },
      },
    ],
    "@semantic-release/github",
  ],
};
