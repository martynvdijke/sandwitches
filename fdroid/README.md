# F-Droid submission

This directory stages the metadata that belongs in the
[`fdroiddata`](https://gitlab.com/fdroid/fdroiddata) repository. The app is
built and signed by F-Droid; the APK attached to a GitHub release is not
uploaded anywhere.

The metadata targets `v3.14.1`, which contains the F-Droid readiness fixes,
native OIDC login and the longer OIDC login window. Do not submit before that
tag exists and the release APK is published.

## Before submitting

The release must contain:

- `INTERNET` permission in the release manifest
- a version code derived from the semantic version (`3.14.1` -> `3014001`)
- a Flutter version pinned by `.fvmrc`
- Fastlane metadata under `fastlane/metadata/android/en-US/`

## Submit

1. Fork and clone `fdroiddata`:

   ```sh
   git clone git@gitlab.com:<your-user>/fdroiddata.git
   cd fdroiddata
   git switch -c xyz.vandijke.sandwitches master
   ```

2. Copy the metadata file:

   ```sh
   cp /path/to/sandwitches/fdroid/xyz.vandijke.sandwitches.yml metadata/
   ```

3. Validate locally:

   ```sh
   pipx install fdroidserver
   fdroid readmeta
   fdroid rewritemeta xyz.vandijke.sandwitches
   fdroid lint xyz.vandijke.sandwitches
   ```

   `fdroid build xyz.vandijke.sandwitches` needs the F-Droid buildserver, so
   run it on your fork's CI pipeline instead of a normal workstation.

4. Commit and push:

   ```sh
   git add metadata/xyz.vandijke.sandwitches.yml
   git commit -m "New App: xyz.vandijke.sandwitches"
   git push origin xyz.vandijke.sandwitches
   ```

5. Open a merge request against `fdroid/fdroiddata` and add the `New App`
   label. Mention that the app is a client for a self-hostable MIT server in
   the same repository, so it does not depend on a proprietary service.

## After the first build

`fdroid checkupdates` uses `UpdateCheckData` to read the version code from
`mobile/pubspec.yaml`, so `AutoUpdateMode: Version` can add new build blocks
for future `vX.Y.Z` tags. The first build block is committed manually.

## Signature

F-Droid signs its own APK. Users cannot switch between the F-Droid build and
the GitHub release APK without uninstalling. If that becomes a problem,
reproducible builds (`Binaries` plus `AllowedAPKSigningKeys`) are the way to
ship the developer-signed APK through F-Droid, but Flutter builds are not
trivially reproducible.
