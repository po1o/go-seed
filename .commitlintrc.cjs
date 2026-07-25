// Commit-message rules (Conventional Commits). This is a .cjs file, not YAML,
// because `ignores` must be a function — which YAML cannot express.
module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'body-max-line-length': [2, 'always', 200],
    'type-enum': [
      2,
      'always',
      ['chore', 'ci', 'docs', 'feat', 'fix', 'perf', 'refactor', 'revert', 'style', 'test'],
    ],
  },
  // Skip Dependabot's own bump commits: their changelog bodies routinely exceed
  // any line-length cap and we don't author them. The auto-tidy commit this repo
  // pushes is a normal Conventional Commit and is still linted.
  ignores: [(message) => /Signed-off-by: dependabot\[bot\]/.test(message)],
};
