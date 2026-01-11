# GitHub Pages Setup Guide

This guide explains how to set up and deploy Tsunami documentation to GitHub Pages using Volcano.

## Prerequisites

- GitHub repository with documentation in `docs/` folder
- GitHub Actions enabled for your repository
- Go installed locally (for building docs locally)

## Step 1: Enable GitHub Pages

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Pages**
3. Under "Build and deployment":
   - **Source**: Select "GitHub Actions"
4. Save your settings

That's it! GitHub Actions will now handle the deployment automatically.

## Step 2: Configure Repository Permissions

The GitHub Actions workflow needs appropriate permissions:

1. Go to **Settings** → **Actions** → **General**
2. Scroll to "Workflow permissions"
3. Select **"Read and write permissions"**
4. Check **"Allow GitHub Actions to create and approve pull requests"**
5. Save

## Step 3: Deploy Documentation

### Automatic Deployment

The documentation will automatically deploy when:
- You push changes to the `main` branch
- Changes are made to files in the `docs/` folder
- Changes are made to `.github/workflows/deploy-docs.yml`

### Manual Deployment

You can also trigger deployment manually:

1. Go to **Actions** tab
2. Select "Deploy Documentation" workflow
3. Click **"Run workflow"**
4. Select branch and click **"Run workflow"**

## Step 4: Access Your Documentation

After the first successful deployment:

1. Go to **Settings** → **Pages**
2. You'll see: "Your site is live at `https://<username>.github.io/<repository>/`"
3. Click the URL to view your documentation

**Example URL**: `https://wusher.github.io/tsunami/`

## Custom Domain (Optional)

To use a custom domain:

1. Purchase a domain from a registrar
2. Add a CNAME file to `docs/` folder:
   ```bash
   echo "docs.tsunami.dev" > docs/CNAME
   ```
3. Configure DNS at your registrar:
   ```
   Type: CNAME
   Name: docs (or @)
   Value: <username>.github.io
   ```
4. In GitHub Settings → Pages:
   - Enter your custom domain: `docs.tsunami.dev`
   - Check "Enforce HTTPS" (after DNS propagates)

## Local Development

### Install Volcano

```bash
go install github.com/volcano/volcano/cmd/volcano@latest
```

Add to your PATH:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Build Documentation Locally

```bash
./scripts/build-docs.sh
```

Output will be in `public/` folder.

### Serve Documentation Locally

```bash
./scripts/serve-docs.sh
```

This starts a local server at `http://localhost:8000`

You can specify a custom port:
```bash
./scripts/serve-docs.sh 3000
```

### Hot Reload During Development

While serving, Volcano watches for file changes and auto-reloads:

```bash
cd docs
volcano serve --watch
```

## Troubleshooting

### Build Fails in GitHub Actions

**Issue**: Workflow fails with "volcano: command not found"

**Solution**: The workflow installs Volcano automatically. Check that:
1. The `go-version` in workflow matches your go.mod version
2. The Volcano package path is correct

### Pages Not Updating

**Issue**: Changes pushed but site doesn't update

**Solution**:
1. Check **Actions** tab for workflow status
2. Ensure workflow completed successfully
3. Wait 2-3 minutes for CDN cache to clear
4. Hard refresh browser (Ctrl+Shift+R)

### 404 Error on GitHub Pages

**Issue**: Site shows 404 error

**Solution**:
1. Verify GitHub Pages is enabled (Settings → Pages)
2. Check that source is set to "GitHub Actions"
3. Ensure workflow ran successfully
4. Wait a few minutes for initial deployment

### Broken Links in Documentation

**Issue**: Internal links don't work

**Solution**:
- Use relative paths: `./guides/basic-usage.md`
- Not absolute: `/guides/basic-usage.md`
- Volcano handles path resolution automatically

### Build Works Locally But Fails in CI

**Issue**: `volcano build` works locally but fails in GitHub Actions

**Solution**:
1. Check go.mod version matches workflow
2. Ensure all documentation files are committed
3. Check for file path case sensitivity (CI uses Linux)
4. Verify volcano.yml configuration

## Configuration

### volcano.yml

Key configuration options in `docs/volcano.yml`:

```yaml
# Site URL (update with your GitHub Pages URL)
site:
  url: "https://wusher.github.io/tsunami"

# Output directory
build:
  output: "../public"

# Theme
theme: "docs"  # or "blog", "minimal", "custom"
```

### Workflow Customization

Edit `.github/workflows/deploy-docs.yml` to:

**Deploy only on docs changes:**
```yaml
on:
  push:
    branches: [main]
    paths:
      - 'docs/**'
```

**Deploy on tags:**
```yaml
on:
  push:
    tags:
      - 'v*'
```

**Use custom build command:**
```yaml
- name: Build documentation
  run: |
    volcano build --theme custom --output ./public
```

## Maintenance

### Updating Documentation

1. Edit markdown files in `docs/`
2. Test locally: `./scripts/serve-docs.sh`
3. Commit and push to main branch
4. GitHub Actions deploys automatically

### Adding New Pages

1. Create new `.md` file in appropriate folder
2. Add to navigation in `docs/volcano.yml`:
   ```yaml
   nav:
     - New Page: path/to/new-page.md
   ```
3. Link from other pages: `[Link Text](./path/to/new-page.md)`

### Updating Theme

Update `theme` in `docs/volcano.yml`:

```yaml
theme: "custom"  # Options: docs, blog, minimal, custom
```

For custom themes, create `docs/theme/` folder with templates.

## Badges

Add documentation badge to README:

```markdown
[![Documentation](https://img.shields.io/badge/docs-latest-blue.svg)](https://wusher.github.io/tsunami)
```

Add build status badge:

```markdown
[![Docs](https://github.com/wusher/tsunami/actions/workflows/deploy-docs.yml/badge.svg)](https://github.com/wusher/tsunami/actions/workflows/deploy-docs.yml)
```

## Cost

GitHub Pages is **free** for public repositories with these limits:
- Source repository < 1 GB
- Published site < 1 GB
- 100 GB bandwidth/month
- 10 builds/hour

These limits are more than sufficient for documentation sites.

## Security

### HTTPS

GitHub Pages provides free HTTPS via Let's Encrypt:
- Automatic for `*.github.io` domains
- Custom domains: Enable "Enforce HTTPS" in Settings

### Permissions

The workflow uses minimal permissions:
```yaml
permissions:
  contents: read      # Read repo contents
  pages: write        # Write to GitHub Pages
  id-token: write     # OIDC token for deployment
```

## Next Steps

- ✅ Documentation is now live
- 📝 Update documentation content
- 🎨 Customize theme and styling
- 🔍 Add search functionality (built into Volcano)
- 📊 Set up analytics (optional)
- 🌐 Configure custom domain (optional)

## Support

- **Volcano Documentation**: [volcano.dev/docs](https://volcano.dev/docs)
- **GitHub Pages**: [docs.github.com/pages](https://docs.github.com/en/pages)
- **Issues**: Open an issue in the repository

---

**Last updated**: 2026-01-11
