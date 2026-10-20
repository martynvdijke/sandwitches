# sandwitches

Sandwitches native Android client

## Release signing

Release builds read signing material from `mobile/android/key.properties`:

```properties
storeFile=upload-keystore.jks
storePassword=...
keyAlias=upload
keyPassword=...
```

The keystore and `key.properties` are gitignored. CI recreates both from the
`ANDROID_KEYSTORE_BASE64`, `ANDROID_KEYSTORE_PASSWORD`, `ANDROID_KEY_ALIAS` and
`ANDROID_KEY_PASSWORD` repository secrets before semantic-release builds the
APK. Keep an offline backup of the keystore: losing it means existing installs
can no longer be updated.

When no keystore is present, release builds fall back to the debug key so local
`flutter run --release` keeps working. F-Droid removes the signing config
entirely and signs the APK with its own key.

## F-Droid

See [`../fdroid/README.md`](../fdroid/README.md) for the submission recipe.
Store listing text and graphics live in `../fastlane/metadata/android/en-US/`.

## Getting Started

This project is a starting point for a Flutter application.

A few resources to get you started if this is your first Flutter project:

- [Learn Flutter](https://docs.flutter.dev/get-started/learn-flutter)
- [Write your first Flutter app](https://docs.flutter.dev/get-started/codelab)
- [Flutter learning resources](https://docs.flutter.dev/reference/learning-resources)

For help getting started with Flutter development, view the
[online documentation](https://docs.flutter.dev/), which offers tutorials,
samples, guidance on mobile development, and a full API reference.
