# Sandwitches Mobile App

This is the native Flutter application for the Sandwitches platform. It connects to the Sandwitches Django API to provide a seamless mobile experience for browsing and viewing recipes.

## Features

- **Recipe Discovery**: Browse a list of all available recipes.
- **Detailed Views**: See ingredients, instructions, and tags for each recipe.
- **State Management**: Efficient data handling using the `Provider` pattern.
- **Image Caching**: Optimized image loading with `cached_network_image`.

## Prerequisites

- [Flutter SDK](https://docs.flutter.dev/get-started/install) (version 3.0.0 or higher)
- [Dart SDK](https://dart.dev/get-dart)
- An Android emulator, iOS simulator, or a physical device.
- The Sandwitches Django backend running and accessible.

## Getting Started

### 1. Install Dependencies

Navigate to the `mobile` directory and run:

```bash
cd mobile
flutter pub get
```

### 2. Configure API Base URL

By default, the app is configured to look for the backend at `http://localhost:8000`.

If you are running on a physical device or a different host, update the `baseUrl` in `lib/services/api_service.dart`:

```dart
// lib/services/api_service.dart
ApiService({this.baseUrl = 'http://YOUR_LOCAL_IP:8000/api/'});
```

*Note: For Android Emulators, use `http://10.0.2.2:8000/api/` to access the host's localhost.*

### 3. Run the App

Ensure your device is connected or an emulator is running, then execute:

```bash
flutter run
```

## Project Structure

- `lib/models/`: Data models reflecting the Django Ninja schemas (`Recipe`, `Tag`, `User`).
- `lib/services/`: API communication logic (`ApiService`).
- `lib/providers/`: Application state management (`RecipeProvider`).
- `lib/screens/`: Main UI screens (`RecipeListScreen`, `RecipeDetailScreen`).
- `lib/widgets/`: Reusable UI components (`RecipeCard`).
- `lib/main.dart`: App entry point and theme configuration.

## Development

### Adding New API Endpoints

1. Create a model in `lib/models/` if necessary.
2. Add a new method to `ApiService` in `lib/services/api_service.dart`.
3. Update `RecipeProvider` (or create a new provider) in `lib/providers/` to manage the new data.
4. Consume the data in your screens/widgets.

## Future Roadmap

- **Authentication**: Implement login and user profile management.
- **Ordering**: Add the ability to order sandwiches directly from the app.
- **Favorites**: Sync favorite recipes with the user's account.
- **Offline Support**: Cache recipe data for offline browsing.
